package llm

import (
	"context"
	"fmt"
	"strings"

	"vtb-copilot-backend/internal/domain"
)

func renderPrompt(template, contextText, task, answer string) string {
	escape := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	// A single replacement pass prevents placeholders within user data from expanding.
	return strings.NewReplacer("{context}", contextText, "{task}", escape.Replace(task), "{trainee_answer}", escape.Replace(answer)).Replace(template)
}

func (c *Client) Generate(ctx context.Context, ticket string, chunks []domain.ChunkMatch) (domain.CopilotResponse, error) {
	if len(chunks) == 0 {
		return domain.CopilotResponse{}, ErrUngrounded
	}
	var contextText strings.Builder
	for _, chunk := range chunks {
		fmt.Fprintf(&contextText, "[%s] %s\n", chunk.Citation, chunk.Text)
	}
	data, err := c.call(ctx, renderPrompt(systemPrompt, contextText.String(), ticket, ""), copilotSchema(chunks))
	if err != nil {
		return domain.CopilotResponse{}, err
	}
	return parseResponse(data, chunks)
}

func copilotSchema(chunks []domain.ChunkMatch) map[string]any {
	citations := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		citations = append(citations, chunk.Citation)
	}
	summary := str()
	summary["maxLength"] = 250
	action := str()
	action["maxLength"] = 200
	actions := array(action)
	actions["maxItems"] = 7
	return obj(map[string]any{
		"task_type":      enumeration("T1", "T2", "T3"),
		"regulation":     nullable(enumeration("REG-01", "REG-02", "REG-03")),
		"priority":       nullable(enumeration("P1", "P2", "P3", "P4")),
		"autonomy_level": integer(1, 3),
		"route":          obj(map[string]any{"executor": nullableString(), "controller": nullableString()}),
		"citations":      array(obj(map[string]any{"citation": enumeration(citations...), "supports": str()})),
		"resolution":     obj(map[string]any{"summary": summary, "actions": actions, "deadlines": nullableString(), "client_draft": nullableString()}),
		"missing_data":   array(str()), "confidence": enumeration("высокая", "средняя", "низкая"),
		"escalation_required": boolean(), "escalation_reason": nullableString(),
		"escalation_brief": nullable(obj(map[string]any{"requested": str(), "detected": str(), "done": str(), "risks": str(), "decision_needed": str(), "deadline": str()})),
	})
}

func parseResponse(data []byte, chunks []domain.ChunkMatch) (domain.CopilotResponse, error) {
	var analysis domain.CopilotAnalysis
	if err := decodeStrict(data, copilotSchema(chunks), &analysis); err != nil {
		return domain.CopilotResponse{}, err
	}
	invalid := func() (domain.CopilotResponse, error) { return domain.CopilotResponse{}, ErrInvalidResponse }
	if analysis.EscalationRequired != (analysis.Confidence == "низкая") || (analysis.Confidence == "высокая" && len(analysis.MissingData) > 0) {
		return invalid()
	}
	if strings.TrimSpace(analysis.Resolution.Summary) == "" {
		return invalid()
	}
	if analysis.EscalationRequired {
		if analysis.TaskType != "T3" || analysis.Route.Executor == nil || *analysis.Route.Executor != "Руководитель смены (через L2)" || analysis.Route.Controller != nil || len(analysis.Resolution.Actions) != 0 || analysis.EscalationBrief == nil || analysis.EscalationReason == nil || analysis.Resolution.Deadlines != nil || analysis.Resolution.ClientDraft != nil || analysis.Resolution.Summary != "Прямого ответа в переданных пунктах нет, задача передаётся на эскалацию." {
			return invalid()
		}
	} else if len(analysis.Resolution.Actions) == 0 || len(analysis.Citations) == 0 || analysis.EscalationBrief != nil || analysis.EscalationReason != nil {
		return invalid()
	}
	for _, action := range analysis.Resolution.Actions {
		if strings.TrimSpace(action) == "" {
			return invalid()
		}
	}
	if analysis.Resolution.ClientDraft != nil {
		hasTemplate := false
		for _, chunk := range chunks {
			if strings.HasPrefix(chunk.Citation, "Шаблон П-") {
				hasTemplate = true
			}
		}
		if analysis.AutonomyLevel != 2 || !hasTemplate {
			return invalid()
		}
	}
	resp := domain.CopilotResponse{Analysis: &analysis, EscalationRequired: analysis.EscalationRequired, Sources: []domain.SourceItem{}}
	resp.ConfidenceScore = map[string]float64{"высокая": 0.9, "средняя": 0.6, "низкая": 0.2}[analysis.Confidence]
	if analysis.Route.Executor != nil {
		resp.SuggestedRoute = *analysis.Route.Executor
	}
	if analysis.Route.Controller != nil {
		resp.SuggestedRoute += " → " + *analysis.Route.Controller
	}
	var draft []string
	draft = append(draft, analysis.Resolution.Summary)
	for i, action := range analysis.Resolution.Actions {
		draft = append(draft, fmt.Sprintf("%d. %s", i+1, action))
	}
	if analysis.Resolution.Deadlines != nil {
		draft = append(draft, "Сроки: "+*analysis.Resolution.Deadlines)
	}
	if analysis.Resolution.ClientDraft != nil {
		draft = append(draft, "Проект сообщения клиенту: "+*analysis.Resolution.ClientDraft)
	}
	if brief := analysis.EscalationBrief; brief != nil {
		draft = append(draft, "Запрошено: "+brief.Requested, "Обнаружено: "+brief.Detected, "Сделано: "+brief.Done, "Риски: "+brief.Risks, "Нужное решение: "+brief.DecisionNeeded, "Срок: "+brief.Deadline)
	}
	resp.ResolutionDraft = strings.Join(draft, "\n")
	seen := map[string]bool{}
	regulationFound := analysis.Regulation == nil
	for _, citation := range analysis.Citations {
		if seen[citation.Citation] {
			return invalid()
		}
		seen[citation.Citation] = true
		for _, chunk := range chunks {
			if chunk.Citation == citation.Citation {
				if analysis.Regulation != nil && chunk.Document == *analysis.Regulation {
					regulationFound = true
				}
				if analysis.EscalationRequired && !strings.Contains(strings.ToLower(chunk.Text), "эскал") {
					return invalid()
				}
				if len(strings.Fields(citation.Supports)) > 15 {
					return invalid()
				}
				resp.Sources = append(resp.Sources, domain.SourceItem{DocumentName: chunk.DocumentTitle, Clause: chunk.Citation, Quote: chunk.Text})
			}
		}
	}
	if !regulationFound {
		return invalid()
	}
	return resp, nil
}
