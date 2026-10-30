package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"vtb-copilot-backend/internal/domain"
)

// ResolveCopilot обрабатывает POST /api/v1/copilot/resolve
// Формирует регламентную резолюцию, уверенность, маршрут и подтверждающие источники.
func (h *Handler) ResolveCopilot(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var req domain.CopilotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Неверный формат JSON запроса: " + err.Error(),
		})
		return
	}

	taskText := strings.TrimSpace(req.GetTask())
	retrieval := h.rules.Search(taskText)
	if retrieval.EscalationRequired {
		writeRetrievedCopilot(w, domain.CopilotResponse{
			SuggestedRoute:     "Профильный эксперт • Эскалация",
			EscalationRequired: true,
			ResolutionDraft:    "Недостаточно релевантных нормативных оснований. Передайте задачу эксперту; автоматическая резолюция не сформирована.",
		}, retrieval)
		return
	}
	if h.generator == nil {
		writeRetrievedCopilot(w, safeEscalation("Модель не настроена. Передайте задачу эксперту."), retrieval)
		return
	}
	resp, err := h.generator.Generate(r.Context(), taskText, retrieval.Chunks)
	if err != nil {
		// No ticket, key, provider response or generated text is written to logs.
		log.Printf("[LLM] Генерация недоступна: %v", err)
		writeRetrievedCopilot(w, safeEscalation("Не удалось получить проверенный ответ модели. Передайте задачу эксперту."), retrieval)
		return
	}
	writeRetrievedCopilot(w, resp, retrieval)
}

func safeEscalation(message string) domain.CopilotResponse {
	return domain.CopilotResponse{
		SuggestedRoute:     "Профильный эксперт • Эскалация",
		EscalationRequired: true,
		ResolutionDraft:    message,
	}
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
