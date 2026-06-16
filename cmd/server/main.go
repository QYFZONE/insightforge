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

// main 是服务启动入口，只负责装配依赖和启动 HTTP Server。
func main() {
	// 使用后台上下文完成启动期依赖初始化；请求级取消由 HTTP 层自己的 context 负责。
	ctx := context.Background()
	cfg := config.Load()

	// Store 和 Runner 都通过构造函数创建，避免业务层关心具体实现。
	sessionStore, closeStore, err := buildSessionStore(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer closeStore()

	agentRunner, err := buildAgentRunner(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// research.Service 是应用服务层，向 HTTP 层暴露稳定的业务接口。
	researchService, err := research.New(research.Config{
		Sessions: sessionStore,
		Events:   sse.NewBroker(),
		Runner:   agentRunner,
	})
	if err != nil {
		log.Fatal(err)
	}

	// HTTP Server 保持薄适配：只处理协议细节，不直接碰存储和 Agent。
	api, err := httpapi.New(httpapi.Config{
		Research: researchService,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("InsightForge server listening on %s, store=%s, agent=%s", cfg.HTTPAddr, cfg.StoreDriver, cfg.AgentDriver)
	if err := http.ListenAndServe(cfg.HTTPAddr, api.Handler()); err != nil {
		log.Fatal(err)
	}
}

// buildSessionStore 根据配置选择存储实现，并返回对应的关闭函数。
func buildSessionStore(ctx context.Context, cfg config.Config) (research.SessionStore, func(), error) {
	switch cfg.StoreDriver {
	case config.StoreDriverMemory:
		// memory 适合本地演示，进程重启后数据会丢失。
		return memory.NewStore(), func() {}, nil
	case config.StoreDriverSQLite:
		// sqlite 是当前默认持久化方案，后续可以在这里扩展 mysql/postgres。
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
		return nil, nil, fmt.Errorf("不支持的 STORE_DRIVER：%s", cfg.StoreDriver)
	}
}

// buildAgentRunner 根据配置选择 Agent 执行器。
func buildAgentRunner(ctx context.Context, cfg config.Config) (agent.Runner, error) {
	switch cfg.AgentDriver {
	case config.AgentDriverMock:
		// mock 先保证 HTTP/SSE/Store 主链路稳定，方便阶段性开发。
		return agent.NewMockRunner(), nil
	case config.AgentDriverArk:
		// ark 分支负责真实模型调用，具体实现放在 internal/agent/ark.go。
		return agent.NewArkRunner(ctx, agent.ArkConfig{
			APIKey:  cfg.ArkAPIKey,
			ModelID: cfg.ArkModelID,
			BaseURL: cfg.ArkBaseURL,
		})
	default:
		return nil, fmt.Errorf("不支持的 AGENT_DRIVER：%s", cfg.AgentDriver)
	}
}
