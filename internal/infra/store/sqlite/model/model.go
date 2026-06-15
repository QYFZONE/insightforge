package model

import "time"

type Session struct {
	ID        string    `gorm:"primaryKey;column:id"`
	Topic     string    `gorm:"column:topic;not null"`
	Status    string    `gorm:"column:status;not null"`
	CreatedAt time.Time `gorm:"index:idx_sessions_created_at"`
	UpdatedAt time.Time
}

func (Session) TableName() string {
	return "sessions"
}

type Event struct {
	ID        string    `gorm:"primaryKey;column:id"`
	SessionID string    `gorm:"column:session_id;not null;index:idx_events_session_created_at,priority:1"`
	Type      string    `gorm:"column:type;not null"`
	Message   string    `gorm:"column:message;not null"`
	Payload   string    `gorm:"column:payload;type:TEXT"`
	CreatedAt time.Time `gorm:"column:created_at;not null;index:idx_events_session_created_at,priority:2"`
}

func (Event) TableName() string {
	return "events"
}
