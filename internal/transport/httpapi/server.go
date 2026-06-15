package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"insightforge/internal/domain/session"
)

type ResearchService interface {
	CreateSession(ctx context.Context, topic string) (session.Session, error)
	ListSessions(ctx context.Context) ([]session.Session, error)
	SendMessage(ctx context.Context, sessionID string, content string) error
	ListEvents(ctx context.Context, sessionID string) ([]session.Event, error)
	SubscribeEvents(sessionID string) (<-chan session.Event, func())
}

type Config struct {
	Research ResearchService
}

type Server struct {
	research ResearchService
}

func New(cfg Config) (*Server, error) {
	if cfg.Research == nil {
		return nil, errors.New("httpapi: research service is required")
	}

	return &Server{research: cfg.Research}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /sessions", s.handleCreateSession)
	mux.HandleFunc("GET /sessions", s.handleListSessions)
	mux.HandleFunc("GET /sessions/", s.handleSessionRoute)
	mux.HandleFunc("POST /sessions/", s.handleSessionRoute)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
