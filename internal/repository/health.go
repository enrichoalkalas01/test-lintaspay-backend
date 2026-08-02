package repository

import (
	"context"
	"time"

	"test-lintaspay/internal/domain/entity"

	"gorm.io/gorm"
)

type healthRepository struct {
	db *gorm.DB
}

func NewHealthRepository(db *gorm.DB) entity.HealthRepository {
	return &healthRepository{db: db}
}

func (r *healthRepository) PingDB(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}
