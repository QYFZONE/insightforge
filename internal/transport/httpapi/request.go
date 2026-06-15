package httpapi

type createSessionRequest struct {
	Topic string `json:"topic"`
}

type sendMessageRequest struct {
	Content string `json:"content"`
}
