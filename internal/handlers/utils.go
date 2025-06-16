package handlers

import (
	"encoding/json"
	"fmt"
)

// formatJSONResponse converts a response struct to a formatted JSON string
func formatJSONResponse(response interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}
	
	return string(jsonBytes), nil
}