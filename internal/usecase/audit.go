package usecase

import (
	"context"

	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/pkg/common/httpresponse"

	"go.uber.org/zap"
)

type auditLogUsecase struct {
	repo entity.AuditLogRepository
	log  *zap.Logger
}

func NewAuditLogUsecase(repo entity.AuditLogRepository, log *zap.Logger) entity.AuditLogUsecase {
	return &auditLogUsecase{repo: repo, log: log}
}

func (u *auditLogUsecase) List(ctx context.Context, f *entity.AuditLogFilter) ([]entity.AuditLog, int64, error) {
	f.Normalize()

	if err := validateDateRange(f.DateFrom, f.DateTo); err != nil {
		return nil, 0, err
	}

	rows, total, err := u.repo.FindAll(ctx, f)
	if err != nil {
		return nil, 0, httpresponse.Internal("failed to fetch audit logs").WithErr(err)
	}

	return rows, total, nil
}
