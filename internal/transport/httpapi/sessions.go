package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// handleCreateSession 处理创建研究会话请求。
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeInvalidJSON(w)
		return
	}
	item, err := s.research.CreateSession(r.Context(), req.Topic)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toSessionResponse(item))
}

// handleListSessions 返回当前保存的研究会话列表。
func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	items, err := s.research.ListSessions(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listSessionsResponse{Sessions: toSessionResponses(items)})
}

// handleSessionRoute 分发 /sessions/{id}/... 下面的子路由。
func (s *Server) handleSessionRoute(w http.ResponseWriter, r *http.Request) {
	sessionID, tail, ok := parseSessionPath(r.URL.Path)
	if !ok {
		writeRouteNotFound(w)
		return
	}

	switch {
	case r.Method == http.MethodGet && tail == "events":
		s.handleSessionEvents(w, r, sessionID)
	case r.Method == http.MethodPost && tail == "messages":
		s.handleCreateMessage(w, r, sessionID)
	default:
		writeRouteNotFound(w)
	}
}

// handleSessionEvents 通过 SSE 返回历史事件，并持续推送实时事件。
func (s *Server) handleSessionEvents(w http.ResponseWriter, r *http.Request, sessionID string) {
	events, err := s.research.ListEvents(r.Context(), sessionID)
	if err != nil {
		writeResearchError(w, err)
		return
	}
	// SSE 响应头必须在第一次 Write 之前设置。
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Flusher 是 SSE 必需能力；没有它就无法及时把事件刷给客户端。
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeStreamingUnsupported(w)
		return
	}

	// 先写一条注释帧，帮助客户端确认连接已经建立。
	if _, err := w.Write([]byte(": connected\n\n")); err != nil {
		return
	}
	flusher.Flush()

	// 先补历史，再订阅实时，保证刷新页面时能看到完整 timeline。
	for _, event := range events {
		if err := writeSSE(w, event); err != nil {
			return
		}
		flusher.Flush()
	}

	ch, cancel := s.research.SubscribeEvents(sessionID)
	defer cancel()

	// 心跳避免代理或浏览器长时间无数据时断开连接。
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			if err := writeSSE(w, event); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := w.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// handleCreateMessage 保存用户消息，并触发后台 Agent 执行。
func (s *Server) handleCreateMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req sendMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeInvalidJSON(w)
		return
	}

	err = s.research.SendMessage(r.Context(), sessionID, req.Content)
	if err != nil {
		writeResearchError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, sendMessageResponse{
		Status:    "accepted",
		SessionID: sessionID,
	})
}

// parseSessionPath 解析 /sessions/{sessionID}/{tail} 这种固定两段路径。
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
