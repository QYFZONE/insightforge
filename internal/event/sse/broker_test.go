package sse

import (
	"testing"
	"time"

	"insightforge/internal/domain/session"
)

// TestBrokerPublishToSameSession 验证事件只会发给同一个 session 的订阅者。
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

// TestBrokerDoesNotCrossSession 验证不同 session 之间的事件隔离。
func TestBrokerDoesNotCrossSession(t *testing.T) {
	broker := NewBroker()

	ch, cancel := broker.Subscribe("ses_1")
	defer cancel()

	broker.Publish(session.Event{
		SessionID: "ses_2",
		Type:      "user_message",
		Message:   "hello from ses_2",
	})

	assertNoEvent(t, ch)
}

// TestBrokerCancelClosesSubscription 验证订阅取消后的 channel 关闭行为。
func TestBrokerCancelClosesSubscription(t *testing.T) {
	broker := NewBroker()

	ch, cancel := broker.Subscribe("ses_1")
	cancel()
	cancel()

	_, ok := <-ch
	if ok {
		t.Fatal("cancel 后订阅 channel 应该被关闭")
	}
}

// receiveEvent 等待一条事件，超时则让测试失败。
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

// assertNoEvent 验证指定时间内没有事件到达。
func assertNoEvent(t *testing.T, ch <-chan session.Event) {
	t.Helper()

	select {
	case event := <-ch:
		t.Fatalf("不应该收到事件，但收到了：%+v", event)
	case <-time.After(50 * time.Millisecond):
	}
}
