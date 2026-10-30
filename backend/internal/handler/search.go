package handler

import (
	"encoding/json"
	"net/http"

	"vtb-copilot-backend/internal/domain"
)

// SearchChunks exposes retrieval independently from resolution generation.
func (h *Handler) SearchChunks(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат JSON"})
		return
	}
	writeJSON(w, http.StatusOK, h.rules.Search(req.Text))
}

func writeRetrievedCopilot(w http.ResponseWriter, resp domain.CopilotResponse, retrieval domain.SearchResult) {
	resp.RetrievedChunks = retrieval.Chunks
	resp.RetrievalScore = retrieval.BestScore
	resp.EscalationRequired = resp.EscalationRequired || retrieval.EscalationRequired
	// The LLM client validates selected sources. Fallbacks attach retrieved candidates.
	if resp.Sources == nil {
		resp.Sources = make([]domain.SourceItem, 0, len(retrieval.Chunks))
		for _, match := range retrieval.Chunks {
			resp.Sources = append(resp.Sources, domain.SourceItem{
				DocumentName: match.DocumentTitle,
				Clause:       match.Citation,
				Quote:        match.Text,
			})
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
