package research

import (
	"context"
	"errors"
	"testing"
	"time"

	"insightforge/internal/agent"
	"insightforge/internal/domain/session"
)

// TestServiceSendMessageSavesUserMessage 验证 SendMessage 的成功路径。
func TestServiceSendMessageSavesUserMessage(t *testing.T) {
	svc, store, broker, runner := newTestService(t)
	store.createSessionForTest("ses_1", "测试主题")

	err := svc.SendMessage(context.Background(), "ses_1", " hello ")
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	events := store.events["ses_1"]
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].Type != "user_message" {
		t.Fatalf("event Type = %q, want %q", events[0].Type, "user_message")
	}
	if events[0].Message != "hello" {
		t.Fatalf("event Message = %q, want %q", events[0].Message, "hello")
	}

	if len(broker.published) != 1 {
		t.Fatalf("published len = %d, want 1", len(broker.published))
	}
	if broker.published[0].Message != "hello" {
		t.Fatalf("published Message = %q, want %q", broker.published[0].Message, "hello")
	}

	call := runner.receiveCall(t)
	if call.sessionID != "ses_1" {
		t.Fatalf("runner sessionID = %q, want %q", call.sessionID, "ses_1")
	}
	if call.topic != "测试主题" {
		t.Fatalf("runner topic = %q, want %q", call.topic, "测试主题")
	}
	if call.userText != "hello" {
		t.Fatalf("runner userText = %q, want %q", call.userText, "hello")
	}
	if len(call.history) != 1 {
		t.Fatalf("runner history len = %d, want 1", len(call.history))
	}
	if call.history[0].Message != "hello" {
		t.Fatalf("runner history[0].Message = %q, want %q", call.history[0].Message, "hello")
	}
}

// TestServiceSendMessageRejectsEmptyContent 验证空消息校验。
func TestServiceSendMessageRejectsEmptyContent(t *testing.T) {
	svc, store, broker, runner := newTestService(t)

	err := svc.SendMessage(context.Background(), "ses_1", "   ")
	requireErrorIs(t, err, ErrEmptyContent)

	if len(store.events) != 0 {
		t.Fatalf("events len = %d, want 0", len(store.events))
	}
	if len(broker.published) != 0 {
		t.Fatalf("published len = %d, want 0", len(broker.published))
	}
	runner.assertNoCall(t)
}

// TestServiceSendMessageReturnsNotFound 验证会话不存在时的错误路径。
func TestServiceSendMessageReturnsNotFound(t *testing.T) {
	svc, store, broker, runner := newTestService(t)

	err := svc.SendMessage(context.Background(), "missing", "hello")
	requireErrorIs(t, err, session.ErrNotFound)

	if len(store.events) != 0 {
		t.Fatalf("events len = %d, want 0", len(store.events))
	}
	if len(broker.published) != 0 {
		t.Fatalf("published len = %d, want 0", len(broker.published))
	}
	runner.assertNoCall(t)
}

// TestServiceEmitTransientPublishesWithoutPersisting 验证实时事件只推送、不落库。
func TestServiceEmitTransientPublishesWithoutPersisting(t *testing.T) {
	svc, store, broker, _ := newTestService(t)
	store.createSessionForTest("ses_1", "测试主题")

	err := svc.EmitTransient(context.Background(), session.Event{
		SessionID: "ses_1",
		Type:      "assistant_delta",
		Message:   "he",
	})
	if err != nil {
		t.Fatalf("EmitTransient() error = %v", err)
	}

	if len(store.events["ses_1"]) != 0 {
		t.Fatalf("events len = %d, want 0", len(store.events["ses_1"]))
	}
	if len(broker.published) != 1 {
		t.Fatalf("published len = %d, want 1", len(broker.published))
	}
	if broker.published[0].ID == "" {
		t.Fatal("published event ID should not be empty")
	}
	if broker.published[0].CreatedAt.IsZero() {
		t.Fatal("published event CreatedAt should not be zero")
	}
}

