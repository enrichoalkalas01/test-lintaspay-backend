package usecase

import (
	"context"

	"test-lintaspay/internal/domain/entity"

	"go.uber.org/zap"
)

type healthUsecase struct {
	repo entity.HealthRepository
	log  *zap.Logger
}

func NewHealthUsecase(repo entity.HealthRepository, log *zap.Logger) entity.HealthUsecase {
	return &healthUsecase{repo: repo, log: log}
}

func (u *healthUsecase) Status(ctx context.Context) *entity.HealthStatus {
	status := &entity.HealthStatus{
		Status:   "ok",
		Database: "up",
	}

	if err := u.repo.PingDB(ctx); err != nil {
		u.log.Error("database health check failed", zap.Error(err))
		status.Status = "degraded"
		status.Database = "down"
	}

	return status
}
