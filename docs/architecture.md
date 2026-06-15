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
cmd/server                    HTTP 服务入口，只负责装配依赖和启动服务
internal/config               配置加载，集中读取环境变量
internal/domain/session       Session / Event 领域模型和领域错误
internal/app/research         研究任务业务层，编排 Session、事件、Agent
internal/transport/httpapi    HTTP / SSE 适配层，只处理请求和响应
internal/event/sse            内存 SSE Broker，后续可替换为 Redis Pub/Sub
internal/infra/store/memory   内存存储实现
internal/infra/store/sqlite   SQLite / GORM 存储实现
internal/agent                Eino Agent 构建和调度
internal/workflow             Workflow / Graph Tool
internal/tools                外部工具和本地工具
internal/callbacks            Eino Callback Trace
internal/approval             Interrupt / Resume 审批
skills                        可复用 SKILL.md
web                           前端应用
reports                       生成的 Markdown 报告
data                          上传文件和缓存
docs                          项目文档
```

## 请求流程

```text
POST /sessions
  -> transport/httpapi
  -> research.Service
  -> domain/session
  -> infra/store

GET /sessions/{id}/events
  -> transport/httpapi
  -> research.Service.ListEvents
  -> research.Service.SubscribeEvents
  -> event/sse.Broker

POST /sessions/{id}/messages
  -> transport/httpapi
  -> research.Service.SendMessage
  -> infra/store / event/sse.Broker
  -> agent.Runner
  -> research.Service.Emit
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
