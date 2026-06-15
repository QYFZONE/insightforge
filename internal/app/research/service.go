package research

import (
	"context"
	"errors"
	"insightforge/internal/agent"
	"insightforge/internal/domain/session"
	"strings"
)

type SessionStore interface {
	Create(ctx context.Context, topic string) (session.Session, error)
	List(ctx context.Context) ([]session.Session, error)
	Get(ctx context.Context, sessionID string) (session.Session, error)
	SetStatus(ctx context.Context, sessionID string, status session.Status) error
	AddEvent(ctx context.Context, event session.Event) (session.Event, error)
	ListEvents(ctx context.Context, sessionID string) ([]session.Event, error)
}

type EventBroker interface {
	Publish(event session.Event)
	Subscribe(sessionID string) (<-chan session.Event, func())
}

type Runner func(ctx context.Context, sink agent.EventSink, sessionID string, topic string)

type Config struct {
	Sessions    SessionStore
	Events      EventBroker
	RunResearch Runner
}

type Service struct {
	sessions    SessionStore
	events      EventBroker
	runResearch Runner
}

func New(cfg Config) (*Service, error) {
	if cfg.Sessions == nil {
		return nil, errors.New("research: session store is required")
	}
	if cfg.Events == nil {
		return nil, errors.New("research: event broker is required")
	}
	if cfg.RunResearch == nil {
		return nil, errors.New("research: runner is required")
	}

	return &Service{
		sessions:    cfg.Sessions,
		events:      cfg.Events,
		runResearch: cfg.RunResearch,
	}, nil
}

// CreateSession 创建一个研究会话。
//
// 可能返回：
//   - 底层 SessionStore 返回的错误。
func (s *Service) CreateSession(ctx context.Context, topic string) (session.Session, error) {
	// 业务层创建 Session，不要让 HTTP handler 直接碰 Store。
	return s.sessions.Create(ctx, topic)
}

// ListSessions 返回所有研究会话。
//
// 可能返回：
//   - 底层 SessionStore 返回的错误。
func (s *Service) ListSessions(ctx context.Context) ([]session.Session, error) {
	// 调用 s.sessions.List(ctx)，后面可以在这里加分页、权限、过滤。
	return s.sessions.List(ctx)
}

// SendMessage 保存用户消息，并启动研究任务执行器。
//
// 可能返回：
//   - ErrEmptyContent：content 去掉空白字符后为空。
//   - session.ErrNotFound：sessionID 对应的会话不存在。
//   - Emit 返回的错误。
func (s *Service) SendMessage(ctx context.Context, sessionID string, content string) error {
	// 1. 用 s.sessions.Get(ctx, sessionID) 确认 Session 存在，并拿到 topic。
	// 2. strings.TrimSpace(content)，为空时返回业务错误。
	// 3. 调用 s.Emit 保存并发布 user_message 事件。
	// 4. 用 goroutine 启动 s.runResearch(context.Background(), s, sessionID, item.Topic)。
	content = strings.TrimSpace(content)
	if content == "" {
		return ErrEmptyContent
	}

	item, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	err = s.Emit(ctx, session.Event{
		SessionID: sessionID,
		Type:      "user_message",
		Message:   content,
	})

	if err != nil {
		return err
	}
	go s.runResearch(context.Background(), s, sessionID, item.Topic)

	return nil
}

// ListEvents 返回某个会话已经保存的时间线事件。
//
// 可能返回：
//   - session.ErrNotFound：sessionID 对应的会话不存在。
//   - 底层 SessionStore 返回的错误。
func (s *Service) ListEvents(ctx context.Context, sessionID string) ([]session.Event, error) {
	// 调用 s.sessions.ListEvents(ctx, sessionID)。
	return s.sessions.ListEvents(ctx, sessionID)
}

// SubscribeEvents 订阅某个会话的实时事件。
func (s *Service) SubscribeEvents(sessionID string) (<-chan session.Event, func()) {
	return s.events.Subscribe(sessionID)
}

// Emit 保存事件，并把它推送给在线订阅者。
//
// 可能返回：
//   - session.ErrNotFound：event.SessionID 对应的会话不存在。
//   - 底层 SessionStore 返回的错误。
func (s *Service) Emit(ctx context.Context, event session.Event) error {
	saved, err := s.sessions.AddEvent(ctx, event)
	if err != nil {
		return err
	}
	s.events.Publish(saved)
	return nil
}

// SetStatus 更新会话状态。
//
// 可能返回：
//   - session.ErrNotFound：sessionID 对应的会话不存在。
//   - 底层 SessionStore 返回的错误。
func (s *Service) SetStatus(ctx context.Context, sessionID string, status session.Status) error {
	return s.sessions.SetStatus(ctx, sessionID, status)
}
