package session

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("session not found")

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	events   map[string][]Event
}

// NewStore 创建第一阶段使用的内存 Store。
// TODO: 第二阶段替换为 SQLite/GORM 实现，接口行为保持不变。
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
		events:   make(map[string][]Event),
	}
}

func (s *Store) Create(ctx context.Context, topic string) (Session, error) {
	// TODO:
	// 1. strings.TrimSpace(topic)
	// 2. 如果 topic 为空，使用“未命名研究任务”
	// 3. 创建 Session，ID 用 newID("ses")
	// 4. Status 设置为 StatusCreated
	// 5. CreatedAt / UpdatedAt 设置为 time.Now()
	// 6. 加写锁，把 session 存进 s.sessions
	_ = ctx
	_ = topic
	return Session{}, errors.New("TODO: implement Store.Create")
}

func (s *Store) List(ctx context.Context) ([]Session, error) {
	// TODO:
	// 1. 加读锁
	// 2. 把 map 里的 *Session 拷贝成 []Session
	// 3. 按 CreatedAt 倒序排序
	_ = ctx
	return nil, errors.New("TODO: implement Store.List")
}

func (s *Store) Get(ctx context.Context, sessionID string) (Session, error) {
	// TODO:
	// 1. 加读锁
	// 2. 从 s.sessions 查 sessionID
	// 3. 不存在时返回 ErrNotFound
	_ = ctx
	_ = sessionID
	return Session{}, errors.New("TODO: implement Store.Get")
}

func (s *Store) SetStatus(ctx context.Context, sessionID string, status Status) error {
	// TODO:
	// 1. 加写锁
	// 2. 找到 session
	// 3. 更新 Status 和 UpdatedAt
	_ = ctx
	_ = sessionID
	_ = status
	return errors.New("TODO: implement Store.SetStatus")
}

func (s *Store) AddEvent(ctx context.Context, event Event) (Event, error) {
	// TODO:
	// 1. 如果 event.ID 为空，用 newID("evt")
	// 2. 如果 CreatedAt 为空，用 time.Now()
	// 3. 加写锁
	// 4. 确认 event.SessionID 对应的 Session 存在，否则 ErrNotFound
	// 5. append 到 s.events[event.SessionID]
	_ = ctx
	_ = event
	return Event{}, errors.New("TODO: implement Store.AddEvent")
}

func (s *Store) ListEvents(ctx context.Context, sessionID string) ([]Event, error) {
	// TODO:
	// 1. 加读锁
	// 2. 确认 Session 存在
	// 3. 拷贝事件切片后返回，避免外部修改内部状态
	_ = ctx
	_ = sessionID
	return nil, errors.New("TODO: implement Store.ListEvents")
}
