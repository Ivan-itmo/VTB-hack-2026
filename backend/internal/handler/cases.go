package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"vtb-copilot-backend/internal/domain"
	"vtb-copilot-backend/internal/search"
)

// Handler инкапсулирует логику обработки запросов и доступ к нормативным демо-кейсам.
type Handler struct {
	dataPath string
	payload  domain.DemoCasesPayload
	mu       sync.RWMutex
	rules    *search.Index
}

// New создает новый экземпляр Handler и загружает данные из demo_cases.json.
func New(customPath string) (*Handler, error) {
	h := &Handler{
		dataPath: customPath,
	}

	if err := h.reloadCases(); err != nil {
		return nil, err
	}
	rulesPath := os.Getenv("RULES_PATH")
	if rulesPath == "" {
		casesPath, err := h.locateDataFile()
		if err != nil {
			return nil, err
		}
		rulesPath = filepath.Join(filepath.Dir(casesPath), "rules.json")
	}
	index, err := search.Load(rulesPath, search.DefaultThreshold)
	if err != nil {
		return nil, err
	}
	h.rules = index
	log.Printf("[Search] Загружено чанков: %d (файл: %s)", index.Len(), rulesPath)

	return h, nil
}

// locateDataFile производит надежный поиск demo_cases.json по нескольким путям.
func (h *Handler) locateDataFile() (string, error) {
	candidates := []string{}

	if h.dataPath != "" {
		candidates = append(candidates, h.dataPath)
	}

	if envPath := os.Getenv("DEMO_CASES_PATH"); envPath != "" {
		candidates = append(candidates, envPath)
	}

	// Относительные пути от текущей рабочей директории
	candidates = append(candidates,
		"data/demo_cases.json",
		"backend/data/demo_cases.json",
		"../data/demo_cases.json",
		"../../data/demo_cases.json",
		"demo_cases.json",
	)

	// Пути относительно запущенного бинарника
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates,
			filepath.Join(execDir, "data", "demo_cases.json"),
			filepath.Join(execDir, "..", "data", "demo_cases.json"),
			filepath.Join(execDir, "..", "..", "data", "demo_cases.json"),
			filepath.Join(execDir, "..", "..", "backend", "data", "demo_cases.json"),
		)
	}

	for _, path := range candidates {
		if abs, err := filepath.Abs(path); err == nil {
			if fi, err := os.Stat(abs); err == nil && !fi.IsDir() {
				return abs, nil
			}
		}
	}

	return "", fmt.Errorf("demo_cases.json не найден ни по одному из проверенных путей: %v", candidates)
}

// reloadCases считывает и валидирует demo_cases.json.
func (h *Handler) reloadCases() error {
	filePath, err := h.locateDataFile()
	if err != nil {
		return err
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла %s: %w", filePath, err)
	}

	var payload domain.DemoCasesPayload
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return fmt.Errorf("ошибка парсинга JSON из %s: %w", filePath, err)
	}

	// Гарантируем, что слайсы инициализированы (для сериализации в [] вместо null)
	if payload.CopilotCases == nil {
		payload.CopilotCases = []domain.CopilotCase{}
	}
	for i := range payload.CopilotCases {
		if payload.CopilotCases[i].Sources == nil {
			payload.CopilotCases[i].Sources = []domain.SourceItem{}
		}
	}

	if payload.SandboxCases == nil {
		payload.SandboxCases = []domain.SandboxCase{}
	}
	for i := range payload.SandboxCases {
		if payload.SandboxCases[i].MissedRules == nil {
			payload.SandboxCases[i].MissedRules = []string{}
		}
	}

	h.mu.Lock()
	h.payload = payload
	h.mu.Unlock()

	log.Printf("[Handler] Успешно загружено демо-кейсов: copilot=%d, sandbox=%d (файл: %s)",
		len(payload.CopilotCases), len(payload.SandboxCases), filePath)

	return nil
}

// GetCasesPayload возвращает потокобезопасную копию данных кейсов.
func (h *Handler) GetCasesPayload() domain.DemoCasesPayload {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.payload
}

// Root обрабатывает GET / — проверка работоспособности (Healthcheck).
func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	resp := domain.HealthResponse{
		Status:  "ok",
		Service: "vtb-copilot-go",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetDemoCases обрабатывает GET /api/v1/cases — чтение и выдачу demo_cases.json.
func (h *Handler) GetDemoCases(w http.ResponseWriter, r *http.Request) {
	payload := h.GetCasesPayload()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("[Handler.GetDemoCases] Ошибка сериализации JSON: %v", err)
	}
}
