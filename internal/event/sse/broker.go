package sse

import (
	"insightforge/internal/domain/session"
	"sync"
)

// Broker 是单进程内存版 SSE 发布订阅中心。
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan session.Event]struct{}
}

// NewBroker 创建单进程内存 SSE Broker；多实例部署时可替换为 Redis Pub/Sub 或消息队列。
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]map[chan session.Event]struct{}),
	}
}

// Subscribe 订阅某个 Session 的实时事件。
// 返回 cancel 用于请求结束时注销订阅，避免 channel 泄漏。
func (b *Broker) Subscribe(sessionID string) (<-chan session.Event, func()) {
	// 带缓冲可以吸收短时间事件突增，避免发布端被单个客户端拖住。
	ch := make(chan session.Event, 16)
	b.mu.Lock()

	// 每个 sessionID 维护一组订阅者 channel。
	if _, ok := b.subscribers[sessionID]; !ok {
		b.subscribers[sessionID] = make(map[chan session.Event]struct{})
	}
	b.subscribers[sessionID][ch] = struct{}{}

	b.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()

			delete(b.subscribers[sessionID], ch)
			close(ch)
			if len(b.subscribers[sessionID]) == 0 {
				delete(b.subscribers, sessionID)
			}
		})
	}

	return ch, cancel
}

// Publish 会把事件推给当前在线的订阅者。
// 如果某个客户端太慢，当前实现会丢弃该客户端的这一条实时事件；
// 历史事件仍然保存在 Store 中，刷新页面可以补回来。
func (b *Broker) Publish(event session.Event) {
	// Publish 只负责实时推送；可靠历史由 Store 中的 ListEvents 保证。
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers[event.SessionID] {
		select {
		case ch <- event:
		default:
		}
	}
}
