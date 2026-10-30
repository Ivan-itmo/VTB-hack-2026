package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"vtb-copilot-backend/internal/domain"
	"vtb-copilot-backend/internal/handler"
)

type generatorFunc func(context.Context, string, []domain.ChunkMatch) (domain.CopilotResponse, error)

func (f generatorFunc) Generate(ctx context.Context, text string, chunks []domain.ChunkMatch) (domain.CopilotResponse, error) {
	return f(ctx, text, chunks)
}

func TestCopilotGenerationAndFallback(t *testing.T) {
	for _, tc := range []struct {
		name      string
		failure   error
		unrelated bool
	}{
		{name: "success"},
		{name: "timeout", failure: errors.New("mock timeout")},
		{name: "invalid JSON", failure: errors.New("mock invalid JSON")},
		{name: "retrieval gate", unrelated: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			h, err := handler.NewWithGenerator("../../data/demo_cases.json", generatorFunc(func(ctx context.Context, text string, chunks []domain.ChunkMatch) (domain.CopilotResponse, error) {
				calls++
				if tc.failure != nil {
					return domain.CopilotResponse{}, tc.failure
				}
				c := chunks[0]
				return domain.CopilotResponse{SuggestedRoute: "L1", ConfidenceScore: 0.8, ResolutionDraft: "Проект ответа модели", Sources: []domain.SourceItem{{DocumentName: c.DocumentTitle, Clause: c.Citation, Quote: c.Text}}}, nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			text := "Неверный ИНН КПП получателя"
			if tc.unrelated {
				text = "Космический телескоп"
			}
			body, _ := json.Marshal(map[string]string{"task_text": text})
			rec := httptest.NewRecorder()
			h.ResolveCopilot(rec, httptest.NewRequest(http.MethodPost, "/api/v1/copilot/resolve", bytes.NewReader(body)))
			if rec.Code != 200 {
				t.Fatalf("status %d", rec.Code)
			}
			var resp domain.CopilotResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.EscalationRequired != (tc.failure != nil || tc.unrelated) {
				t.Fatalf("wrong escalation: %+v", resp)
			}
			if tc.unrelated && calls != 0 {
				t.Fatal("retrieval gate called model")
			}
			if !tc.unrelated && calls != 1 {
				t.Fatal("model not called exactly once")
			}
			if tc.failure != nil && resp.ConfidenceScore != 0 {
				t.Fatal("fallback claims confidence")
			}
			if tc.failure == nil && !tc.unrelated && (resp.ResolutionDraft != "Проект ответа модели" || len(resp.Sources) != 1) {
				t.Fatal("generated result was overwritten")
			}
		})
	}
}

func TestSearchChunksEndpoint(t *testing.T) {
	h, err := handler.New("../../data/demo_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		body       string
		status     int
		escalation bool
	}{
		{`{"text":"Неверный ИНН КПП получателя"}`, 200, false},
		{`{"text":"Космический телескоп"}`, 200, true},
		{`{"text":""}`, 200, true},
		{`{`, 400, false},
	} {
		rec := httptest.NewRecorder()
		h.SearchChunks(rec, httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBufferString(tc.body)))
		if rec.Code != tc.status {
			t.Fatalf("status %d for %s", rec.Code, tc.body)
		}
		if rec.Code != 200 {
			continue
		}
		var result domain.SearchResult
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.EscalationRequired != tc.escalation {
			t.Fatalf("wrong escalation for %s: %+v", tc.body, result)
		}
	}
}

func TestCopilotReturnsRetrievedQuotes(t *testing.T) {
	router := setupTestRouter(t)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/copilot/resolve",
		bytes.NewBufferString(`{"task_text":"Неверный ИНН КПП получателя, ошибка в БИК"}`)))
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	var result domain.CopilotResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.EscalationRequired || len(result.RetrievedChunks) != 3 || result.RetrievalScore <= 0 {
		t.Fatalf("unexpected retrieval: %+v", result)
	}
	if len(result.Sources) != len(result.RetrievedChunks) {
		t.Fatal("sources and chunks differ")
	}
	for i, chunk := range result.RetrievedChunks {
		if result.Sources[i].Quote != chunk.Text || result.Sources[i].Clause != chunk.Citation {
			t.Fatal("source is not the original retrieved quote")
		}
	}
}
