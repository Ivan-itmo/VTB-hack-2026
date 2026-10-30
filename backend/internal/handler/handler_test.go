package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"vtb-copilot-backend/internal/domain"
	"vtb-copilot-backend/internal/handler"
	"vtb-copilot-backend/internal/middleware"
)

func setupTestRouter(t *testing.T) *chi.Mux {
	t.Helper()
	h, err := handler.New("../../data/demo_cases.json")
	if err != nil {
		t.Fatalf("Не удалось инициализировать handler: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.CORS)

	r.Get("/", h.Root)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/cases", h.GetDemoCases)
		r.Post("/copilot/resolve", h.ResolveCopilot)
		r.Post("/sandbox/evaluate", h.EvaluateSandbox)
	})

	return r
}

func TestHealthCheck(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен %d", rec.Code)
	}

	var resp domain.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	if resp.Status != "ok" || resp.Service != "vtb-copilot-go" {
		t.Errorf("Некорректный ответ healthcheck: %+v", resp)
	}
}

func TestGetDemoCases(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен %d", rec.Code)
	}

	var payload domain.DemoCasesPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Ошибка декодирования демо-кейсов: %v", err)
	}

	if len(payload.CopilotCases) == 0 {
		t.Errorf("CopilotCases пуст")
	}
	if len(payload.SandboxCases) == 0 {
		t.Errorf("SandboxCases пуст")
	}
}

func TestCORSPreflight(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/copilot/resolve", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Ожидался статус 200 для OPTIONS, получен %d", rec.Code)
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("Ожидался заголовок Access-Control-Allow-Origin: *, получено %q", origin)
	}
}

func TestCopilotResolveEscalationRetroDate(t *testing.T) {
	r := setupTestRouter(t)

	reqBody := []byte(`{"task_description": "Прошу провести платежное поручение изменением даты задним числом"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/resolve", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен %d", rec.Code)
	}

	var resp domain.CopilotResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Ошибка декодирования: %v", err)
	}

	if !resp.EscalationRequired {
		t.Errorf("Ожидался escalation_required = true")
	}
	if resp.ConfidenceScore != 0.20 {
		t.Errorf("Ожидался confidence_score = 0.20, получено %f", resp.ConfidenceScore)
	}
	if len(resp.Sources) == 0 {
		t.Errorf("Ожидался непустой список sources")
	}
}

func TestCopilotResolveEscalatesUnrelatedTextWithCaseID(t *testing.T) {
	r := setupTestRouter(t)

	reqBody := []byte(`{"case_id": "COPILOT-01", "task_text": "Космический телескоп наблюдает галактики"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/resolve", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен %d", rec.Code)
	}

	var resp domain.CopilotResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Ошибка декодирования: %v", err)
	}

	if !resp.EscalationRequired {
		t.Errorf("Идентификатор кейса не должен обходить проверку релевантности текста")
	}
	if resp.ConfidenceScore != 0 {
		t.Errorf("Для нерелевантного текста не должно быть уверенного ответа")
	}
	if len(resp.Sources) != 0 {
		t.Errorf("Для текста без совпадений не должно быть источников")
	}
}

func TestSandboxEvaluateWrongAnswer(t *testing.T) {
	r := setupTestRouter(t)

	reqBody := []byte(`{
		"case_id": "sandbox-case-01",
		"trainee_solution": "Карту разблокировал сразу по звонку, так как клиент назвал паспортные данные"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/evaluate", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен %d", rec.Code)
	}

	var resp domain.SandboxResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Ошибка декодирования: %v", err)
	}

	if resp.IsCorrect {
		t.Errorf("Ожидался is_correct = false для некорректного ответа")
	}
	if resp.Score > 5 {
		t.Errorf("Ожидался низкий балл, получено %d", resp.Score)
	}
	if len(resp.MissedRules) == 0 {
		t.Errorf("Ожидался непустой список missed_rules")
	}
}

func TestSandboxEvaluateCorrectAnswer(t *testing.T) {
	r := setupTestRouter(t)

	reqBody := []byte(`{
		"case_id": "sandbox-case-01",
		"trainee_answer": "Вежливый отказ без проверки кодового слова. Предложил разблокировку через мобильное приложение ВТБ Онлайн либо визит в офис."
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/evaluate", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен %d", rec.Code)
	}

	var resp domain.SandboxResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Ошибка декодирования: %v", err)
	}

	if !resp.IsCorrect {
		t.Errorf("Ожидался is_correct = true для верного регламентного ответа")
	}
	if resp.Score != 9 {
		t.Errorf("Ожидался балл 9, получено %d", resp.Score)
	}
}
