package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ExtractJSON extracts JSON from a string that may contain other text
// This is useful when LLMs return JSON wrapped in markdown code blocks
func ExtractJSON(input string) string {
	// Try to find JSON in markdown code blocks first
	codeBlockRegex := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	matches := codeBlockRegex.FindStringSubmatch(input)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find raw JSON object
	jsonRegex := regexp.MustCompile(`\{[\s\S]*\}`)
	match := jsonRegex.FindString(input)
	if match != "" {
		return match
	}

	// Return original input if no JSON found
	return strings.TrimSpace(input)
}

// ParseJSON parses JSON string into a map
func ParseJSON(input string) (map[string]interface{}, error) {
	cleaned := ExtractJSON(input)
	var result map[string]interface{}
	err := json.Unmarshal([]byte(cleaned), &result)
	return result, err
}

// ToJSON converts an interface to JSON string
func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToPrettyJSON converts an interface to formatted JSON string
func ToPrettyJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetString safely gets a string value from a map
func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt safely gets an int value from a map
func GetInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case int64:
			return int(val)
		}
	}
	return 0
}

// GetFloat safely gets a float64 value from a map
func GetFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

// GetBool safely gets a bool value from a map
func GetBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetMap safely gets a nested map from a map
func GetMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if nested, ok := v.(map[string]interface{}); ok {
			return nested
		}
	}
	return nil
}
