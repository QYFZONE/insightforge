package research

import (
	"context"
	"errors"
	"testing"
	"time"

	"insightforge/internal/agent"
	"insightforge/internal/domain/session"
)

func TestServiceSendMessageSavesUserMessage(t *testing.T) {
	// TODO:
	// 1. 用 newTestService 创建 service 和 fake 依赖。
	// 2. 调用 store.createSessionForTest 创建一个 session。
	// 3. 调用 svc.SendMessage(ctx, sessionID, " hello ")。
	// 4. 验证 store.events 里保存了一条 user_message。
	// 5. 验证 broker.published 里发布了一条 user_message。
	// 6. 验证 runner 被启动了一次。
}

func TestServiceSendMessageRejectsEmptyContent(t *testing.T) {
	// TODO:
	// 1. 创建 service。
	// 2. 调用 svc.SendMessage(ctx, "ses_1", "   ")。
	// 3. 用 errors.Is(err, ErrEmptyContent) 验证错误。
}

func TestServiceSendMessageReturnsNotFound(t *testing.T) {
	// TODO:
	// 1. 创建 service，但不要创建 session。
	// 2. 调用 svc.SendMessage(ctx, "missing", "hello")。
	// 3. 用 errors.Is(err, session.ErrNotFound) 验证错误。
}

func newTestService(t *testing.T) (*Service, *fakeSessionStore, *fakeEventBroker, *fakeRunner) {
	t.Helper()

	store := newFakeSessionStore()
	broker := &fakeEventBroker{}
	runner := &fakeRunner{
		calls: make(chan runnerCall, 1),
	}

	svc, err := New(Config{
		Sessions:    store,
		Events:      broker,
		RunResearch: runner.run,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return svc, store, broker, runner
}

type fakeSessionStore struct {
	sessions map[string]session.Session
	events   map[string][]session.Event
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{
		sessions: make(map[string]session.Session),
		events:   make(map[string][]session.Event),
	}
}

func (s *fakeSessionStore) createSessionForTest(sessionID string, topic string) session.Session {
	item := session.Session{
		ID:     sessionID,
		Topic:  topic,
		Status: session.StatusCreated,
	}
	s.sessions[sessionID] = item
	return item
}

func (s *fakeSessionStore) Create(ctx context.Context, topic string) (session.Session, error) {
	return s.createSessionForTest("ses_test", topic), nil
}

func (s *fakeSessionStore) List(ctx context.Context) ([]session.Session, error) {
	items := make([]session.Session, 0, len(s.sessions))
	for _, item := range s.sessions {
		items = append(items, item)
	}
	return items, nil
}

func (s *fakeSessionStore) Get(ctx context.Context, sessionID string) (session.Session, error) {
	item, ok := s.sessions[sessionID]
	if !ok {
		return session.Session{}, session.ErrNotFound
	}
	return item, nil
}

func (s *fakeSessionStore) SetStatus(ctx context.Context, sessionID string, status session.Status) error {
	item, ok := s.sessions[sessionID]
	if !ok {
		return session.ErrNotFound
	}
	item.Status = status
	s.sessions[sessionID] = item
	return nil
}

func (s *fakeSessionStore) AddEvent(ctx context.Context, event session.Event) (session.Event, error) {
	if _, ok := s.sessions[event.SessionID]; !ok {
		return session.Event{}, session.ErrNotFound
	}
	s.events[event.SessionID] = append(s.events[event.SessionID], event)
	return event, nil
}

func (s *fakeSessionStore) ListEvents(ctx context.Context, sessionID string) ([]session.Event, error) {
	if _, ok := s.sessions[sessionID]; !ok {
		return nil, session.ErrNotFound
	}
	out := make([]session.Event, len(s.events[sessionID]))
	copy(out, s.events[sessionID])
	return out, nil
}

type fakeEventBroker struct {
	published []session.Event
	ch        chan session.Event
}

func (b *fakeEventBroker) Publish(event session.Event) {
	b.published = append(b.published, event)
}

func (b *fakeEventBroker) Subscribe(sessionID string) (<-chan session.Event, func()) {
	b.ch = make(chan session.Event, 1)
	return b.ch, func() {
		close(b.ch)
	}
}

type fakeRunner struct {
	calls chan runnerCall
}

type runnerCall struct {
	sessionID string
	topic     string
}

func (r *fakeRunner) run(ctx context.Context, sink agent.EventSink, sessionID string, topic string) {
	r.calls <- runnerCall{
		sessionID: sessionID,
		topic:     topic,
	}
}

func (r *fakeRunner) receiveCall(t *testing.T) runnerCall {
	t.Helper()

	select {
	case call := <-r.calls:
		return call
	case <-time.After(time.Second):
		t.Fatal("等待 runner 启动超时")
		return runnerCall{}
	}
}

func (r *fakeRunner) assertNoCall(t *testing.T) {
	t.Helper()

	select {
	case call := <-r.calls:
		t.Fatalf("不应该启动 runner，但启动了：%+v", call)
	case <-time.After(50 * time.Millisecond):
	}
}

func requireErrorIs(t *testing.T, got error, want error) {
	t.Helper()

	if !errors.Is(got, want) {
		t.Fatalf("error = %v, want errors.Is(..., %v)", got, want)
	}
}
