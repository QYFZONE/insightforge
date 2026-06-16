package agent

import (
	"context"
	"time"

	"insightforge/internal/domain/session"
)

// MockRunner 用固定事件模拟真实 Agent 执行过程。
type MockRunner struct {
	Delay time.Duration
}

// NewMockRunner 创建本地演示用 Runner。
func NewMockRunner() *MockRunner {
	return &MockRunner{Delay: 350 * time.Millisecond}
}

// Run 是本地演示用的假 Agent timeline；真实模型链路由 ArkRunner 提供。
func (r *MockRunner) Run(ctx context.Context, sink EventSink, input RunInput) error {
	// mock 也走真实状态机，方便前端和存储逻辑在早期就被验证。
	if err := sink.SetStatus(ctx, input.SessionID, session.StatusRunning); err != nil {
		return err
	}

	// steps 用一组固定事件模拟 Planner/Researcher/Approval 的执行轨迹。
	steps := []session.Event{
		{
			SessionID: input.SessionID,
			Type:      "session_started",
			Message:   "研究任务已启动",
			Payload: map[string]any{
				"topic":     input.Topic,
				"user_text": input.UserText,
			},
		},
		{
			SessionID: input.SessionID,
			Type:      "agent_started",
			Message:   "Planner Agent 正在生成研究计划",
		},
		{
			SessionID: input.SessionID,
			Type:      "workflow_step",
			Message:   "已生成初始研究计划",
			Payload: map[string]any{
				"agent": "planner",
			},
		},
		{
			SessionID: input.SessionID,
			Type:      "tool_call",
			Message:   "Researcher Agent 调用 mock_search",
			Payload: map[string]any{
				"tool": "mock_search",
			},
		},
		{
			SessionID: input.SessionID,
			Type:      "tool_result",
			Message:   "mock_search 返回 3 条候选资料",
			Payload: map[string]any{
				"count": 3,
			},
		},
		{
			SessionID: input.SessionID,
			Type:      "approval_required",
			Message:   "报告大纲已生成，等待用户审批",
			Payload: map[string]any{
				"outline": []string{"背景", "核心发现", "证据分析", "风险", "建议"},
			},
		},
	}

	for _, event := range steps {
		// 每一步之间留一点延迟，让 SSE 前端能看到 timeline 逐步推进。
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.delay()):
		}
		if err := sink.Emit(ctx, event); err != nil {
			return err
		}
	}

	return sink.SetStatus(ctx, input.SessionID, session.StatusWaitingApproval)
}

// delay 返回 mock 每一步之间的等待时间。
func (r *MockRunner) delay() time.Duration {
	if r == nil || r.Delay <= 0 {
		return 350 * time.Millisecond
	}
	return r.Delay
}
