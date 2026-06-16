package sqlite

import (
	"context"
	"errors"
	"insightforge/internal/infra/store/sqlite/model"
	"strings"
	"time"

	"insightforge/internal/domain/session"

	"gorm.io/gorm"
)

// Store 是基于 GORM 的 SQLite 持久化实现。
type Store struct {
	db *gorm.DB
}

// Create 创建研究会话并持久化到 sessions 表。
func (s *Store) Create(ctx context.Context, topic string) (session.Session, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		topic = "未命名研究任务"
	}
	sessionID := session.NewID("ses")
	nowTime := time.Now()
	item := session.Session{
		ID:        sessionID,
		Topic:     topic,
		Status:    session.StatusCreated,
		CreatedAt: nowTime,
		UpdatedAt: nowTime,
	}
	record := toSessionModel(item)
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return session.Session{}, err
	}
	return item, nil
}

// List 返回所有会话，并按创建时间倒序排列。
func (s *Store) List(ctx context.Context) ([]session.Session, error) {
	var records []model.Session
	err := s.db.WithContext(ctx).Order("created_at desc").Find(&records).Error
	if err != nil {
		return nil, err
	}
	sessions := make([]session.Session, 0, len(records))
	for _, record := range records {
		sessions = append(sessions, toSession(record))
	}
	return sessions, nil
}

// Get 根据 ID 查询会话，并把 GORM 的 not found 映射为领域错误。
func (s *Store) Get(ctx context.Context, sessionID string) (session.Session, error) {
	record := model.Session{}
	err := s.db.WithContext(ctx).First(&record, "id = ?", sessionID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return session.Session{}, session.ErrNotFound
	}
	if err != nil {
		return session.Session{}, err
	}
	return toSession(record), nil
}

// SetStatus 更新会话状态。
func (s *Store) SetStatus(ctx context.Context, sessionID string, status session.Status) error {
	result := s.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("id = ?", sessionID).
		Updates(map[string]any{
			"status":     string(status),
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return session.ErrNotFound
	}
	return nil
}

// AddEvent 保存 timeline 事件。
func (s *Store) AddEvent(ctx context.Context, event session.Event) (session.Event, error) {
	if _, err := s.Get(ctx, event.SessionID); err != nil {
		return session.Event{}, err
	}
	if event.ID == "" {
		event.ID = session.NewID("evt")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	payload, err := encodePayload(event.Payload)
	if err != nil {
		return session.Event{}, err
	}
	record := toEventModel(event, payload)
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return session.Event{}, err
	}
	return event, nil
}

// ListEvents 返回某个会话的历史事件。
func (s *Store) ListEvents(ctx context.Context, sessionID string) ([]session.Event, error) {
	if _, err := s.Get(ctx, sessionID); err != nil {
		return nil, err
	}

	var records []model.Event
	if err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}

	events := make([]session.Event, 0, len(records))
	for _, record := range records {
		event, err := toEvent(record)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
