package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"vtb-copilot-backend/internal/domain"
)

func loadTestIndex(t *testing.T) *Index {
	t.Helper()
	idx, err := Load("../../data/rules.json", DefaultThreshold)
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

func TestLoadAllChunks(t *testing.T) {
	data, err := os.ReadFile("../../data/rules.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Chunks []domain.Chunk `json:"chunks"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	idx := loadTestIndex(t)
	if idx.Len() != len(payload.Chunks) {
		t.Fatalf("loaded %d of %d chunks", idx.Len(), len(payload.Chunks))
	}
	// Search must not reread files and must preserve original text/citations.
	for _, chunk := range payload.Chunks {
		for _, match := range idx.Search(chunk.Text).Chunks {
			found := false
			for _, original := range payload.Chunks {
				if match.Chunk == original {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("modified or fabricated chunk: %+v", match)
			}
		}
	}
}

func TestRelevantSearch(t *testing.T) {
	idx := loadTestIndex(t)
	for _, tc := range []struct{ text, want string }{
		{"Неверный ИНН КПП получателя, ошибка в БИК", "REG-01-1.2"},
		{"Платёжное поручение №304 от ООО «Альфа-Трейд» на сумму 7 800 000 руб. Ошибка контроля: в назначении платежа не указан номер договора. Поручение поступило сегодня.", "REG-01-1.2"},
		{"Сумма выше 5 000 000 рублей", "REG-01-2.2"},
		{"383-П уточнение реквизитов распоряжения", "REG-01-1.1"},
		{"161-ФЗ антифрод приостановка перевода", "GEN-2.2"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			result := idx.Search(tc.text)
			t.Logf("score=%.3f", result.BestScore)
			if result.EscalationRequired {
				t.Fatalf("relevant query escalated: %.3f", result.BestScore)
			}
			if len(result.Chunks) < 2 || len(result.Chunks) > 3 {
				t.Fatalf("unexpected match count %d", len(result.Chunks))
			}
			found := false
			for i, match := range result.Chunks {
				if match.ID == tc.want {
					found = true
				}
				if i > 0 && match.Score > result.Chunks[i-1].Score {
					t.Fatal("results are not sorted")
				}
			}
			if !found {
				t.Fatalf("expected %s in top 3", tc.want)
			}
		})
	}
}

func TestUnknownAndEmptyEscalate(t *testing.T) {
	idx := loadTestIndex(t)
	for _, text := range []string{"", "  ?! ", "пожалуйста мне нужно", "Космический телескоп наблюдает галактики", "Космический телескоп наблюдает галактики ИНН"} {
		result := idx.Search(text)
		if !result.EscalationRequired || result.BestScore >= DefaultThreshold {
			t.Fatalf("unrelated query accepted: %q %+v", text, result)
		}
	}
}

func TestNormalization(t *testing.T) {
	idx := loadTestIndex(t)
	for _, pair := range [][2]string{
		{"5 000 000", "5000000"}, {"5\u00a0000\u00a0000", "5000000"},
		{"161‑ФЗ", "161-ФЗ"}, {"СЧЁТ ПОЛУЧАТЕЛЯ", "счет получателя"},
	} {
		a, _ := json.Marshal(idx.Search(pair[0]))
		b, _ := json.Marshal(idx.Search(pair[1]))
		if string(a) != string(b) {
			t.Fatalf("normalization mismatch: %q / %q", pair[0], pair[1])
		}
	}
}

func TestLoadValidationAndThreshold(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.json")
	for _, data := range []string{`broken`, `{"chunks":[]}`, `{"chunks":[{"id":"x"}]}`} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path, DefaultThreshold); err == nil {
			t.Fatalf("accepted invalid data: %s", data)
		}
	}
	if _, err := Load(path+"missing", DefaultThreshold); err == nil {
		t.Fatal("missing file accepted")
	}
	idx := loadTestIndex(t)
	score := idx.Search("ИНН КПП получателя").BestScore
	idx.threshold = score
	if idx.Search("ИНН КПП получателя").EscalationRequired {
		t.Fatal("score at threshold should be accepted")
	}
	idx.threshold = score + 0.001
	if !idx.Search("ИНН КПП получателя").EscalationRequired {
		t.Fatal("score below threshold should escalate")
	}
}

func TestConcurrentSearch(t *testing.T) {
	idx := loadTestIndex(t)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); idx.Search("ИНН КПП 383-П") }()
	}
	wg.Wait()
}

func BenchmarkSearch(b *testing.B) {
	idx, err := Load("../../data/rules.json", DefaultThreshold)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Search("Неверный ИНН КПП получателя платёжного поручения на 5 000 000 рублей")
	}
}
