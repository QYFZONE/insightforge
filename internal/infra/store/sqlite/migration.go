package sqlite

import (
	"context"

	dbmodel "insightforge/internal/infra/store/sqlite/model"
)

// migrate 使用 GORM 自动迁移当前需要的表结构。
func (s *Store) migrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&dbmodel.Session{}, &dbmodel.Event{})
}
