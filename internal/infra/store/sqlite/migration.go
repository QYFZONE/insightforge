package sqlite

import (
	"context"

	dbmodel "insightforge/internal/infra/store/sqlite/model"
)

func (s *Store) migrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&dbmodel.Session{}, &dbmodel.Event{})
}
