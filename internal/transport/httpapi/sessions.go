package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	// 1. 定义 var req createSessionRequest。
	// 2. 用 json.NewDecoder(r.Body).Decode(&req) 解析请求体。
	// 3. 解析失败时返回 400。
	// 4. 调用 s.research.CreateSession(r.Context(), req.Topic)。
	// 5. 成功时用 toSessionResponse(item) 返回 201。
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

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	// 1. 调用 s.research.ListSessions(r.Context())。
	// 2. 出错时返回 500。
	// 3. 成功时用 toSessionResponses(items) 返回 listSessionsResponse。
	items, err := s.research.ListSessions(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listSessionsResponse{Sessions: toSessionResponses(items)})
}

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

func (s *Server) handleSessionEvents(w http.ResponseWriter, r *http.Request, sessionID string) {
	// 1. 用 s.research.ListEvents(r.Context(), sessionID) 获取历史事件。
	// 2. 设置 SSE headers：Content-Type、Cache-Control、Connection。
	// 3. 获取 http.Flusher；拿不到就返回 500。
	// 4. 先把历史事件逐条 writeSSE + Flush。
	// 5. 调用 s.research.SubscribeEvents(sessionID) 订阅实时事件，记得 defer cancel()。
	// 6. select 循环：
	//    - r.Context().Done()：客户端断开，退出。
	//    - 收到事件：writeSSE + Flush。
	//    - 心跳 ticker：写入 ": ping\n\n" + Flush。
	events, err := s.research.ListEvents(r.Context(), sessionID)
	if err != nil {
		writeResearchError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeStreamingUnsupported(w)
		return
	}

	if _, err := w.Write([]byte(": connected\n\n")); err != nil {
		return
	}
	flusher.Flush()

	for _, event := range events {
		if err := writeSSE(w, event); err != nil {
			return
		}
		flusher.Flush()
	}

	ch, cancel := s.research.SubscribeEvents(sessionID)
	defer cancel()

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

func (s *Server) handleCreateMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	// 1. 定义 var req sendMessageRequest。
	// 2. 用 json.NewDecoder(r.Body).Decode(&req) 解析请求体。
	// 3. 调用 s.research.SendMessage(r.Context(), sessionID, req.Content)。
	// 4. 用 writeResearchError 统一翻译业务错误。
	// 5. 成功时返回 sendMessageResponse。
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
