package sqlite

import (
	"encoding/json"
	"strings"
)

// encodePayload 把事件 payload 序列化为 JSON 字符串。
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

// decodePayload 把数据库中的 JSON 字符串还原为 payload。
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
