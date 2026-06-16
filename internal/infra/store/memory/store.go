package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"insightforge/internal/domain/session"
)

// Store 是内存版 SessionStore，主要用于本地开发和快速演示。
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*session.Session
	events   map[string][]session.Event
}

// NewStore 创建内存 Store；它和持久化 Store 暴露同一套接口，方便本地调试切换。
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*session.Session),
		events:   make(map[string][]session.Event),
	}
}

// Create 创建研究会话，并把它保存到内存 map。
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = &item
	return item, nil
}

// List 返回所有会话，并按创建时间倒序排列。
func (s *Store) List(ctx context.Context) ([]session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回值使用值拷贝，避免调用方修改 Store 内部状态。
	items := make([]session.Session, 0, len(s.sessions))
	for _, item := range s.sessions {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

// Get 根据 ID 读取会话。
func (s *Store) Get(ctx context.Context, sessionID string) (session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.sessions[sessionID]
	if !ok {
		return session.Session{}, session.ErrNotFound
	}
	return *item, nil
}

// SetStatus 更新会话状态和更新时间。
func (s *Store) SetStatus(ctx context.Context, sessionID string, status session.Status) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.sessions[sessionID]
	if !ok {
		return session.ErrNotFound
	}
	item.Status = status
	item.UpdatedAt = time.Now()
	return nil
}

// AddEvent 保存 timeline 事件。
func (s *Store) AddEvent(ctx context.Context, event session.Event) (session.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID == "" {
		event.ID = session.NewID("evt")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if _, ok := s.sessions[event.SessionID]; !ok {
		return session.Event{}, session.ErrNotFound
	}
	s.events[event.SessionID] = append(s.events[event.SessionID], event)
	return event, nil
}

// ListEvents 返回某个会话的历史事件副本。
func (s *Store) ListEvents(ctx context.Context, sessionID string) ([]session.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.sessions[sessionID]; !ok {
		return nil, session.ErrNotFound
	}

	// 拷贝事件切片，避免调用方拿到内部 slice 后意外改写 Store 状态。
	items := s.events[sessionID]
	out := make([]session.Event, len(items))
	copy(out, items)
	return out, nil
}
