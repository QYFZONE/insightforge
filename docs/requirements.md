# 项目需求

## 产品目标

InsightForge 的目标是帮助用户从一个研究主题出发，自动生成结构化研究报告。

系统需要完成任务规划、资料收集、证据分析、报告大纲生成、人工审批和最终 Markdown 报告生成。

## 目标用户

- 需要做技术选型的研发工程师
- 需要快速整理竞品或方案的产品经理
- 需要收集资料和生成报告的学生或研究人员
- 需要评估框架、工具或技术路线的团队

## 核心场景

用户输入：

```text
调研 CloudWeGo Eino 和 LangChain 的差异，并给出 Go 项目技术选型建议。
```

系统输出：

```text
结构化 Markdown 报告
关键结论
证据来源
优缺点对比
适用场景
风险提示
参考资料
```

## 功能需求

### Research Session

- 用户可以创建研究任务
- 系统为每个任务创建 Session
- Session 保存用户输入、Agent 消息、工具调用、事件和报告路径
- 用户可以重新打开历史 Session

### Multi-Agent Workflow

系统至少包含以下 Agent：

- Supervisor Agent：负责整体调度
- Planner Agent：拆解研究计划
- Researcher Agent：收集资料和来源
- Analyst Agent：提取观点、证据和风险
- Writer Agent：生成大纲和报告
- Reviewer Agent：检查报告质量和引用完整性

第一版可以先使用确定性顺序：

```text
Planner -> Researcher -> Analyst -> Writer -> Reviewer
```

后续再升级为 Supervisor 动态调度。

### Tools

系统需要支持以下工具：

- Web Search
- URL Reader
- PDF Parser
- Local Document Reader
- RAG Evidence Selector
- Report Writer

第一版可以先实现：

- 本地 Markdown / TXT 读取
- Mock Search
- Markdown 报告保存

### RAG Evidence Selection

系统需要把资料处理成：

- 文档分块
- 相关性评分
- Top-K 证据片段
- 来源记录
- 证据摘要

第一版可以使用关键词评分，后续升级为 Embedding 或模型评分。

### Human Approval

Writer Agent 生成报告大纲后，系统需要暂停：

- 用户可以批准大纲
- 用户可以拒绝大纲
- 用户可以补充修改意见

批准后系统通过 Interrupt / Resume 继续生成完整报告。

### SSE Timeline

前端需要实时接收以下事件：

- session_started
- agent_started
- agent_finished
- tool_call
- tool_result
- workflow_step
- approval_required
- report_saved
- error

### Skill Middleware

系统支持通过 `SKILL.md` 管理可复用能力：

- report-writing
- tech-analysis
- citation-style
- risk-review

Agent 在需要时可以加载对应 Skill。

## 非功能需求

- 安全性：文件读取和写入必须限制在项目 data / reports 目录
- 可恢复：审批后可以从中断点继续执行
- 可观测：每次模型调用、工具调用和工作流节点都可追踪
- 可扩展：新增 Agent、Tool、Skill 时不影响主流程
- 易部署：支持 `.env` 配置和 Docker Compose

## 第一阶段验收标准

- 可以启动 HTTP 服务
- 可以创建 mock SSE timeline
- 可以创建 Session
- 可以保存 Agent Event
- 可以生成 Markdown 报告文件
- 可以在 Web UI 中看到任务执行过程
