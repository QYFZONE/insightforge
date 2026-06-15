package sse

import (
	"insightforge/internal/domain/session"
	"sync"
)

type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan session.Event]struct{}
}

// NewBroker 创建内存 SSE Broker。
// TODO: 多实例部署时需要替换为 Redis Pub/Sub 或消息队列。
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]map[chan session.Event]struct{}),
	}
}

// Subscribe 订阅某个 Session 的实时事件。
// 返回 cancel 用于请求结束时注销订阅，避免 channel 泄漏。
func (b *Broker) Subscribe(sessionID string) (<-chan session.Event, func()) {
	// 1. 创建带缓冲 channel
	// 2. 加写锁
	// 3. 如果 b.subscribers[sessionID] 不存在，先 make
	// 4. 把 channel 加进订阅者集合
	// 5. 返回只读 channel 和 cancel 函数
	ch := make(chan session.Event, 16)
	b.mu.Lock()

	if _, ok := b.subscribers[sessionID]; !ok {
		b.subscribers[sessionID] = make(map[chan session.Event]struct{})
	}
	b.subscribers[sessionID][ch] = struct{}{}

	b.mu.Unlock()

	// cancel 需要：
	// 1. 加写锁
	// 2. 从集合中 delete channel
	// 3. 如果集合为空，删除 sessionID
	// 4. close channel
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
	// 1. 加读锁
	// 2. 遍历 b.subscribers[event.SessionID]
	// 3. 非阻塞发送事件：
	//    select { case ch <- event: default: }
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers[event.SessionID] {
		select {
		case ch <- event:
		default:
		}
	}
}
