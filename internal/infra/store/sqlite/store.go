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

type Store struct {
	db *gorm.DB
}

func (s *Store) Create(ctx context.Context, topic string) (session.Session, error) {
	// 1. 清理 topic，为空时使用默认标题。
	// 2. 组装 session.Session，ID 用 session.NewID("ses")。
	// 3. 用 toSessionModel 转成 GORM 数据对象。
	// 4. 调用 s.db.WithContext(ctx).Create(&record).Error。
	// 5. 返回领域对象 item。
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

func (s *Store) List(ctx context.Context) ([]session.Session, error) {
	// 1. 定义 []model.Session。
	// 2. 用 Order("created_at DESC").Find(&records) 查询。
	// 3. 把 records 转成 []session.Session。
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

func (s *Store) Get(ctx context.Context, sessionID string) (session.Session, error) {
	// 1. 用 First(&record, "id = ?", sessionID) 查询。
	// 2. 如果 errors.Is(err, gorm.ErrRecordNotFound)，返回 session.ErrNotFound。
	// 3. 其他错误原样返回。
	// 4. 成功时用 toSession(record) 返回领域对象。
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

func (s *Store) SetStatus(ctx context.Context, sessionID string, status session.Status) error {
	// 1. 用 Model(&model.Session{}).Where("id = ?", sessionID).Updates(...)。
	// 2. 同时更新 status 和 updated_at。
	// 3. 如果 RowsAffected == 0，返回 session.ErrNotFound。
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

func (s *Store) AddEvent(ctx context.Context, event session.Event) (session.Event, error) {
	// 1. 先调用 s.Get(ctx, event.SessionID)，确认 session 存在。
	// 2. event.ID 为空时用 session.NewID("evt")。
	// 3. event.CreatedAt 为空时用 time.Now()。
	// 4. 用 encodePayload 序列化 Payload。
	// 5. 用 toEventModel 转成 GORM 数据对象。
	// 6. 调用 s.db.WithContext(ctx).Create(&record).Error。
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

func (s *Store) ListEvents(ctx context.Context, sessionID string) ([]session.Event, error) {
	// 1. 先调用 s.Get(ctx, sessionID)，确认 session 存在。
	// 2. 用 Where("session_id = ?", sessionID).Order("created_at ASC").Find(&records)。
	// 3. 用 toEvent(record) 转成 []session.Event。
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
