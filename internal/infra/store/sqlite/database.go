package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Config 是 SQLite Store 的启动配置。
type Config struct {
	Path string
}

// Open 打开 SQLite 数据库，完成连通性检查和自动迁移。
func Open(ctx context.Context, cfg Config) (*Store, error) {
	path := strings.TrimSpace(cfg.Path)
	if path == "" {
		return nil, errors.New("sqlite: path is required")
	}

	// SQLite 文件可能位于 data/ 之类的目录，打开前先确保目录存在。
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	// GORM 的 db 对象被 Store 持有，业务层不直接接触 ORM。
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}
	// 启动阶段尽早失败，避免服务启动后才发现数据库不可用。
	if err := store.ping(ctx); err != nil {
		_ = store.Close()
		return nil, err
	}
	if err := store.migrate(ctx); err != nil {
		_ = store.Close()
		return nil, err
	}

	return store, nil
}

// Close 关闭底层数据库连接。
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// ping 检查数据库连接是否可用。
func (s *Store) ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
