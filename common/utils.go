package common

import (
	"encoding/json"
	"fmt"
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
