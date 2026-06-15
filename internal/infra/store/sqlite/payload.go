package sqlite

import (
	"encoding/json"
	"strings"
)

func encodePayload(payload map[string]any) (string, error) {
	if len(payload) == 0 {
		return "", nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodePayload(value string) (map[string]any, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
