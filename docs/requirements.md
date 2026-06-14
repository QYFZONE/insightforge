# Requirements

## Product Goal

InsightForge helps users produce research reports from a topic and source materials.

The system should plan the research, collect evidence, analyze sources, generate an outline, wait for user approval, and then generate a final Markdown report.

## Users

- developers doing technical research
- product managers preparing comparison reports
- students and researchers collecting sources
- engineers evaluating frameworks or tools

## Core Features

### Research Session

- create a session from a research topic
- store messages and agent events
- reopen historical sessions

### Multi-Agent Workflow

- Supervisor Agent controls the task
- Planner Agent creates a research plan
- Researcher Agent collects sources
- Analyst Agent extracts claims and evidence
- Writer Agent creates the report
- Reviewer Agent checks quality and risk

### Tools

- web search
- URL reader
- PDF parser
- local document reader
- report writer
- RAG evidence selector

### Human Approval

- pause after outline generation
- user approves or requests changes
- resume report generation after approval

### SSE Timeline

The frontend should receive events such as:

- session_started
- agent_started
- tool_call
- tool_result
- workflow_step
- approval_required
- report_saved
- error

## First Milestone

- Go HTTP server
- SSE endpoint
- in-memory session store
- mock agent timeline
- Markdown report save path
