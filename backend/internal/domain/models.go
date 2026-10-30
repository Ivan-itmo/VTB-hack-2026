package domain

import "encoding/json"

// SourceItem представляет подтверждающую цитату или нормативное обоснование из регламентов (Strict Grounding).
type SourceItem struct {
	DocumentName string `json:"document_name"`
	Clause       string `json:"clause"`
	EditionDate  string `json:"edition_date"`
	Quote        string `json:"quote"`
}

// CopilotRequest представляет входящий запрос на автоматическую резолюцию и маршрутизацию.
// Поддерживает как строгие контрактные поля (ticket_id, task_description),
// так и совместимость с фронтендом (case_id, task_text).
type CopilotRequest struct {
	TicketID        string `json:"ticket_id,omitempty"`
	Category        string `json:"category,omitempty"`
	TaskDescription string `json:"task_description,omitempty"`

	// Совместимость с фронтенд-клиентом
	CaseID   string `json:"case_id,omitempty"`
	TaskText string `json:"task_text,omitempty"`
}

// UnmarshalJSON обеспечивает прозрачное сопоставление полей ticket_id/case_id и task_description/task_text.
func (r *CopilotRequest) UnmarshalJSON(data []byte) error {
	type Alias CopilotRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Синхронизация идентификатора кейса/тикета
	if r.TicketID == "" && r.CaseID != "" {
		r.TicketID = r.CaseID
	}
	if r.CaseID == "" && r.TicketID != "" {
		r.CaseID = r.TicketID
	}

	// Синхронизация текста задачи
	if r.TaskDescription == "" && r.TaskText != "" {
		r.TaskDescription = r.TaskText
	}
	if r.TaskText == "" && r.TaskDescription != "" {
		r.TaskText = r.TaskDescription
	}

	return nil
}

// GetTask возвращает текст входящей задачи независимо от использованного ключа JSON.
func (r *CopilotRequest) GetTask() string {
	if r.TaskDescription != "" {
		return r.TaskDescription
	}
	return r.TaskText
}

// GetID возвращает идентификатор кейса/тикета.
func (r *CopilotRequest) GetID() string {
	if r.TicketID != "" {
		return r.TicketID
	}
	return r.CaseID
}

// CopilotResponse представляет ответ ассистента с рекомендацией, черновиком и источниками.
type CopilotResponse struct {
	Analysis           *CopilotAnalysis `json:"analysis,omitempty"`
	RetrievedChunks    []ChunkMatch     `json:"retrieved_chunks"`
	RetrievalScore     float64          `json:"retrieval_score"`
	SuggestedRoute     string           `json:"suggested_route"`
	ConfidenceScore    float64          `json:"confidence_score"`
	ResolutionDraft    string           `json:"resolution_draft"`
	EscalationRequired bool             `json:"escalation_required"`
	Sources            []SourceItem     `json:"sources"`
}

// SandboxRequest представляет входящий запрос на проверку ответа стажера в тренажере.
// Поддерживает trainee_answer и trainee_solution для обратной совместимости.
type SandboxRequest struct {
	CaseID        string `json:"case_id,omitempty"`
	TraineeAnswer string `json:"trainee_answer,omitempty"`

	// Совместимость с фронтендом
	TaskText        string `json:"task_text,omitempty"`
	TraineeSolution string `json:"trainee_solution,omitempty"`
}

// UnmarshalJSON обеспечивает прозрачное сопоставление trainee_answer и trainee_solution.
func (r *SandboxRequest) UnmarshalJSON(data []byte) error {
	type Alias SandboxRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if r.TraineeAnswer == "" && r.TraineeSolution != "" {
		r.TraineeAnswer = r.TraineeSolution
	}
	if r.TraineeSolution == "" && r.TraineeAnswer != "" {
		r.TraineeSolution = r.TraineeAnswer
	}

	return nil
}

// GetAnswer возвращает ответ стажера независимо от имени поля в JSON.
func (r *SandboxRequest) GetAnswer() string {
	if r.TraineeAnswer != "" {
		return r.TraineeAnswer
	}
	return r.TraineeSolution
}

// SandboxResponse представляет вердикт проверки ответа стажера.
type SandboxResponse struct {
	Evaluation            *SandboxEvaluation `json:"evaluation,omitempty"`
	EvaluationUnavailable bool               `json:"evaluation_unavailable,omitempty"`
	EscalationRequired    bool               `json:"escalation_required,omitempty"`
	IsCorrect             bool               `json:"is_correct"`
	Score                 int                `json:"score"`
	Feedback              string             `json:"feedback"`
	MissedRules           []string           `json:"missed_rules"`
	CorrectResolution     string             `json:"correct_resolution"`
}

// CopilotCase описывает демо-кейс ко-пилота из demo_cases.json.
type CopilotCase struct {
	TaskDescription    string       `json:"task_description,omitempty"`
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	Category           string       `json:"category"`
	RegCode            string       `json:"reg_code"`
	IncomingTask       string       `json:"incoming_task"`
	SuggestedRoute     string       `json:"suggested_route"`
	ConfidenceScore    float64      `json:"confidence_score"`
	EscalationRequired bool         `json:"escalation_required"`
	ResolutionDraft    string       `json:"resolution_draft"`
	Sources            []SourceItem `json:"sources"`
}

// SandboxCase описывает тренировочный кейс песочницы из demo_cases.json.
type SandboxCase struct {
	TraineeTask         string                `json:"trainee_task,omitempty"`
	Hints               []string              `json:"hints,omitempty"`
	ReferenceAnswer     string                `json:"-"`
	ExpectedRule        string                `json:"-"`
	TypicalMistakes     []string              `json:"-"`
	EvaluationCriteria  []EvaluationCriterion `json:"-"`
	ID                  string                `json:"id"`
	Title               string                `json:"title"`
	IncomingTask        string                `json:"incoming_task"`
	SampleTraineeAnswer string                `json:"sample_trainee_answer"`
	Score               int                   `json:"score"`
	IsCorrect           bool                  `json:"is_correct"`
	Feedback            string                `json:"feedback"`
	MissedRules         []string              `json:"missed_rules"`
	CorrectResolution   string                `json:"correct_resolution"`
}

// Keep grading context server-side; /cases never exposes reference answers.
func (c *SandboxCase) UnmarshalJSON(data []byte) error {
	type alias SandboxCase
	var wire struct {
		*alias
		TaskDescription    string                `json:"task_description"`
		ReferenceAnswer    string                `json:"reference_answer"`
		ExpectedRule       string                `json:"expected_rule"`
		TypicalMistakes    []string              `json:"typical_mistakes"`
		EvaluationCriteria []EvaluationCriterion `json:"evaluation_criteria"`
	}
	wire.alias = (*alias)(c)
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if c.IncomingTask == "" {
		c.IncomingTask = wire.TaskDescription
	}
	c.ReferenceAnswer, c.ExpectedRule = wire.ReferenceAnswer, wire.ExpectedRule
	c.TypicalMistakes, c.EvaluationCriteria = wire.TypicalMistakes, wire.EvaluationCriteria
	return nil
}

// DemoCasesPayload структура всего файла demo_cases.json.
type DemoCasesPayload struct {
	CopilotCases []CopilotCase `json:"copilot_cases"`
	SandboxCases []SandboxCase `json:"sandbox_cases"`
}

// HealthResponse ответ корневого эндпоинта проверки работоспособности.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}
