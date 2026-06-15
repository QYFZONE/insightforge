package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	StoreDriverMemory = "memory"
	StoreDriverSQLite = "sqlite"
)

type Config struct {
	HTTPAddr    string
	StoreDriver string
	SQLitePath  string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		StoreDriver: normalize(getenv("STORE_DRIVER", StoreDriverMemory)),
		SQLitePath:  getenv("SQLITE_PATH", "data/insightforge.db"),
	}
}

func getenv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
