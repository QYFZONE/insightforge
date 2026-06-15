package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"insightforge/internal/agent"
	"insightforge/internal/app/research"
	"insightforge/internal/config"
	"insightforge/internal/event/sse"
	"insightforge/internal/infra/store/memory"
	sqlitestore "insightforge/internal/infra/store/sqlite"
	"insightforge/internal/transport/httpapi"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	sessionStore, closeStore, err := buildSessionStore(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer closeStore()

	researchService, err := research.New(research.Config{
		Sessions:    sessionStore,
		Events:      sse.NewBroker(),
		RunResearch: agent.RunMockResearch,
	})
	if err != nil {
		log.Fatal(err)
	}

	api, err := httpapi.New(httpapi.Config{
		Research: researchService,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("InsightForge server listening on %s, store=%s", cfg.HTTPAddr, cfg.StoreDriver)
	if err := http.ListenAndServe(cfg.HTTPAddr, api.Handler()); err != nil {
		log.Fatal(err)
	}
}

func buildSessionStore(ctx context.Context, cfg config.Config) (research.SessionStore, func(), error) {
	switch cfg.StoreDriver {
	case config.StoreDriverMemory:
		return memory.NewStore(), func() {}, nil
	case config.StoreDriverSQLite:
		store, err := sqlitestore.Open(ctx, sqlitestore.Config{Path: cfg.SQLitePath})
		if err != nil {
			return nil, nil, err
		}
		return store, func() {
			if err := store.Close(); err != nil {
				log.Printf("关闭 SQLite Store 失败：%v", err)
			}
		}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported STORE_DRIVER: %s", cfg.StoreDriver)
	}
}
