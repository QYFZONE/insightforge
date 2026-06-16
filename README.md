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

复制环境变量示例：

```powershell
Copy-Item .env.example .env
```

启动后端服务：

```powershell
go run ./cmd/server
```

服务启动时会自动读取项目根目录下的 `.env`。

健康检查：

```powershell
curl http://localhost:8080/healthz
```

创建会话并连接 SSE 事件流：

```powershell
$session = Invoke-RestMethod -Method Post -Uri http://localhost:8080/sessions -ContentType "application/json" -Body '{"topic":"Eino Agent 调研"}'
curl.exe -N "http://localhost:8080/sessions/$($session.id)/events"
```

## 目录说明

```text
cmd/server                    服务入口，只负责装配依赖和启动 HTTP 服务
internal/config               配置加载
internal/domain/session       Session / Event 领域模型
internal/app/research         研究任务业务层
internal/transport/httpapi    HTTP / SSE 适配层
internal/event/sse            SSE 事件推送
internal/infra/store/memory   内存存储实现
internal/infra/store/sqlite   SQLite / GORM 存储实现
internal/agent                Agent 执行器，包含 mock 和 Ark 真实模型实现
skills                        可复用 SKILL.md
web                           前端应用
reports                       生成的报告
data                          上传文件和缓存
docs                          项目文档
```

## 环境变量

```text
HTTP_ADDR      HTTP 监听地址，默认 :8080
STORE_DRIVER   存储类型，支持 memory / sqlite
SQLITE_PATH    SQLite 数据库文件路径，默认 data/insightforge.db
AGENT_DRIVER   Agent 类型，支持 mock / ark
ARK_API_KEY    火山 Ark API Key，AGENT_DRIVER=ark 时使用
ARK_MODEL_ID   火山 Ark 模型接入点 ID，AGENT_DRIVER=ark 时使用
ARK_BASE_URL   火山 Ark API 地址，默认 https://ark.cn-beijing.volces.com/api/v3
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