// newTestService 创建带 fake 依赖的 research.Service。
func newTestService(t *testing.T) (*Service, *fakeSessionStore, *fakeEventBroker, *fakeRunner) {
	t.Helper()

	store := newFakeSessionStore()
	broker := &fakeEventBroker{}
	runner := &fakeRunner{
		calls: make(chan runnerCall, 1),
	}

	svc, err := New(Config{
		Sessions: store,
		Events:   broker,
		Runner:   runner,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return svc, store, broker, runner
}

// fakeSessionStore 是测试用内存存储，只实现 Service 需要的最小行为。
type fakeSessionStore struct {
	sessions map[string]session.Session
	events   map[string][]session.Event
}

// newFakeSessionStore 创建空的测试存储。
func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{
		sessions: make(map[string]session.Session),
		events:   make(map[string][]session.Event),
	}
}

// createSessionForTest 直接塞入测试会话，避免测试绕远路依赖 Create。
func (s *fakeSessionStore) createSessionForTest(sessionID string, topic string) session.Session {
	item := session.Session{
		ID:     sessionID,
		Topic:  topic,
		Status: session.StatusCreated,
	}
	s.sessions[sessionID] = item
	return item
}

// Create 模拟真实 Store 的创建会话行为。
func (s *fakeSessionStore) Create(ctx context.Context, topic string) (session.Session, error) {
	return s.createSessionForTest("ses_test", topic), nil
}

// List 返回 fake store 里保存的所有会话。
func (s *fakeSessionStore) List(ctx context.Context) ([]session.Session, error) {
	items := make([]session.Session, 0, len(s.sessions))
	for _, item := range s.sessions {
		items = append(items, item)
	}
	return items, nil
}

// Get 模拟按 ID 查询会话。
func (s *fakeSessionStore) Get(ctx context.Context, sessionID string) (session.Session, error) {
	item, ok := s.sessions[sessionID]
	if !ok {
		return session.Session{}, session.ErrNotFound
	}
	return item, nil
}

// SetStatus 模拟会话状态更新。
func (s *fakeSessionStore) SetStatus(ctx context.Context, sessionID string, status session.Status) error {
	item, ok := s.sessions[sessionID]
	if !ok {
		return session.ErrNotFound
	}
	item.Status = status
	s.sessions[sessionID] = item
	return nil
}

// AddEvent 模拟事件落库。
func (s *fakeSessionStore) AddEvent(ctx context.Context, event session.Event) (session.Event, error) {
	if _, ok := s.sessions[event.SessionID]; !ok {
		return session.Event{}, session.ErrNotFound
	}
	s.events[event.SessionID] = append(s.events[event.SessionID], event)
	return event, nil
}

// ListEvents 返回事件副本，避免测试代码直接改内部切片。
func (s *fakeSessionStore) ListEvents(ctx context.Context, sessionID string) ([]session.Event, error) {
	if _, ok := s.sessions[sessionID]; !ok {
		return nil, session.ErrNotFound
	}
	out := make([]session.Event, len(s.events[sessionID]))
	copy(out, s.events[sessionID])
	return out, nil
}

// fakeEventBroker 记录 Publish 调用，方便测试断言是否推送事件。
type fakeEventBroker struct {
	published []session.Event
	ch        chan session.Event
}

// Publish 保存被发布的事件。
func (b *fakeEventBroker) Publish(event session.Event) {
	b.published = append(b.published, event)
}

// Subscribe 返回测试用 channel。
func (b *fakeEventBroker) Subscribe(sessionID string) (<-chan session.Event, func()) {
	b.ch = make(chan session.Event, 1)
	return b.ch, func() {
		close(b.ch)
	}
}

// fakeRunner 记录后台 Agent 是否被启动。
type fakeRunner struct {
	calls chan runnerCall
}

// runnerCall 保存 fakeRunner 收到的一次调用参数。
type runnerCall struct {
	sessionID string
	topic     string
	userText  string
	history   []session.Event
}

// Run 实现 agent.Runner，并把调用参数写入 channel。
func (r *fakeRunner) Run(ctx context.Context, sink agent.EventSink, input agent.RunInput) error {
	r.calls <- runnerCall{
		sessionID: input.SessionID,
		topic:     input.Topic,
		userText:  input.UserText,
		history:   input.History,
	}
	return nil
}

// receiveCall 等待 fakeRunner 被调用。
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

// assertNoCall 验证 fakeRunner 没有被启动。
func (r *fakeRunner) assertNoCall(t *testing.T) {
	t.Helper()

	select {
	case call := <-r.calls:
		t.Fatalf("不应该启动 runner，但启动了：%+v", call)
	case <-time.After(50 * time.Millisecond):
	}
}

// requireErrorIs 封装 errors.Is 断言，让测试主体更短。
func requireErrorIs(t *testing.T, got error, want error) {
	t.Helper()

	if !errors.Is(got, want) {
		t.Fatalf("error = %v, want errors.Is(..., %v)", got, want)
	}
}
