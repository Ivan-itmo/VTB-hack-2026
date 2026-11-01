package llm

import (
	"bytes"
	"encoding/json"
	"strings"
)

// CleanJSON выполняет предварительную очистку сырого ответа модели:
// - удаляет BOM (\uFEFF);
// - заменяет неразрывные пробелы (\u00A0) на стандартные пробелы;
// - очищает обрамление ```json ... ``` и тройные кавычки ```;
// - извлекает подстроку от первого символа '{' до последнего символа '}';
// - фильтрует непечатные и контрольные символы (сохраняя \t, \n, \r).
func CleanJSON(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	// 1. Удаление UTF-8 BOM
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	s := string(data)

	// 2. Нормализация неразрывных пробелов
	s = strings.ReplaceAll(s, "\u00a0", " ")

	// 3. Удаление Markdown-блоков ```json и ```
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "```") {
		if idx := strings.IndexByte(trimmed, '\n'); idx != -1 {
			trimmed = trimmed[idx+1:]
		} else {
			trimmed = strings.TrimPrefix(trimmed, "```")
		}
		if idx := strings.LastIndex(trimmed, "```"); idx != -1 {
			trimmed = trimmed[:idx]
		}
		s = trimmed
	}

	// 4. Извлечение подстроки от первого '{' до последнего '}'
	first := strings.IndexByte(s, '{')
	last := strings.LastIndexByte(s, '}')
	if first != -1 && last > first {
		s = s[first : last+1]
	}

	// 5. Удаление непечатных управляющих символов (кроме \t, \n, \r)
	var buf strings.Builder
	buf.Grow(len(s))
	for _, r := range s {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			continue
		}
		buf.WriteRune(r)
	}

	return []byte(buf.String())
}

// Reject duplicate keys as well as malformed JSON before typed decoding.
func uniqueKeys(dec *json.Decoder) bool {
	return uniqueKeysAtDepth(dec, 0)
}

func uniqueKeysAtDepth(dec *json.Decoder, depth int) bool {
	if depth > 32 {
		return false
	}
	token, err := dec.Token()
	if err != nil {
		return false
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return true
	}
	if delim != '{' && delim != '[' {
		return false
	}
	seen := make(map[string]bool)
	for dec.More() {
		if delim == '{' {
			key, err := dec.Token()
			if err != nil {
				return false
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return false
			}
			seen[name] = true
		}
		if !uniqueKeysAtDepth(dec, depth+1) {
			return false
		}
	}
	end, err := dec.Token()
	return err == nil && (delim == '{' && end == json.Delim('}') || delim == '[' && end == json.Delim(']'))
}
