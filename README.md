# InsightForge

InsightForge 是一个基于 Go 和 CloudWeGo Eino 构建的多 Agent 智能研究报告生成平台。

用户输入研究主题后，系统会通过多 Agent 协作完成任务规划、资料收集、证据分析、大纲审批和 Markdown 报告生成。

## 核心能力

- Multi-Agent 协作：Supervisor / Planner / Researcher / Analyst / Writer / Reviewer
- Tool Calling：搜索、网页读取、PDF 解析、本地文档读取、报告保存
- Workflow / Graph Tool：编排 RAG 证据筛选和报告生成流程
- Human-in-the-loop：基于 Interrupt / Resume 的大纲审批
- SSE Timeline：实时展示 Agent 执行过程
- Session History：保存研究任务、消息、事件和报告
- Skill Middleware：复用报告写作、技术分析、引用规范等技能
- Callback Trace：记录模型调用、工具调用、耗时和错误

## 目标架构

```text
Web UI
  -> HTTP API / SSE
  -> Session Manager
  -> Agent Orchestrator
  -> Supervisor / Planner / Researcher / Analyst / Writer / Reviewer
  -> Tools: search, web reader, PDF parser, RAG, report writer
  -> Storage: sessions, events, reports, uploaded documents
```

## 本地开发

启动后端服务：

```powershell
go run ./cmd/server
```

健康检查：

```powershell
curl http://localhost:8080/healthz
```

SSE 事件流演示：

```powershell
curl http://localhost:8080/events
```

## 目录说明

```text
cmd/server          服务入口
internal/agent      Eino Agent 构建和调度
internal/workflow   Workflow / Graph Tool
internal/tools      搜索、网页、PDF、报告等工具
internal/session    Session 模型和服务
internal/sse        SSE 事件推送
internal/callbacks  Callback Trace
internal/approval   Interrupt / Resume 审批状态
internal/store      持久化层
skills              可复用 SKILL.md
web                 前端应用
reports             生成的报告
data                上传文件和缓存
docs                项目文档
```

## Roadmap

1. HTTP Server + SSE
2. Session 和事件存储
3. Eino Agent Runtime
4. Planner / Researcher / Analyst / Writer 多 Agent
5. RAG Graph Tool
6. Interrupt / Resume 大纲审批
7. Skill Middleware
8. Web UI
9. 搜索、PDF 和 URL Reader
10. Docker Compose 部署
