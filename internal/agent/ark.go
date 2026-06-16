package agent

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/agenticark"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"

	"insightforge/internal/domain/session"
)

// ArkConfig 收集 Ark 模型运行所需的最小配置。
type ArkConfig struct {
	APIKey  string
	ModelID string
	BaseURL string
}

// ArkRunner 是真实 Ark/Eino Agent 的执行器骨架。
// 它实现 agent.Runner 接口，所以可以被 research.Service 直接调度。
type ArkRunner struct {
	modelID   string
	chatModel model.AgenticModel
}

// NewArkRunner 创建 ArkRunner，并在启动阶段提前检查必要配置。
func NewArkRunner(ctx context.Context, cfg ArkConfig) (*ArkRunner, error) {
	// 先清理配置里的空白字符，避免 .env 中多出的空格造成隐蔽问题。
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.ModelID = strings.TrimSpace(cfg.ModelID)
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)

	// 配置缺失时直接返回中文错误，让启动失败原因更清楚。
	if cfg.APIKey == "" {
		return nil, errors.New("缺少 ARK_API_KEY")
	}
	if cfg.ModelID == "" {
		return nil, errors.New("缺少 ARK_MODEL_ID")
	}
	if cfg.BaseURL == "" {
		return nil, errors.New("缺少 ARK_BASE_URL")
	}

	// ChatModel 在 Runner 创建时初始化一次，后续请求复用它。
	chatModel, err := agenticark.New(ctx, &agenticark.Config{
		APIKey:  cfg.APIKey,
		Model:   cfg.ModelID,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return nil, err
	}

	return &ArkRunner{
		modelID:   cfg.ModelID,
		chatModel: chatModel,
	}, nil
}

// Run 是 ArkRunner 的主流程，只负责编排步骤，不把模型调用细节塞在一起。
func (r *ArkRunner) Run(ctx context.Context, sink EventSink, input RunInput) error {
	// 第一步：标记会话进入 running，并给前端 timeline 发开始事件。
	if err := r.start(ctx, sink, input); err != nil {
		return err
	}

	// 第二步：把历史事件整理成模型可理解的消息列表。
	messages := r.buildMessages(input)

	// 第三步：调用真实模型。模型调用细节集中在 streamGenerate。
	answer, err := r.streamGenerate(ctx, sink, input, messages)
	if err != nil {
		return err
	}

	// 第四步：把模型输出转换成系统统一的 session.Event。
	if err := r.emitAnswer(ctx, sink, input, answer); err != nil {
		return err
	}

	// 第五步：流程成功结束，更新会话状态。
	return r.complete(ctx, sink, input)
}

// start 负责发出 Agent 启动事件，让 HTTP/SSE 层能看到任务已经开始。
func (r *ArkRunner) start(ctx context.Context, sink EventSink, input RunInput) error {
	// 状态先落库，再发事件；这样前端刷新历史时不会看到状态滞后。
	if err := sink.SetStatus(ctx, input.SessionID, session.StatusRunning); err != nil {
		return err
	}

	// timeline 事件只放对用户/前端有意义的信息，不暴露 apiKey。
	return sink.Emit(ctx, session.Event{
		SessionID: input.SessionID,
		Type:      "agent_started",
		Message:   "Ark Agent 准备开始真实模型调用",
		Payload: map[string]any{
			"topic": input.Topic,
			"model": r.modelID,
		},
	})
}

// buildMessages 把领域事件转换成模型消息。
// 现在先不做压缩，后续上下文压缩会集中改这里。
func (r *ArkRunner) buildMessages(input RunInput) []*schema.AgenticMessage {
	messages := []*schema.AgenticMessage{
		schema.SystemAgenticMessage("你是 InsightForge 的研究助手，请给出结构清晰、可执行的回答。"),
		schema.UserAgenticMessage("当前研究主题：" + input.Topic),
	}

	// 只把真实对话消息放进模型上下文，tool_call/error 等 timeline 事件不进入上下文。
	for _, event := range input.History {
		text := strings.TrimSpace(event.Message)
		if text == "" {
			continue
		}
		switch event.Type {
		case "user_message":
			messages = append(messages, schema.UserAgenticMessage(text))
		case "assistant_message":
			messages = append(messages, assistantAgenticMessage(text))
		}
	}
	if !hasCurrentUserMessage(input.History, input.UserText) {
		messages = append(messages, schema.UserAgenticMessage(strings.TrimSpace(input.UserText)))
	}

	return messages
}

