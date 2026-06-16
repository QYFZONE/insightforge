package model

import "time"

// Session 是 sessions 表的 GORM 数据模型。
type Session struct {
	ID        string    `gorm:"primaryKey;column:id"`
	Topic     string    `gorm:"column:topic;not null"`
	Status    string    `gorm:"column:status;not null"`
	CreatedAt time.Time `gorm:"index:idx_sessions_created_at"`
	UpdatedAt time.Time
}

// TableName 固定 Session 模型对应的表名。
func (Session) TableName() string {
	return "sessions"
}

// Event 是 events 表的 GORM 数据模型。
type Event struct {
	ID        string    `gorm:"primaryKey;column:id"`
	SessionID string    `gorm:"column:session_id;not null;index:idx_events_session_created_at,priority:1"`
	Type      string    `gorm:"column:type;not null"`
	Message   string    `gorm:"column:message;not null"`
	Payload   string    `gorm:"column:payload;type:TEXT"`
	CreatedAt time.Time `gorm:"column:created_at;not null;index:idx_events_session_created_at,priority:2"`
}

// TableName 固定 Event 模型对应的表名。
func (Event) TableName() string {
	return "events"
}
