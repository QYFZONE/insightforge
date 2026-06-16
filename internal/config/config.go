package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Store 和 Agent 的 driver 常量集中放在配置层，避免字符串散落在 main 里。
const (
	StoreDriverMemory = "memory"
	StoreDriverSQLite = "sqlite"

	AgentDriverMock = "mock"
	AgentDriverArk  = "ark"
)

// Config 是应用启动期需要的全部配置快照。
type Config struct {
	HTTPAddr    string
	StoreDriver string
	SQLitePath  string
	AgentDriver string
	ArkAPIKey   string
	ArkModelID  string
	ArkBaseURL  string
}

// Load 从 .env 和环境变量读取配置。
func Load() Config {
	// godotenv.Load 找不到 .env 时不会中断启动，部署环境可以直接使用系统环境变量。
	_ = godotenv.Load()

	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		StoreDriver: normalize(getenv("STORE_DRIVER", StoreDriverMemory)),
		SQLitePath:  getenv("SQLITE_PATH", "data/insightforge.db"),
		AgentDriver: normalize(getenv("AGENT_DRIVER", AgentDriverMock)),
		ArkAPIKey:   getenv("ARK_API_KEY", ""),
		ArkModelID:  getenv("ARK_MODEL_ID", ""),
		ArkBaseURL:  getenv("ARK_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
	}
}

// getenv 统一处理环境变量读取和默认值逻辑。
func getenv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

// normalize 用于规范 driver 类配置，避免大小写和空白导致匹配失败。
func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
