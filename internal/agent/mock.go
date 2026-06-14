package agent

import (
	"context"

	"insightforge/internal/session"
)

type EventSink interface {
	Emit(ctx context.Context, event session.Event) error
	SetStatus(ctx context.Context, sessionID string, status session.Status) error
}

// RunMockResearch 是第一阶段的假 Agent timeline。
// TODO: 第二阶段用 Eino Runner 替换这里，事件来源改为 Callback + Tool/Agent 输出。
func RunMockResearch(ctx context.Context, store EventSink, sessionID string, topic string) {
	// TODO:
	// 1. store.SetStatus(ctx, sessionID, session.StatusRunning)
	// 2. 构造 []session.Event，依次包含：
	//    - session_started
	//    - agent_started
	//    - workflow_step
	//    - tool_call
	//    - tool_result
	//    - approval_required
	// 3. 每个事件之间 time.Sleep 一小段，模拟 Agent 执行过程
	// 4. 每个事件调用 store.Emit(ctx, event)
	// 5. 最后把状态设为 session.StatusWaitingApproval
	_ = ctx
	_ = store
	_ = sessionID
	_ = topic
}
