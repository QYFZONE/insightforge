package httpapi

import "insightforge/internal/domain/session"

// toSessionResponse 把领域模型转换为 HTTP 响应 DTO。
func toSessionResponse(item session.Session) sessionResponse {
	return sessionResponse{
		ID:        item.ID,
		Topic:     item.Topic,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

// toSessionResponses 批量转换 Session 响应对象。
func toSessionResponses(items []session.Session) []sessionResponse {
	out := make([]sessionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toSessionResponse(item))
	}
	return out
}

// toEventResponse 把 timeline 事件转换为 HTTP/SSE 响应 DTO。
func toEventResponse(event session.Event) eventResponse {
	return eventResponse{
		ID:        event.ID,
		SessionID: event.SessionID,
		Type:      event.Type,
		Message:   event.Message,
		Payload:   event.Payload,
		CreatedAt: event.CreatedAt,
	}
}
