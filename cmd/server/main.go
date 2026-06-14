package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"insightforge/internal/agent"
	"insightforge/internal/session"
	"insightforge/internal/sse"
)

type app struct {
	sessions *session.Store
	broker   *sse.Broker
}

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	a := &app{
		sessions: session.NewStore(),
		broker:   sse.NewBroker(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.handleHealth)
	mux.HandleFunc("POST /sessions", a.handleCreateSession)
	mux.HandleFunc("GET /sessions", a.handleListSessions)
	mux.HandleFunc("GET /sessions/", a.handleSessionRoute)
	mux.HandleFunc("POST /sessions/", a.handleSessionRoute)

	log.Printf("InsightForge server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (a *app) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req session.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	item, err := a.sessions.Create(r.Context(), req.Topic)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (a *app) handleListSessions(w http.ResponseWriter, r *http.Request) {
	items, err := a.sessions.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sessions": items,
	})
}

func (a *app) handleSessionRoute(w http.ResponseWriter, r *http.Request) {
	sessionID, tail, ok := parseSessionPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch {
	case r.Method == http.MethodGet && tail == "events":
		a.handleSessionEvents(w, r, sessionID)
	case r.Method == http.MethodPost && tail == "messages":
		a.handleCreateMessage(w, r, sessionID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (a *app) handleSessionEvents(w http.ResponseWriter, r *http.Request, sessionID string) {
	// TODO:
	// 1. 用 a.sessions.Get 确认 Session 存在
	// 2. 设置 SSE headers
	// 3. 获取 http.Flusher
	// 4. 用 a.sessions.ListEvents 补发历史事件
	// 5. 调用 a.broker.Subscribe(sessionID) 订阅实时事件
	// 6. 循环 select：
	//    - 请求取消时退出
	//    - 收到事件时 writeSSE + Flush
	//    - 心跳时写 ": ping\n\n"
	_ = sessionID
	writeError(w, http.StatusNotImplemented, "TODO: implement session events SSE")
}

func (a *app) handleCreateMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	item, err := a.sessions.Get(r.Context(), sessionID)
	if err != nil {
		writeSessionError(w, err)
		return
	}

	var req session.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	if err := a.Emit(r.Context(), session.Event{
		SessionID: sessionID,
		Type:      "user_message",
		Message:   req.Content,
	}); err != nil {
		writeSessionError(w, err)
		return
	}

	go agent.RunMockResearch(context.Background(), a, sessionID, item.Topic)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":     "accepted",
		"session_id": sessionID,
	})
}

func (a *app) Emit(ctx context.Context, event session.Event) error {
	// TODO:
	// 1. 调用 a.sessions.AddEvent 保存事件
	// 2. 调用 a.broker.Publish 推送实时事件
	_ = ctx
	_ = event
	return errors.New("TODO: implement app.Emit")
}

func (a *app) SetStatus(ctx context.Context, sessionID string, status session.Status) error {
	// TODO:
	// 调用 a.sessions.SetStatus
	_ = ctx
	_ = sessionID
	_ = status
	return errors.New("TODO: implement app.SetStatus")
}

func parseSessionPath(path string) (sessionID string, tail string, ok bool) {
	rest := strings.TrimPrefix(path, "/sessions/")
	if rest == path {
		return "", "", false
	}

	parts := strings.Split(rest, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func writeSSE(w http.ResponseWriter, event session.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", event.Type); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return nil
}

func writeSessionError(w http.ResponseWriter, err error) {
	if errors.Is(err, session.ErrNotFound) {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": message,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write json response: %v", err)
	}
}
