package repository

import (
	"context"

	"test-lintaspay/internal/domain/entity"

	"gorm.io/gorm"
)

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) entity.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditLogRepository) FindAll(ctx context.Context, f *entity.AuditLogFilter) ([]entity.AuditLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.AuditLog{})

	if f.EntityID != "" {
		query = query.Where("entity_id = ?", f.EntityID)
	}
	if f.Action != "" {
		query = query.Where("action = ?", f.Action)
	}
	if f.DateFrom != "" {
		query = query.Where("created_at >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		query = query.Where("created_at < DATE_ADD(?, INTERVAL 1 DAY)", f.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []entity.AuditLog
	err := query.
		Order("created_at DESC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}