// generate 保留非流式模型调用，便于本地诊断或后续做配置化切换。
func (r *ArkRunner) generate(ctx context.Context, messages []*schema.AgenticMessage) (string, error) {
	response, err := r.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	answer := strings.TrimSpace(assistantText(response))
	if answer == "" {
		return "", errors.New("Ark 模型返回内容为空")
	}

	return answer, nil
}

// streamGenerate 调用流式模型接口。
// 每个 chunk 用于实时推送 assistant_delta，结束后交给 Eino 合成最终回答。
func (r *ArkRunner) streamGenerate(ctx context.Context, sink EventSink, input RunInput, messages []*schema.AgenticMessage) (string, error) {
	stream, err := r.chatModel.Stream(ctx, messages)
	if err != nil {
		return "", err
	}
	defer stream.Close()
	chunks := make([]*schema.AgenticMessage, 0, 16)
	for {
		frame, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		chunks = append(chunks, frame)
		delta := assistantText(frame)
		if delta != "" {
			if err := r.emitAssistantDelta(ctx, sink, input, delta); err != nil {
				return "", err
			}
		}
	}
	return finalAnswerFromChunks(chunks)
}

// finalAnswerFromChunks 把流式 AgenticMessage chunks 合成为最终回答。
func finalAnswerFromChunks(chunks []*schema.AgenticMessage) (string, error) {
	if len(chunks) == 0 {
		return "", errors.New("Ark 模型返回内容为空")
	}

	finalMessage, err := schema.ConcatAgenticMessages(chunks)
	if err != nil {
		return "", err
	}
	answer := strings.TrimSpace(assistantText(finalMessage))

	if answer == "" {
		return "", errors.New("Ark 模型返回内容为空")
	}
	return answer, nil
}

// assistantText 提取 AgenticMessage 里的 assistant 文本。
func assistantText(message *schema.AgenticMessage) string {
	if message == nil {
		return ""
	}

	var builder strings.Builder
	for _, block := range message.ContentBlocks {
		if block != nil && block.AssistantGenText != nil {
			builder.WriteString(block.AssistantGenText.Text)
		}
	}
	return builder.String()
}

// hasCurrentUserMessage 判断历史里是否已经包含本次用户输入。
func hasCurrentUserMessage(history []session.Event, userText string) bool {
	userText = strings.TrimSpace(userText)
	if userText == "" {
		return false
	}
	for _, event := range history {
		if event.Type == "user_message" && strings.TrimSpace(event.Message) == userText {
			return true
		}
	}
	return false
}

// assistantAgenticMessage 把历史 assistant 文本还原成 AgenticMessage。
func assistantAgenticMessage(text string) *schema.AgenticMessage {
	block := schema.NewContentBlock(&schema.AssistantGenText{Text: text})
	// Ark Responses API 要求历史 assistant 消息携带 status；模型真实返回的 block 会自带，
	// 但我们从数据库文本还原时需要补上这个适配字段。
	block.Extra = map[string]any{
		"ark-item-status": responses.ItemStatus_completed.String(),
	}

	return &schema.AgenticMessage{
		Role:          schema.AgenticRoleTypeAssistant,
		ContentBlocks: []*schema.ContentBlock{block},
	}
}

// emitAssistantDelta 推送流式 assistant 片段，不写入历史。
func (r *ArkRunner) emitAssistantDelta(ctx context.Context, sink EventSink, input RunInput, delta string) error {
	if delta == "" {
		return nil
	}

	return sink.EmitTransient(ctx, session.Event{
		SessionID: input.SessionID,
		Type:      "assistant_delta",
		Message:   delta,
		Payload: map[string]any{
			"model": r.modelID,
		},
	})
}

// emitAnswer 把模型回答写成统一事件，方便前端 timeline 和历史记录复用。
func (r *ArkRunner) emitAnswer(ctx context.Context, sink EventSink, input RunInput, answer string) error {
	return sink.Emit(ctx, session.Event{
		SessionID: input.SessionID,
		Type:      "assistant_message",
		Message:   answer,
		Payload: map[string]any{
			"model": r.modelID,
		},
	})
}

// complete 只负责最终状态更新，避免状态变更散落在 Run 的各处。
func (r *ArkRunner) complete(ctx context.Context, sink EventSink, input RunInput) error {
	return sink.SetStatus(ctx, input.SessionID, session.StatusCompleted)
}
