package httpapi

// createSessionRequest 是 POST /sessions 的请求体。
type createSessionRequest struct {
	Topic string `json:"topic"`
}

// sendMessageRequest 是 POST /sessions/{id}/messages 的请求体。
type sendMessageRequest struct {
	Content string `json:"content"`
}
