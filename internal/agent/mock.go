package agent

import (
	"context"
	"time"

	"insightforge/internal/domain/session"
)

type EventSink interface {
	Emit(ctx context.Context, event session.Event) error
	SetStatus(ctx context.Context, sessionID string, status session.Status) error
}

// RunMockResearch 是第一阶段的假 Agent timeline。
// TODO: 第二阶段用 Eino Runner 替换这里，事件来源改为 Callback + Tool/Agent 输出。
func RunMockResearch(ctx context.Context, store EventSink, sessionID string, topic string) {
	_ = store.SetStatus(ctx, sessionID, session.StatusRunning)

	steps := []session.Event{
		{
			SessionID: sessionID,
			Type:      "session_started",
			Message:   "研究任务已启动",
			Payload: map[string]any{
				"topic": topic,
			},
		},
		{
			SessionID: sessionID,
			Type:      "agent_started",
			Message:   "Planner Agent 正在生成研究计划",
		},
		{
			SessionID: sessionID,
			Type:      "workflow_step",
			Message:   "已生成初始研究计划",
			Payload: map[string]any{
				"agent": "planner",
			},
		},
		{
			SessionID: sessionID,
			Type:      "tool_call",
			Message:   "Researcher Agent 调用 mock_search",
			Payload: map[string]any{
				"tool": "mock_search",
			},
		},
		{
			SessionID: sessionID,
			Type:      "tool_result",
			Message:   "mock_search 返回 3 条候选资料",
			Payload: map[string]any{
				"count": 3,
			},
		},
		{
			SessionID: sessionID,
			Type:      "approval_required",
			Message:   "报告大纲已生成，等待用户审批",
			Payload: map[string]any{
				"outline": []string{"背景", "核心发现", "证据分析", "风险", "建议"},
			},
		},
	}

	for _, event := range steps {
		select {
		case <-ctx.Done():
			return
		case <-time.After(350 * time.Millisecond):
		}
		_ = store.Emit(ctx, event)
	}

	_ = store.SetStatus(ctx, sessionID, session.StatusWaitingApproval)
}
