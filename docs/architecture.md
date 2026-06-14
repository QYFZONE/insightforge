# 系统架构

## 总体架构

```text
Web UI
  -> HTTP API / SSE
  -> Session Manager
  -> Agent Orchestrator
  -> Multi-Agent Runtime
  -> Tools / Workflow / Skill
  -> Store / Reports / Data
```

## 后端目录结构

```text
cmd/server          HTTP 服务入口
internal/agent      Eino Agent 构建和调度
internal/workflow   Workflow / Graph Tool
internal/tools      外部工具和本地工具
internal/session    Session 模型和服务
internal/sse        SSE 事件推送
internal/callbacks  Eino Callback Trace
internal/approval   Interrupt / Resume 审批
internal/store      数据持久化
skills              可复用 SKILL.md
web                 前端应用
reports             生成的 Markdown 报告
data                上传文件和缓存
docs                项目文档
```

## 请求流程

```text
POST /sessions
  -> 创建 Session

GET /sessions/{id}/events
  -> 建立 SSE 事件流

POST /sessions/{id}/messages
  -> 启动 Agent 任务
  -> 推送 Agent 事件
  -> 等待审批或生成报告

POST /sessions/{id}/approvals
  -> 提交审批结果
  -> Resume Agent 执行

GET /reports/{id}
  -> 查看生成的 Markdown 报告
```

## Agent 协作流程

```text
Supervisor Agent
  -> Planner Agent
  -> Researcher Agent
  -> Analyst Agent
  -> Writer Agent
  -> Human Approval
  -> Writer Agent
  -> Reviewer Agent
```

## 核心模块

### Agent Orchestrator

负责创建和运行 Eino Agent：

- 初始化模型
- 注册工具
- 注册 Middleware
- 注册 Callback
- 管理 Runner

### Workflow / Graph Tool

用于确定性多步骤处理：

```text
load documents
  -> split chunks
  -> score chunks
  -> select evidence
  -> summarize evidence
```

### Tool Layer

工具层负责接入外部能力：

- Search Tool
- URL Reader
- PDF Parser
- Local File Reader
- Report Writer

### Approval

使用 Interrupt / Resume 实现人工审批：

```text
生成大纲
  -> interrupt
  -> 用户审批
  -> resume
  -> 生成完整报告
```

### SSE

SSE 用于把 Agent 执行状态推送给前端：

```text
tool_call
tool_result
workflow_step
approval_required
report_saved
error
```

### Store

第一版可以使用内存存储，后续升级为 SQLite：

- sessions
- messages
- agent_events
- approvals
- reports
- documents
