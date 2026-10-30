package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"vtb-copilot-backend/internal/domain"
)

var testChunks = []domain.ChunkMatch{{Chunk: domain.Chunk{
	ID: "REG-01-1.2", Document: "REG-01", DocumentTitle: "Регламент", Citation: "REG-01, п. 1.2", Text: "Проверить ИНН и КПП.",
}, Score: 0.7}}

func validResponse() string {
	return `{"task_type":"T1","regulation":"REG-01","priority":null,"autonomy_level":1,"route":{"executor":"L1","controller":null},"citations":[{"citation":"REG-01, п. 1.2","supports":"проверка реквизитов"}],"resolution":{"summary":"Проверить реквизиты","actions":["Проверить ИНН и КПП"],"deadlines":null,"client_draft":null},"missing_data":[],"confidence":"высокая","escalation_required":false,"escalation_reason":null,"escalation_brief":null}`
}

func envelope(text, finish string) string {
	data, _ := json.Marshal(map[string]any{"candidates": []any{map[string]any{"finishReason": finish, "content": content{Parts: []part{{Text: text}}}}}})
	return string(data)
}

func TestGenerateHTTPAndPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/models/"+DefaultModel+":generateContent" {
			t.Errorf("wrong request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("x-goog-api-key") != "test-secret" || r.URL.RawQuery != "" {
			t.Error("key must be in header only")
		}
		var req struct {
			System     content        `json:"systemInstruction"`
			Contents   []content      `json:"contents"`
			Generation map[string]any `json:"generationConfig"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		if !strings.Contains(req.System.Parts[0].Text, "Проверить ИНН и КПП.") || !strings.Contains(req.System.Parts[0].Text, "REG-01, п. 1.2") {
			t.Error("missing chunks")
		}
		if !strings.Contains(req.System.Parts[0].Text, "Неверный ИНН") {
			t.Error("missing ticket")
		}
		if req.Generation["responseMimeType"] != "application/json" || req.Generation["responseJsonSchema"] == nil {
			t.Error("missing JSON schema")
		}
		schema := req.Generation["responseJsonSchema"].(map[string]any)
		citations := schema["properties"].(map[string]any)["citations"].(map[string]any)
		fields := citations["items"].(map[string]any)["properties"].(map[string]any)
		if fields["citation"].(map[string]any)["enum"].([]any)[0] != testChunks[0].Citation {
			t.Error("citation is not constrained")
		}

		io.WriteString(w, envelope(validResponse(), "STOP"))
	}))
	defer server.Close()
	client, err := New(Config{APIKey: "test-secret", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Generate(context.Background(), "Неверный ИНН", testChunks)
	if err != nil {
		t.Fatal(err)
	}
	if resp.SuggestedRoute != "L1" || len(resp.Sources) != 1 || resp.EscalationRequired {
		t.Fatalf("wrong response: %+v", resp)
	}
}

func TestStrictResponseParsing(t *testing.T) {
	valid := validResponse()
	for _, data := range []string{
		"plain text", "```json\n" + valid + "\n```", valid + valid, "null", "[]", "{}",
		strings.Replace(valid, `"confidence":"высокая",`, "", 1),
		strings.Replace(valid, `"autonomy_level":1`, `"autonomy_level":4`, 1),
		strings.Replace(valid, `"escalation_required":false`, `"escalation_required":null`, 1),
		strings.Replace(valid, `"task_type":"T1"`, `"task_type":"T1","extra":true`, 1),
		strings.Replace(valid, `"task_type":"T1"`, `"task_type":"T1","task_type":"T2"`, 1),
		strings.Replace(valid, `"citation":"REG-01, п. 1.2"`, `"citation":"REG-99"`, 1),
		strings.Replace(valid, `"confidence":"высокая"`, `"confidence":"низкая"`, 1),
		strings.Replace(valid, `"missing_data":[]`, `"missing_data":["сумма"]`, 1),
		strings.Replace(valid, `"actions":["Проверить ИНН и КПП"]`, `"actions":[]`, 1),
	} {
		if _, err := parseResponse([]byte(data), testChunks); !errors.Is(err, ErrInvalidResponse) {
			t.Errorf("accepted invalid response: %v, %s", err, data)
		}
	}
}

func TestProviderFailureModes(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		want       error
	}{
		{"rate limit", "private provider text", 429, ErrProvider},
		{"unavailable", "private provider text", 503, ErrProvider},
		{"bad key", "private provider text", 400, ErrProvider},
		{"broken envelope", "not json", 200, ErrInvalidResponse},
		{"empty candidates", `{}`, 200, ErrInvalidResponse},
		{"blocked", `{"promptFeedback":{"blockReason":"SAFETY"}}`, 200, ErrInvalidResponse},
		{"truncated", envelope(validResponse(), "MAX_TOKENS"), 200, ErrInvalidResponse},
		{"malformed content", envelope("not json", "STOP"), 200, ErrInvalidResponse},
		{"oversized", strings.Repeat("x", maxResponseBytes+1), 200, ErrInvalidResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); io.WriteString(w, tc.body) }))
			defer srv.Close()
			client, _ := New(Config{APIKey: "test-secret", BaseURL: srv.URL})
			_, err := client.Generate(context.Background(), "ИНН", testChunks)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "test-secret") {
				t.Fatal("leaked private data")
			}
		})
	}
}

func TestTimeoutAndMissingKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(100 * time.Millisecond):
		}
	}))
	defer srv.Close()
	client, _ := New(Config{APIKey: "test-secret", BaseURL: srv.URL, Timeout: 30 * time.Millisecond})
	if _, err := client.Generate(context.Background(), "ИНН", testChunks); !errors.Is(err, ErrTimeout) {
		t.Fatalf("want timeout, got %v", err)
	}
	client, _ = New(Config{})
	if _, err := client.Generate(context.Background(), "ИНН", testChunks); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("want unavailable, got %v", err)
	}
}

func TestRedirectDoesNotLeakKey(t *testing.T) {
	called := false
	dest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer dest.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, dest.URL, http.StatusTemporaryRedirect)
	}))
	defer origin.Close()
	client, _ := New(Config{APIKey: "test-secret", BaseURL: origin.URL})
	if _, err := client.Generate(context.Background(), "ИНН", testChunks); !errors.Is(err, ErrProvider) {
		t.Fatal(err)
	}
	if called {
		t.Fatal("redirect followed")
	}
}
