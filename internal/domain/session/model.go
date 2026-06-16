package session

import "time"

// Status 描述研究任务当前所处的生命周期状态。
type Status string

const (
	// 任务已经创建，还没有开始执行
	StatusCreated Status = "created"
	// 任务正在执行
	StatusRunning Status = "running"
	// 等待用户确认或审批
	StatusWaitingApproval Status = "waiting_approval"
	// 任务执行完成
	StatusCompleted Status = "completed"
	// 任务执行失败
	StatusFailed Status = "failed"
)

// Session 是一次研究任务的顶层记录。
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
