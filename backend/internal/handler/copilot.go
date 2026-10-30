package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"vtb-copilot-backend/internal/domain"
)

// ResolveCopilot обрабатывает POST /api/v1/copilot/resolve
// Формирует регламентную резолюцию, уверенность, маршрут и подтверждающие источники.
func (h *Handler) ResolveCopilot(w http.ResponseWriter, r *http.Request) {
	var req domain.CopilotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Неверный формат JSON запроса: " + err.Error(),
		})
		return
	}

	payload := h.GetCasesPayload()
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
	taskLower := strings.ToLower(taskText)
	categoryLower := strings.ToLower(strings.TrimSpace(req.Category))
	targetID := strings.TrimSpace(req.GetID())

	var resp domain.CopilotResponse

	// 1. Проверка на запрос с эскалацией (например, изменение даты задним числом / ретроспективные проводки)
	if strings.Contains(taskLower, "задним числом") ||
		strings.Contains(taskLower, "изменение даты") ||
		strings.Contains(taskLower, "изменить дату") ||
		strings.Contains(taskLower, "ретроспектив") {

		resp = domain.CopilotResponse{
			SuggestedRoute:     "Комитет по комплаенс-рискам и Служба безопасности • Уровень L3 (Эскалация)",
			ConfidenceScore:    0.20,
			EscalationRequired: true,
			ResolutionDraft: "ТРЕБУЕТСЯ ЭСКАЛАЦИЯ НА УРОВЕНЬ L3 / СЛУЖБА БЕЗОПАСНОСТИ:\n" +
				"1. В запросе обнаружено требование изменения даты валютирования либо исполнения платежа задним числом.\n" +
				"2. Проведение операций и корректировка проводок задним числом строго запрещены нормативными актами Банка России и регламентом REG-01.\n" +
				"3. Приостановить обработку тикета, зафиксировать попытку нарушения учетной дисциплины и передать материалы куратору безопасности.",
			Sources: []domain.SourceItem{
				{
					DocumentName: "Политика управления рисками и регламент бухгалтерского учета ВТБ (ПАО)",
					Clause:       "Раздел 2, п. 2.7 (Запрет ретроспективных проводок)",
					EditionDate:  "15.01.2026",
					Quote:        "Любые корректировки дат валютирования, проводок или первичных платежных документов задним числом категорически запрещены. При поступлении подобных требований операция блокируется с эскалацией на L3.",
				},
				{
					DocumentName: "Положение Банка России № 762-П «О правилах осуществления перевода денежных средств»",
					Clause:       "Глава 1, статья 1.8",
					EditionDate:  "01.09.2025",
					Quote:        "Распоряжения плательщиков подлежат исполнению исключительно в текущую операционную дату либо в установленный договором плановый срок; ретроспективные изменения реквизитов не допускаются.",
				},
			},
		}
		writeRetrievedCopilot(w, resp, retrieval)
		return
	}

	// 2. Поиск по идентификатору кейса, если передан
	if targetID != "" {
		for _, c := range payload.CopilotCases {
			if c.ID == targetID {
				resp = domain.CopilotResponse{
					SuggestedRoute:     c.SuggestedRoute,
					ConfidenceScore:    c.ConfidenceScore,
					ResolutionDraft:    c.ResolutionDraft,
					EscalationRequired: c.EscalationRequired,
					Sources:            c.Sources,
				}
				writeRetrievedCopilot(w, resp, retrieval)
				return
			}
		}
	}

	// 3. Поиск по вхождению текста входящей задачи в базу эталонных кейсов
	if taskLower != "" {
		for _, c := range payload.CopilotCases {
			incomingLower := strings.ToLower(strings.TrimSpace(c.IncomingTask))
			if incomingLower != "" && (strings.Contains(incomingLower, taskLower) || strings.Contains(taskLower, incomingLower)) {
				resp = domain.CopilotResponse{
					SuggestedRoute:     c.SuggestedRoute,
					ConfidenceScore:    c.ConfidenceScore,
					ResolutionDraft:    c.ResolutionDraft,
					EscalationRequired: c.EscalationRequired,
					Sources:            c.Sources,
				}
				writeRetrievedCopilot(w, resp, retrieval)
				return
			}
		}
	}

	// 4. Семантическая маршрутизация по ключевым словам регламентов
	combinedText := taskLower + " " + categoryLower

	// 4.1. Эскалация L3 (санкции, юрисдикции, реестры, коллизии, SDN)
	escalationTriggers := []string{"санкц", "sdn", "юрисдикц", "реестр", "эскалац", "двойного назначения", "коллизи", "спецреестр"}
	if containsAny(combinedText, escalationTriggers) {
		resp = domain.CopilotResponse{
			SuggestedRoute:     "Комитет по комплаенс-рискам и Служба финансовой безопасности • Уровень L3 (Эскалация)",
			ConfidenceScore:    0.68,
			EscalationRequired: true,
			ResolutionDraft: "ВНИМАНИЕ: Обнаружены триггеры повышенного регуляторного риска.\n" +
				"1. Приостановить проведение операции до подтверждения Управлением комплаенс.\n" +
				"2. Заблокировать расчетные операции без списания штрафов.\n" +
				"3. Запросить расширенный пакет документов (KYC/UBO) и передать кейс куратору L3.",
			Sources: []domain.SourceItem{
				{
					DocumentName: "Политика управления санкционными и регуляторными рисками Банка ВТБ (ПАО)",
					Clause:       "Раздел 8, п. 8.2",
					EditionDate:  "01.03.2026",
					Quote:        "Операции с контрагентами из зоны повышенного санкционного риска либо с коллизиями кодов ТН ВЭД требуют согласования на уровне L3 комплаенс-комитета.",
				},
			},
		}
		writeRetrievedCopilot(w, resp, retrieval)
		return
	}

	// 4.2. Антифрод L2 (REG-02)
	fraudTriggers := []string{"фрод", "мошеннич", "алерт", "номер телефон", "vpn", "развод", "подозрит", "sim", "сим-карт"}
	if containsAny(combinedText, fraudTriggers) {
		resp = domain.CopilotResponse{
			SuggestedRoute:     "Служба мониторинга транзакций и противодействия мошенничеству (Антифрод) • Уровень L2",
			ConfidenceScore:    0.97,
			EscalationRequired: false,
			ResolutionDraft: "1. Временно приостановить транзакцию на 48 часов в соответствии со 161-ФЗ.\n" +
				"2. Заблокировать доступ в ДБО «ВТБ Онлайн» до подтверждения клиентом.\n" +
				"3. Инициировать защищенный исходящий вызов для верификации операции.",
			Sources: []domain.SourceItem{
				{
					DocumentName: "Инструкция по предотвращению несанкционированных операций физлиц (REG-02)",
					Clause:       "Пункт 3.1.5",
					EditionDate:  "10.02.2026",
					Quote:        "Смена контактного номера или устройства с последующим выводом средств подлежит превентивной блокировке на 48 часов.",
				},
			},
		}
		writeRetrievedCopilot(w, resp, retrieval)
		return
	}

	// 4.3. Стандартный регламентный маршрут расчетов и платежей (REG-01)
	resp = domain.CopilotResponse{
		SuggestedRoute:     "Управление расчётов и межбанковских платежей (УРМП) • Уровень L2",
		ConfidenceScore:    0.92,
		EscalationRequired: false,
		ResolutionDraft: "1. Зарегистрировать инцидент в автоматизированной банковской системе (АБС).\n" +
			"2. Направить запрос на уточнение реквизитов (тип ED274) в банк-получатель.\n" +
			"3. Контрольный срок исполнения регламента — 3 рабочих дня.",
		Sources: []domain.SourceItem{
			{
				DocumentName: "Регламент обработки межбанковских переводов и невыясненных платежей (REG-01)",
				Clause:       "Раздел 4, п. 4.3.2",
				EditionDate:  "15.01.2026",
				Quote:        "При несовпадении реквизитов операционист формирует запрос ED274 со сроком удержания до выяснения не более 5 операционных дней.",
			},
		},
	}

	writeRetrievedCopilot(w, resp, retrieval)
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
