package domain

// Chunk сохраняет исходный объект rules.json, включая точную цитату в Text.
type Chunk struct {
	ID            string `json:"id"`
	Document      string `json:"document"`
	DocumentTitle string `json:"document_title"`
	Section       string `json:"section"`
	SectionTitle  string `json:"section_title"`
	Clause        string `json:"clause"`
	Citation      string `json:"citation"`
	Text          string `json:"text"`
}

type ChunkMatch struct {
	Chunk
	Score float64 `json:"score"`
}

// Score — мера лексической релевантности [0, 1], не вероятность верного решения.
type SearchResult struct {
	Chunks             []ChunkMatch `json:"chunks"`
	BestScore          float64      `json:"best_score"`
	EscalationRequired bool         `json:"escalation_required"`
}
