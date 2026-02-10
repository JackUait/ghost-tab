package util

import (
	"encoding/json"
	"fmt"
)

// OutputJSON marshals data to JSON and returns as string
func OutputJSON(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes), nil
}
