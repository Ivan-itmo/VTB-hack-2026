package domain

type CopilotAnalysis struct {
	TaskType      string  `json:"task_type"`
	Regulation    *string `json:"regulation"`
	Priority      *string `json:"priority"`
	AutonomyLevel int     `json:"autonomy_level"`
	Route         struct {
		Executor   *string `json:"executor"`
		Controller *string `json:"controller"`
	} `json:"route"`
	Citations []struct {
		Citation string `json:"citation"`
		Supports string `json:"supports"`
	} `json:"citations"`
	Resolution struct {
		Summary     string   `json:"summary"`
		Actions     []string `json:"actions"`
		Deadlines   *string  `json:"deadlines"`
		ClientDraft *string  `json:"client_draft"`
	} `json:"resolution"`
	MissingData        []string         `json:"missing_data"`
	Confidence         string           `json:"confidence"`
	EscalationRequired bool             `json:"escalation_required"`
	EscalationReason   *string          `json:"escalation_reason"`
	EscalationBrief    *EscalationBrief `json:"escalation_brief"`
}

type EscalationBrief struct {
	Requested      string `json:"requested"`
	Detected       string `json:"detected"`
	Done           string `json:"done"`
	Risks          string `json:"risks"`
	DecisionNeeded string `json:"decision_needed"`
	Deadline       string `json:"deadline"`
}

type EvaluationCriterion struct {
	Criterion string `json:"criterion"`
	Points    int    `json:"points"`
}

type SandboxEvaluation struct {
	CaseID          *string           `json:"case_id"`
	Score           int               `json:"score"`
	MaxScore        int               `json:"max_score"`
	Passed          bool              `json:"passed"`
	CriticalError   bool              `json:"critical_error"`
	CriteriaResults []CriterionResult `json:"criteria_results"`
	MissedPoints    []MissedPoint     `json:"missed_points"`
	ErrorsFound     []EvaluationError `json:"errors_found"`
	Feedback        string            `json:"feedback"`
	MainLesson      string            `json:"main_lesson"`
	Flags           []string          `json:"flags"`
}

type CriterionResult struct {
	Criterion     string `json:"criterion"`
	MaxPoints     int    `json:"max_points"`
	AwardedPoints int    `json:"awarded_points"`
	Verdict       string `json:"verdict"`
	Comment       string `json:"comment"`
}
type MissedPoint struct {
	Point     string  `json:"point"`
	Criterion string  `json:"criterion"`
	Citation  *string `json:"citation"`
}
type EvaluationError struct {
	Description string  `json:"description"`
	Critical    bool    `json:"critical"`
	Citation    *string `json:"citation"`
}
