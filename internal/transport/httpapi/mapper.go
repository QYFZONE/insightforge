package httpapi

import "insightforge/internal/domain/session"

func toSessionResponse(item session.Session) sessionResponse {
	return sessionResponse{
		ID:        item.ID,
		Topic:     item.Topic,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func toSessionResponses(items []session.Session) []sessionResponse {
	out := make([]sessionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toSessionResponse(item))
	}
	return out
}

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
