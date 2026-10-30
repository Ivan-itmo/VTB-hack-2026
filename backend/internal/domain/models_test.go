package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTrainingContextIsPrivate(t *testing.T) {
	var card SandboxCase
	err := json.Unmarshal([]byte(`{"id":"SANDBOX-01","task_description":"Условие","reference_answer":"SECRET REFERENCE","expected_rule":"REG-01","evaluation_criteria":[{"criterion":"Проверить","points":10}]}`), &card)
	if err != nil || card.IncomingTask != "Условие" || card.ReferenceAnswer == "" || len(card.EvaluationCriteria) != 1 {
		t.Fatalf("card not loaded: %v", err)
	}
	data, _ := json.Marshal(card)
	if strings.Contains(string(data), "SECRET REFERENCE") || strings.Contains(string(data), "reference_answer") || strings.Contains(string(data), "evaluation_criteria") {
		t.Fatal("training answers exposed")
	}
}
