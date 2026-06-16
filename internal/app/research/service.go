package research

import (
	"context"
	"errors"
	"insightforge/internal/agent"
	"insightforge/internal/domain/session"
	"strings"
	"time"
)

// SessionStore 是研究服务依赖的会话存储端口。
type SessionStore interface {
	Create(ctx context.Context, topic string) (session.Session, error)
	List(ctx context.Context) ([]session.Session, error)
	Get(ctx context.Context, sessionID string) (session.Session, error)
	SetStatus(ctx context.Context, sessionID string, status session.Status) error
	AddEvent(ctx context.Context, event session.Event) (session.Event, error)
	ListEvents(ctx context.Context, sessionID string) ([]session.Event, error)
}

// EventBroker 是研究服务依赖的实时事件发布端口。
type EventBroker interface {
	Publish(event session.Event)
	Subscribe(sessionID string) (<-chan session.Event, func())
}

// Config 收集 research.Service 的全部依赖。
type Config struct {
	Sessions SessionStore
	Events   EventBroker
	Runner   agent.Runner
}

// Service 编排 Session、事件发布和 Agent Runner。
type Service struct {
	sessions SessionStore
	events   EventBroker
	runner   agent.Runner
}

// New 创建研究任务应用服务，并校验依赖是否完整。
func New(cfg Config) (*Service, error) {
	if cfg.Sessions == nil {
		return nil, errors.New("research: session store is required")
	}
	if cfg.Events == nil {
		return nil, errors.New("research: event broker is required")
	}
	if cfg.Runner == nil {
		return nil, errors.New("research: runner is required")
	}

	return &Service{
		sessions: cfg.Sessions,
		events:   cfg.Events,
		runner:   cfg.Runner,
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
	content = strings.TrimSpace(content)
	if content == "" {
		return ErrEmptyContent
	}

	// 先确认会话存在，并拿到 topic 交给 Agent 构造上下文。
	item, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// 用户消息先落库再启动 Agent，保证后续历史上下文包含本轮输入。
	err = s.Emit(ctx, session.Event{
		SessionID: sessionID,
		Type:      "user_message",
		Message:   content,
	})

	if err != nil {
		return err
	}

	// 读取包含当前 user_message 在内的历史事件，交给 Runner 组装模型上下文。
	// 后续做上下文压缩时，也可以先在这里裁剪历史。
	history, err := s.sessions.ListEvents(ctx, sessionID)
	if err != nil {
		return err
	}
	go s.runAgent(context.Background(), agent.RunInput{
		SessionID: sessionID,
		Topic:     item.Topic,
		UserText:  content,
		History:   history,
	})

	return nil
}

// runAgent 在后台执行 Agent，并把失败统一转换为状态和错误事件。
func (s *Service) runAgent(ctx context.Context, input agent.RunInput) {
	if err := s.runner.Run(ctx, s, input); err != nil {
		// 后台任务不能把错误直接返回给 HTTP 请求，只能落库并推送到 timeline。
		_ = s.sessions.SetStatus(ctx, input.SessionID, session.StatusFailed)
		_ = s.Emit(ctx, session.Event{
			SessionID: input.SessionID,
			Type:      "error",
			Message:   "Agent 执行失败",
			Payload: map[string]any{
				"error": err.Error(),
			},
		})
	}
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

// EmitTransient 只把事件推给在线订阅者，不写入历史。
//
// 适合 assistant_delta 这类流式片段：前端需要实时看到，但历史里只需要最终 assistant_message。
func (s *Service) EmitTransient(ctx context.Context, event session.Event) error {
	// 先确认会话存在，避免向不存在的 session 推送孤立事件。
	if _, err := s.sessions.Get(ctx, event.SessionID); err != nil {
		return err
	}

	// transient 事件不落库，只补齐实时推送所需的基础元信息。
	event = fillRealtimeEventMeta(event)
	s.events.Publish(event)
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

// fillRealtimeEventMeta 为不落库的实时事件补齐前端需要的基础字段。
func fillRealtimeEventMeta(event session.Event) session.Event {
	if event.ID == "" {
		event.ID = session.NewID("evt")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	return event
}
