package httpapi

import (
	"time"

	"insightforge/internal/domain/session"
)

// sessionResponse 是对外暴露的 Session 响应结构。
type sessionResponse struct {
	ID        string         `json:"id"`
	Topic     string         `json:"topic"`
	Status    session.Status `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// listSessionsResponse 包装会话列表，便于后续追加分页字段。
type listSessionsResponse struct {
	Sessions []sessionResponse `json:"sessions"`
}

// sendMessageResponse 表示消息已被接受并开始后台处理。
type sendMessageResponse struct {
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
}

// errorResponse 是所有 HTTP 错误响应的统一格式。
type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// eventResponse 是 SSE 和历史事件接口复用的事件 DTO。
type eventResponse struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
