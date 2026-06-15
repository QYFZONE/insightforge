package httpapi

import (
	"time"

	"insightforge/internal/domain/session"
)

type sessionResponse struct {
	ID        string         `json:"id"`
	Topic     string         `json:"topic"`
	Status    session.Status `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type listSessionsResponse struct {
	Sessions []sessionResponse `json:"sessions"`
}

type sendMessageResponse struct {
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type eventResponse struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
