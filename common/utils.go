package common

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ExtractSchemaField(jsonStr string) (string, bool, error) {
	// Unmarshal the JSON string into a generic map
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		return "", false, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if the "$schema" field exists
	if schema, ok := jsonData["$schema"].(string); ok {
		return schema, true, nil
	}

	// Return an empty string if "$schema" is not found
	return "", false, nil
}

// Helper function to convert snake_case to CamelCase
func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		result += CapitalizeFirst(part)
	}
	return result
}

// Helper function to capitalize first letter
func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
