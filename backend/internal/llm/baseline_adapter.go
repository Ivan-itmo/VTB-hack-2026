// Package llm реализует абстракцию шлюза к языковым моделям и строгую валидацию цитирования нормативной базы.
// BaselineAPIAdapter реализует интерфейс LLMProvider для валидации JSON-схем и сквозного тестирования в Спринте №1.
// Архитектура подготовлена к переводу во 2-м спринте на изолированный On-Prem контур (vLLM / Ollama) и отечественные корпоративные шлюзы (GigaChat API) в соответствии с 152-ФЗ.
package llm

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"vtb-copilot-backend/internal/domain"
)

const DefaultModel = "gemini-3.5-flash-lite"
const maxResponseBytes = 1 << 20

//go:embed system_prompt.txt
var systemPrompt string

//go:embed sandbox_prompt.txt
var sandboxPrompt string

var (
	ErrMissingContext  = errors.New("missing evaluation context")
	ErrUnavailable     = errors.New("llm not configured")
	ErrTimeout         = errors.New("llm timeout")
	ErrProvider        = errors.New("llm provider error")
	ErrInvalidResponse = errors.New("invalid llm response")
	ErrUngrounded      = errors.New("llm citation not grounded")
	modelPattern       = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

// LLMProvider определяет контракт интеграции с языковыми моделями для контуров Ко-пилота и Песочницы.
type LLMProvider interface {
	Generate(ctx context.Context, ticket string, chunks []domain.ChunkMatch) (domain.CopilotResponse, error)
	Evaluate(ctx context.Context, card domain.SandboxCase, answer string, chunks []domain.ChunkMatch) (domain.SandboxResponse, error)
}

type Config struct {
	APIKey  string
	Model   string
	Timeout time.Duration
	// BaseURL permits a local mock in tests or an On-Prem LLM gateway (vLLM/Ollama).
	BaseURL string
}

// BaselineAPIAdapter реализует интерфейс LLMProvider.
type BaselineAPIAdapter struct {
	config Config
	http   *http.Client
}

// Client сохранён как псевдоним типа для полной обратной совместимости.
type Client = BaselineAPIAdapter

// Compile-time check, что BaselineAPIAdapter реализует LLMProvider.
var _ LLMProvider = (*BaselineAPIAdapter)(nil)

// FromEnv инициализирует адаптер из переменных окружения с приоритетом LLM_* и обратной совместимостью с GEMINI_*.
func FromEnv() (*BaselineAPIAdapter, error) {
	timeout := 25 * time.Second
	timeoutVal := os.Getenv("LLM_TIMEOUT")
	if timeoutVal == "" {
		timeoutVal = os.Getenv("GEMINI_TIMEOUT")
	}
	if timeoutVal != "" {
		parsed, err := time.ParseDuration(timeoutVal)
		if err != nil {
			return nil, errors.New("LLM_TIMEOUT must be a duration, for example 25s")
		}
		timeout = parsed
	}

	apiKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	}

	model := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if model == "" {
		model = strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	}
	if model == "" {
		model = DefaultModel
	}

	baseURL := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("GEMINI_BASE_URL"))
	}

	return New(Config{
		APIKey:  apiKey,
		Model:   model,
		Timeout: timeout,
		BaseURL: baseURL,
	})
}

// New создает экземпляр BaselineAPIAdapter с валидацией конфигурации.
func New(config Config) (*BaselineAPIAdapter, error) {
	config.APIKey = strings.TrimSpace(config.APIKey)
	if config.Model == "" {
		config.Model = DefaultModel
	}
	if !modelPattern.MatchString(config.Model) {
		return nil, errors.New("invalid LLM model name")
	}
	if config.Timeout == 0 {
		config.Timeout = 25 * time.Second
	}
	if config.Timeout < 0 || config.Timeout > 30*time.Second {
		return nil, errors.New("LLM provider timeout must be positive and at most 30s")
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	return &BaselineAPIAdapter{config: config, http: &http.Client{
		Timeout: config.Timeout,
		// Never forward the credential to a redirect destination.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

type part struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}
type content struct {
	Parts []part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

func (c *BaselineAPIAdapter) call(ctx context.Context, prompt string, schema map[string]any) ([]byte, error) {
	var empty []byte
	if c.config.APIKey == "" {
		return nil, ErrUnavailable
	}
	payload := struct {
		System     content        `json:"systemInstruction"`
		Contents   []content      `json:"contents"`
		Generation map[string]any `json:"generationConfig"`
	}{
		System:   content{Parts: []part{{Text: prompt}}},
		Contents: []content{{Role: "user", Parts: []part{{Text: "Выполни задачу по системному промпту. Верни только JSON."}}}},
		Generation: map[string]any{
			"temperature": 0.1, "maxOutputTokens": 4096,
			"responseMimeType": "application/json", "responseJsonSchema": schema,
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return empty, ErrInvalidResponse
	}
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.config.BaseURL, "/")+"/models/"+c.config.Model+":generateContent", bytes.NewReader(data))
	if err != nil {
		return empty, ErrProvider
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.config.APIKey)
	res, err := c.http.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return empty, ErrTimeout
		}
		return empty, ErrProvider
	}
	defer res.Body.Close()
	// Do not propagate provider bodies, which may contain prompts or credentials.
	if res.StatusCode != http.StatusOK {
		return empty, fmt.Errorf("%w: HTTP %d", ErrProvider, res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes+1))
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return empty, ErrTimeout
		}
		return empty, ErrProvider
	}
	if len(body) > maxResponseBytes {
		return empty, ErrInvalidResponse
	}
	var envelope struct {
		Candidates []struct {
			Content      content `json:"content"`
			FinishReason string  `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason"`
		} `json:"promptFeedback"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return empty, ErrInvalidResponse
	}
	if envelope.PromptFeedback.BlockReason != "" || len(envelope.Candidates) != 1 || envelope.Candidates[0].FinishReason != "STOP" {
		return empty, ErrInvalidResponse
	}
	var text strings.Builder
	for _, p := range envelope.Candidates[0].Content.Parts {
		if !p.Thought {
			text.WriteString(p.Text)
		}
	}
	return []byte(text.String()), nil
}
