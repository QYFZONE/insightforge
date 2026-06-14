package session

import "time"

type Status string

const (
	StatusCreated         Status = "created"
	StatusRunning         Status = "running"
	StatusWaitingApproval Status = "waiting_approval"
	StatusCompleted       Status = "completed"
	StatusFailed          Status = "failed"
)

// Session 是一次研究任务的顶层记录。
// TODO: 第二阶段把它映射到 SQLite sessions 表。
type Session struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Event 是前端 timeline 展示的最小事件单元。
// 它既会写入历史，也会通过 SSE 实时推送。
type Event struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// CreateRequest 对应 POST /sessions。
type CreateRequest struct {
	Topic string `json:"topic"`
}

// MessageRequest 对应 POST /sessions/{id}/messages。
type MessageRequest struct {
	Content string `json:"content"`
}
