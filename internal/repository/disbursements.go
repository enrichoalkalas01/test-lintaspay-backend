package repository

import (
	"context"
	"errors"
	"time"

	"test-lintaspay/internal/domain/entity"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type disbursementRepository struct {
	db *gorm.DB
}

func NewDisbursementRepository(db *gorm.DB) entity.DisbursementRepository {
	return &disbursementRepository{db: db}
}

func (r *disbursementRepository) Create(ctx context.Context, d *entity.Disbursement) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *disbursementRepository) CreateWithIdempotency(ctx context.Context, d *entity.Disbursement, idem *entity.IdempotencyKey) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(idem).Error; err != nil {
			if isDuplicateEntry(err) {
				return entity.ErrIdempotencyKeyExists
			}
			return err
		}

		return tx.Create(d).Error
	})
}

func (r *disbursementRepository) FindByID(ctx context.Context, id string) (*entity.Disbursement, error) {
	var d entity.Disbursement

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &d, nil
}

func (r *disbursementRepository) FindAll(ctx context.Context, f *entity.DisbursementFilter) ([]entity.Disbursement, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Disbursement{})

	if f.Search != "" {
		query = query.Where("recipient_name LIKE ?", "%"+f.Search+"%")
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
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

	var rows []entity.Disbursement
	err := query.
		Order(f.SortBy + " " + f.SortOrder).
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *disbursementRepository) UpdateStatus(ctx context.Context, id, status, note, actor string) (int64, error) {
	values := map[string]any{
		"status":      status,
		"approved_by": actor,
		"updated_at":  time.Now().UTC(),
	}
	if note != "" {
		values["note"] = note
	}

	result := r.db.WithContext(ctx).
		Model(&entity.Disbursement{}).
		Where("id = ? AND status = ?", id, entity.StatusPending).
		Updates(values)

	return result.RowsAffected, result.Error
}

func (r *disbursementRepository) SoftDelete(ctx context.Context, id string) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, entity.StatusPending).
		Delete(&entity.Disbursement{})

	return result.RowsAffected, result.Error
}

func (r *disbursementRepository) FindIdempotencyKey(ctx context.Context, key string) (*entity.IdempotencyKey, error) {
	var idem entity.IdempotencyKey

	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&idem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &idem, nil
}

func (r *disbursementRepository) DeleteIdempotencyKey(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("`key` = ?", key).Delete(&entity.IdempotencyKey{}).Error
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
