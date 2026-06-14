# Architecture

## Backend Layout

```text
cmd/server          server entrypoint
internal/agent      Eino agents and orchestration
internal/workflow   graph tools and deterministic workflows
internal/tools      external tools
internal/session    session models and service
internal/sse        event streaming
internal/callbacks  Eino callback handlers
internal/approval   interrupt/resume approval state
internal/store      persistence layer
skills              reusable SKILL.md files
web                 frontend app
reports             generated reports
data                uploaded documents and cache
```

## Request Flow

```text
POST /sessions
  -> create session

GET /sessions/{id}/events
  -> open SSE stream

POST /sessions/{id}/messages
  -> start agent run
  -> stream agent events
  -> write final report
```

## Agent Flow

```text
Supervisor
  -> Planner
  -> Researcher
  -> Analyst
  -> Writer
  -> Approval
  -> Writer
  -> Reviewer
```
