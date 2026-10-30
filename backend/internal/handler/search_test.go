package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"vtb-copilot-backend/internal/domain"
	"vtb-copilot-backend/internal/handler"
)

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
