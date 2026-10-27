package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"vtb-copilot-backend/internal/domain"
)

// EvaluateSandbox обрабатывает POST /api/v1/sandbox/evaluate
// Оценивает регламентное решение стажера, выявляет нарушения и формирует фидбек наставника.
func (h *Handler) EvaluateSandbox(w http.ResponseWriter, r *http.Request) {
	var req domain.SandboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Неверный формат JSON запроса: " + err.Error(),
		})
		return
	}

	payload := h.GetCasesPayload()
	targetID := strings.TrimSpace(req.CaseID)
	taskText := strings.TrimSpace(req.TaskText)
	traineeAnswer := strings.TrimSpace(req.GetAnswer())
	ansLower := strings.ToLower(traineeAnswer)

	var targetCase *domain.SandboxCase

	// 1. Поиск кейса по идентификатору
	if targetID != "" {
		for i := range payload.SandboxCases {
			if payload.SandboxCases[i].ID == targetID {
				targetCase = &payload.SandboxCases[i]
				break
			}
		}
	}

	// 2. Поиск кейса по вхождению текста задачи
	if targetCase == nil && taskText != "" {
		taskLower := strings.ToLower(taskText)
		for i := range payload.SandboxCases {
			incomingLower := strings.ToLower(strings.TrimSpace(payload.SandboxCases[i].IncomingTask))
			if incomingLower != "" && (strings.Contains(incomingLower, taskLower) || strings.Contains(taskLower, incomingLower)) {
				targetCase = &payload.SandboxCases[i]
				break
			}
		}
	}

	var resp domain.SandboxResponse

	// 3. Анализ ответа на известный кейс
	if targetCase != nil {
		sampleAns := strings.ToLower(strings.TrimSpace(targetCase.SampleTraineeAnswer))
		isBadMatch := false
		if len([]rune(sampleAns)) > 10 && (strings.Contains(ansLower, sampleAns) || strings.Contains(sampleAns, ansLower)) {
			isBadMatch = true
		}

		// Проверка ключевых регламентных терминов и действий
		hasProperKeywords := false
		incomingLower := strings.ToLower(targetCase.IncomingTask)
		if strings.Contains(incomingLower, "пин") {
			hasProperKeywords = containsAny(ansLower, []string{"кодов", "слово", "отказ", "мобильн", "вспомн", "офис", "ебс"})
		} else if strings.Contains(incomingLower, "сбп") {
			hasProperKeywords = containsAny(ansLower, []string{"безотзыв", "невозможно", "согласи", "добровольн", "банк получател"})
		} else if strings.Contains(incomingLower, "овердрафт") {
			hasProperKeywords = containsAny(ansLower, []string{"кэп", "эцп", "комитет", "дбо", "онлайн", "отказ по скан"})
		} else {
			hasProperKeywords = len([]rune(ansLower)) >= 30
		}

		if isBadMatch || !hasProperKeywords {
			missed := targetCase.MissedRules
			if missed == nil {
				missed = []string{}
			}
			resp = domain.SandboxResponse{
				IsCorrect:         false,
				Score:             targetCase.Score,
				Feedback:          targetCase.Feedback,
				MissedRules:       missed,
				CorrectResolution: targetCase.CorrectResolution,
			}
		} else {
			resp = domain.SandboxResponse{
				IsCorrect:         true,
				Score:             9,
				Feedback:          "Отличный регламентный ответ! Вы точно определили ограничения регламента, не поддались на эмоциональное давление клиента и предложили законный алгоритм решения проблемы.",
				MissedRules:       []string{},
				CorrectResolution: targetCase.CorrectResolution,
			}
		}

		writeJSON(w, http.StatusOK, resp)
		return
	}

	// 4. Оценка решения для свободной/произвольной задачи
	if len([]rune(ansLower)) < 25 {
		resp = domain.SandboxResponse{
			IsCorrect: false,
			Score:     2,
			Feedback:  "Ответ слишком краткий и не содержит регламентной аргументации. Решение бэк-офиса должно опираться на нормативную базу банка и содержать пошаговый алгоритм действий.",
			MissedRules: []string{
				"Стандарт оформления служебных резолюций ВТБ (п. 1.3)",
				"Требование ссылок на нормативно-правовую базу при коммуникации с клиентами",
			},
			CorrectResolution: "Необходимо зарегистрировать обращение, зафиксировать идентификационные параметры клиента, проверить транзакцию по чек-листу безопасности и направить официальный ответ с разъяснением порядка урегулирования.",
		}
	} else {
		resp = domain.SandboxResponse{
			IsCorrect: true,
			Score:     8,
			Feedback:  "Решение в целом соответствует стандартам обработки обращений бэк-офиса. Рекомендуется дополнить резолюцию точными сроками исполнения (SLA) и кодами сообщений АБС.",
			MissedRules: []string{
				"Регламент соблюдения контрольных сроков обработки (SLA L1/L2)",
			},
			CorrectResolution: "Зафиксировать инцидент в АБС, провести верификацию по регламенту безопасности, уведомить смежные подразделения и направить клиенту статус исполнения через мобильное приложение.",
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
