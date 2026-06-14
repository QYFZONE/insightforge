# InsightForge

InsightForge is a multi-agent research report platform built with Go and CloudWeGo Eino.

The goal is to turn a user research topic into a structured Markdown report through:

- multi-agent collaboration
- tool calling
- document parsing
- RAG workflow
- human approval
- SSE timeline
- session history

## Target Architecture

```text
Web UI
  -> HTTP API / SSE
  -> Session Manager
  -> Agent Orchestrator
  -> Supervisor / Planner / Researcher / Analyst / Writer / Reviewer
  -> Tools: search, web reader, PDF parser, RAG, report writer
  -> Storage: sessions, events, reports, uploaded documents
```

## Development

```powershell
go run ./cmd/server
```

Health check:

```powershell
curl http://localhost:8080/healthz
```

SSE demo:

```powershell
curl http://localhost:8080/events
```

## Roadmap

1. HTTP server + SSE
2. Session and event storage
3. Eino Agent runtime
4. Planner / Researcher / Analyst / Writer agents
5. RAG Graph Tool
6. Interrupt / Resume approval
7. Skill Middleware
8. Web UI
9. Search, PDF, and URL readers
10. Docker Compose
