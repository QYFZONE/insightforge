package sse

import (
	"testing"
	"time"

	"insightforge/internal/domain/session"
)

func TestBrokerPublishToSameSession(t *testing.T) {
	broker := NewBroker()

	ch, cancel := broker.Subscribe("ses_1")
	defer cancel()

	broker.Publish(session.Event{
		SessionID: "ses_1",
		Type:      "user_message",
		Message:   "hello",
	})

	event := receiveEvent(t, ch)
	if event.Message != "hello" {
		t.Fatalf("Message = %q, want %q", event.Message, "hello")
	}
}

func TestBrokerDoesNotCrossSession(t *testing.T) {
	// TODO:
	// 1. 订阅 ses_1。
	// 2. 发布一条 SessionID 为 ses_2 的事件。
	// 3. 用 assertNoEvent 验证 ses_1 没有收到事件。
}

func TestBrokerCancelClosesSubscription(t *testing.T) {
	// TODO:
	// 1. 订阅 ses_1。
	// 2. 调用 cancel()。
	// 3. 从 ch 读取，验证 ok == false。
}

func receiveEvent(t *testing.T, ch <-chan session.Event) session.Event {
	t.Helper()

	select {
	case event := <-ch:
		return event
	case <-time.After(time.Second):
		t.Fatal("等待事件超时")
		return session.Event{}
	}
}

func assertNoEvent(t *testing.T, ch <-chan session.Event) {
	t.Helper()

	select {
	case event := <-ch:
		t.Fatalf("不应该收到事件，但收到了：%+v", event)
	case <-time.After(50 * time.Millisecond):
	}
}
