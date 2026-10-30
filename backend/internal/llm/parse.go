package llm

import "encoding/json"

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
