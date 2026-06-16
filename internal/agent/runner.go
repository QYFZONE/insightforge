package agent

import (
	"context"

	"insightforge/internal/domain/session"
)

// EventSink 是 Agent 向业务层写事件和状态的最小出口。
type EventSink interface {
	// Emit 会把事件写入历史，并推送给在线订阅者。
	Emit(ctx context.Context, event session.Event) error
	// EmitTransient 只推送实时事件，不写入历史，适合流式输出 delta。
	EmitTransient(ctx context.Context, event session.Event) error
	SetStatus(ctx context.Context, sessionID string, status session.Status) error
}

// RunInput 是一次 Agent 执行所需的业务输入。
type RunInput struct {
	SessionID string
	Topic     string
	UserText  string
	History   []session.Event
}

// Runner 抽象 Agent 执行器，让 mock、Ark、未来多 Agent 编排都能替换接入。
type Runner interface {
	Run(ctx context.Context, sink EventSink, input RunInput) error
}
