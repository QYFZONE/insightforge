package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"insightforge/internal/domain/session"
)

// ResearchService 是 HTTP 层依赖的业务服务端口。
type ResearchService interface {
	CreateSession(ctx context.Context, topic string) (session.Session, error)
	ListSessions(ctx context.Context) ([]session.Session, error)
	SendMessage(ctx context.Context, sessionID string, content string) error
	ListEvents(ctx context.Context, sessionID string) ([]session.Event, error)
	SubscribeEvents(sessionID string) (<-chan session.Event, func())
}

// Config 收集 HTTP Server 的依赖。
type Config struct {
	Research ResearchService
}

// Server 负责 HTTP 路由和协议适配，不承载业务规则。
type Server struct {
	research ResearchService
}

// New 创建 HTTP Server，并校验必要依赖。
func New(cfg Config) (*Server, error) {
	if cfg.Research == nil {
		return nil, errors.New("httpapi: research service is required")
	}

	return &Server{research: cfg.Research}, nil
}

// Handler 注册所有 HTTP 路由。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /sessions", s.handleCreateSession)
	mux.HandleFunc("GET /sessions", s.handleListSessions)
	mux.HandleFunc("GET /sessions/", s.handleSessionRoute)
	mux.HandleFunc("POST /sessions/", s.handleSessionRoute)
	return mux
}

// handleHealth 返回服务存活状态。
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
