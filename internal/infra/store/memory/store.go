package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"insightforge/internal/domain/session"
)

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*session.Session
	events   map[string][]session.Event
}

// NewStore 创建第一阶段使用的内存 Store。
// TODO: 第二阶段替换为 SQLite/GORM 实现，接口行为保持不变。
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*session.Session),
		events:   make(map[string][]session.Event),
	}
}

func (s *Store) Create(ctx context.Context, topic string) (session.Session, error) {
	// 1. strings.TrimSpace(topic)
	// 2. 如果 topic 为空，使用“未命名研究任务”
	// 3. 创建 Session，ID 用 session.NewID("ses")
	// 4. Status 设置为 session.StatusCreated
	// 5. CreatedAt / UpdatedAt 设置为 time.Now()
	// 6. 加写锁，把 session 存进 s.sessions
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

func (s *Store) List(ctx context.Context) ([]session.Session, error) {
	// 1. 加读锁
	// 2. 把 map 里的 *Session 拷贝成 []Session
	// 3. 按 CreatedAt 倒序排序
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]session.Session, 0, len(s.sessions))
	for _, item := range s.sessions {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

func (s *Store) Get(ctx context.Context, sessionID string) (session.Session, error) {
	// 1. 加读锁
	// 2. 从 s.sessions 查 sessionID
	// 3. 不存在时返回 session.ErrNotFound
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.sessions[sessionID]
	if !ok {
		return session.Session{}, session.ErrNotFound
	}
	return *item, nil
}

func (s *Store) SetStatus(ctx context.Context, sessionID string, status session.Status) error {
	// 1. 加写锁
	// 2. 找到 session
	// 3. 更新 Status 和 UpdatedAt
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

func (s *Store) AddEvent(ctx context.Context, event session.Event) (session.Event, error) {
	// 1. 如果 event.ID 为空，用 session.NewID("evt")
	// 2. 如果 CreatedAt 为空，用 time.Now()
	// 3. 加写锁
	// 4. 确认 event.SessionID 对应的 Session 存在，否则 session.ErrNotFound
	// 5. append 到 s.events[event.SessionID]
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

func (s *Store) ListEvents(ctx context.Context, sessionID string) ([]session.Event, error) {
	// 1. 加读锁
	// 2. 确认 Session 存在
	// 3. 拷贝事件切片后返回，避免外部修改内部状态
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.sessions[sessionID]; !ok {
		return nil, session.ErrNotFound
	}
	items := s.events[sessionID]
	out := make([]session.Event, len(items))
	copy(out, items)
	return out, nil
}
