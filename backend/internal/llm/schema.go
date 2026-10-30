package llm

import (
	"bytes"
	"encoding/json"
	"math"
	"reflect"
	"unicode/utf8"
)

func obj(properties map[string]any) map[string]any {
	required := make([]string, 0, len(properties))
	for key := range properties {
		required = append(required, key)
	}
	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}
func str() map[string]any            { return map[string]any{"type": "string"} }
func nullableString() map[string]any { return map[string]any{"type": []string{"string", "null"}} }
func enumeration(values ...string) map[string]any {
	return map[string]any{"type": "string", "enum": values}
}
func integer(min, max int) map[string]any {
	return map[string]any{"type": "integer", "minimum": min, "maximum": max}
}
func boolean() map[string]any { return map[string]any{"type": "boolean"} }
func array(items map[string]any) map[string]any {
	return map[string]any{"type": "array", "items": items}
}
func nullable(schema map[string]any) map[string]any {
	return map[string]any{"anyOf": []any{schema, map[string]any{"type": "null"}}}
}

// Check the same schema locally, including required/null fields, then decode.
func decodeStrict(data []byte, schema map[string]any, target any) error {
	if !uniqueKeys(json.NewDecoder(bytes.NewReader(data))) {
		return ErrInvalidResponse
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return ErrInvalidResponse
	}
	// Normalize schema maps/slices and numeric bounds to their JSON representation.
	encoded, err := json.Marshal(schema)
	if err != nil {
		return ErrInvalidResponse
	}
	var normalized map[string]any
	if json.Unmarshal(encoded, &normalized) != nil || !validSchema(value, normalized) {
		return ErrInvalidResponse
	}
	if err := json.Unmarshal(data, target); err != nil {
		return ErrInvalidResponse
	}
	return nil
}

func validSchema(value any, schema map[string]any) bool {
	if alternatives, ok := schema["anyOf"].([]any); ok {
		for _, alternative := range alternatives {
			if validSchema(value, alternative.(map[string]any)) {
				return true
			}
		}
		return false
	}
	if types, ok := schema["type"].([]any); ok {
		for _, typ := range types {
			copy := map[string]any{}
			for k, v := range schema {
				copy[k] = v
			}
			copy["type"] = typ
			if validSchema(value, copy) {
				return true
			}
		}
		return false
	}
	if choices, ok := schema["enum"].([]any); ok {
		found := false
		for _, choice := range choices {
			if reflect.DeepEqual(value, choice) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	switch schema["type"] {
	case "null":
		return value == nil
	case "string":
		s, ok := value.(string)
		if !ok {
			return false
		}
		if max, ok := schema["maxLength"].(float64); ok && float64(utf8.RuneCountInString(s)) > max {
			return false
		}
		return true
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "integer", "number":
		n, ok := value.(float64)
		if !ok {
			return false
		}
		if schema["type"] == "integer" && n != math.Trunc(n) {
			return false
		}
		if min, ok := schema["minimum"].(float64); ok && n < min {
			return false
		}
		if max, ok := schema["maximum"].(float64); ok && n > max {
			return false
		}
		return true
	case "array":
		items, ok := value.([]any)
		if !ok {
			return false
		}
		if max, ok := schema["maxItems"].(float64); ok && float64(len(items)) > max {
			return false
		}
		for _, item := range items {
			if !validSchema(item, schema["items"].(map[string]any)) {
				return false
			}
		}
		return true
	case "object":
		fields, ok := value.(map[string]any)
		if !ok {
			return false
		}
		properties := schema["properties"].(map[string]any)
		for _, key := range schema["required"].([]any) {
			if _, ok := fields[key.(string)]; !ok {
				return false
			}
		}
		for key, v := range fields {
			property, ok := properties[key]
			if !ok {
				return false
			}
			if !validSchema(v, property.(map[string]any)) {
				return false
			}
		}
		return true
	}
	return false
}
