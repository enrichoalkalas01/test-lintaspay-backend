package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/pkg/common/httpresponse"
	"test-lintaspay/pkg/common/reqctx"
	"test-lintaspay/pkg/logger"
	"test-lintaspay/pkg/utils"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type disbursementUsecase struct {
	repo      entity.DisbursementRepository
	auditRepo entity.AuditLogRepository
	log       *zap.Logger
	idemTTL   time.Duration
}

func NewDisbursementUsecase(env *viper.Viper, repo entity.DisbursementRepository, auditRepo entity.AuditLogRepository, log *zap.Logger) entity.DisbursementUsecase {
	idemTTL := env.GetDuration("IDEMPOTENCY_TTL")
	if idemTTL == 0 {
		idemTTL = 24 * time.Hour
	}

	return &disbursementUsecase{
		repo:      repo,
		auditRepo: auditRepo,
		log:       log,
		idemTTL:   idemTTL,
	}
}

func CalculateAdminFee(amount int64) int64 {
	if amount >= 5_000_000 {
		return 5000
	}
	return 2500
}

func (u *disbursementUsecase) Create(ctx context.Context, req *entity.CreateDisbursementRequest, idempotencyKey string) (*entity.Disbursement, bool, error) {
	actor := reqctx.Username(ctx)
	now := time.Now().UTC()

	d := &entity.Disbursement{
		ID:            uuid.NewString(),
		RecipientName: req.RecipientName,
		AccountNumber: req.AccountNumber,
		BankCode:      req.BankCode,
		Amount:        req.Amount,
		AdminFee:      CalculateAdminFee(req.Amount),
		Note:          req.Note,
		Status:        entity.StatusPending,
		CreatedBy:     actor,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if idempotencyKey == "" {
		if err := u.repo.Create(ctx, d); err != nil {
			return nil, false, httpresponse.Internal("failed to create disbursement").WithErr(err)
		}

		u.writeAudit(ctx, d.ID, entity.AuditActionCreated, nil, d)
		return d, false, nil
	}

	if _, err := uuid.Parse(idempotencyKey); err != nil {
		return nil, false, httpresponse.BadRequest("Idempotency-Key must be a valid UUID")
	}

	responseBody, err := json.Marshal(d)
	if err != nil {
		return nil, false, httpresponse.Internal("failed to build idempotency cache").WithErr(err)
	}

	idem := &entity.IdempotencyKey{
		Key:          idempotencyKey,
		UserID:       reqctx.UserID(ctx),
		RequestHash:  hashRequest(req),
		ResponseCode: 201,
		ResponseBody: responseBody,
		ExpiresAt:    now.Add(u.idemTTL),
		CreatedAt:    now,
	}

	for range 2 {
		err := u.repo.CreateWithIdempotency(ctx, d, idem)
		if err == nil {
			u.writeAudit(ctx, d.ID, entity.AuditActionCreated, nil, d)
			return d, false, nil
		}
		if !errors.Is(err, entity.ErrIdempotencyKeyExists) {
			return nil, false, httpresponse.Internal("failed to create disbursement").WithErr(err)
		}

		existing, err := u.repo.FindIdempotencyKey(ctx, idempotencyKey)
		if err != nil {
			return nil, false, httpresponse.Internal("failed to check idempotency key").WithErr(err)
		}

		if now.After(existing.ExpiresAt) {
			if err := u.repo.DeleteIdempotencyKey(ctx, idempotencyKey); err != nil {
				return nil, false, httpresponse.Internal("failed to clean up expired idempotency key").WithErr(err)
			}
			continue
		}

		if existing.UserID != idem.UserID || existing.RequestHash != idem.RequestHash {
			return nil, false, httpresponse.Conflict("idempotency key was already used with a different request")
		}

		var cached entity.Disbursement
		if err := json.Unmarshal(existing.ResponseBody, &cached); err != nil {
			return nil, false, httpresponse.Internal("failed to read cached response").WithErr(err)
		}

		logger.WithCtx(ctx, u.log).Info("idempotent request replayed",
			zap.String("idempotency_key", idempotencyKey),
			zap.String("disbursement_id", cached.ID),
		)
		return &cached, true, nil
	}

	return nil, false, httpresponse.Conflict("idempotency key conflict, please retry")
}

// CreateBatch processes items independently: an invalid or failed item only
// fails that item, the rest still gets created (partial success).
func (u *disbursementUsecase) CreateBatch(ctx context.Context, req *entity.BatchCreateDisbursementRequest) ([]entity.BatchItemResult, error) {
	actor := reqctx.Username(ctx)
	results := make([]entity.BatchItemResult, 0, len(req.Items))

	for i := range req.Items {
		item := req.Items[i]

		if err := utils.ValidateStruct(&item); err != nil {
			results = append(results, entity.BatchItemResult{
				Index: i,
				Error: strings.Join(utils.FormatValidationError(err), ", "),
			})
			continue
		}

		now := time.Now().UTC()
		d := &entity.Disbursement{
			ID:            uuid.NewString(),
			RecipientName: item.RecipientName,
			AccountNumber: item.AccountNumber,
			BankCode:      item.BankCode,
			Amount:        item.Amount,
			AdminFee:      CalculateAdminFee(item.Amount),
			Note:          item.Note,
			Status:        entity.StatusPending,
			CreatedBy:     actor,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := u.repo.Create(ctx, d); err != nil {
			logger.WithCtx(ctx, u.log).Error("failed to create disbursement in batch",
				zap.Int("index", i),
				zap.Error(err),
			)
			results = append(results, entity.BatchItemResult{Index: i, Error: "failed to create disbursement"})
			continue
		}

		u.writeAudit(ctx, d.ID, entity.AuditActionCreated, nil, d)
		results = append(results, entity.BatchItemResult{Index: i, Success: true, Data: d})
	}

	return results, nil
}

func (u *disbursementUsecase) List(ctx context.Context, f *entity.DisbursementFilter) ([]entity.Disbursement, int64, error) {
	f.Normalize()

	if f.Status != "" && f.Status != entity.StatusPending && f.Status != entity.StatusApproved && f.Status != entity.StatusRejected {
		return nil, 0, httpresponse.BadRequest("status must be one of PENDING, APPROVED, REJECTED")
	}
	if err := validateDateRange(f.DateFrom, f.DateTo); err != nil {
		return nil, 0, err
	}

	rows, total, err := u.repo.FindAll(ctx, f)
	if err != nil {
		return nil, 0, httpresponse.Internal("failed to fetch disbursements").WithErr(err)
	}

	return rows, total, nil
}

func (u *disbursementUsecase) Detail(ctx context.Context, id string) (*entity.Disbursement, error) {
	d, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, httpresponse.NotFound("disbursement not found")
		}
		return nil, httpresponse.Internal("failed to fetch disbursement").WithErr(err)
	}

	return d, nil
}

func (u *disbursementUsecase) UpdateStatus(ctx context.Context, id string, req *entity.UpdateStatusRequest) (*entity.Disbursement, error) {
	if req.Status != entity.StatusApproved && req.Status != entity.StatusRejected {
		return nil, httpresponse.BadRequest("status must be APPROVED or REJECTED")
	}

	current, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, httpresponse.NotFound("disbursement not found")
		}
		return nil, httpresponse.Internal("failed to fetch disbursement").WithErr(err)
	}

	if current.Status != entity.StatusPending {
		return nil, httpresponse.Conflict(fmt.Sprintf("disbursement is already %s and can no longer be changed", current.Status))
	}

	actor := reqctx.Username(ctx)

	rows, err := u.repo.UpdateStatus(ctx, id, req.Status, req.Note, actor)
	if err != nil {
		return nil, httpresponse.Internal("failed to update disbursement status").WithErr(err)
	}
	if rows == 0 {

		return nil, httpresponse.Conflict("disbursement was just processed by another request, please re-check its status")
	}

	updated, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, httpresponse.Internal("failed to fetch updated disbursement").WithErr(err)
	}

	u.writeAudit(ctx, id, entity.AuditActionStatusChanged,
		map[string]any{"status": current.Status, "note": current.Note},
		map[string]any{"status": updated.Status, "note": updated.Note, "approved_by": actor},
	)

	logger.WithCtx(ctx, u.log).Info("disbursement status updated",
		zap.String("disbursement_id", id),
		zap.String("status", req.Status),
	)

	return updated, nil
}

func (u *disbursementUsecase) Delete(ctx context.Context, id string) error {
	current, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return httpresponse.NotFound("disbursement not found")
		}
		return httpresponse.Internal("failed to fetch disbursement").WithErr(err)
	}

	if current.Status != entity.StatusPending {
		return httpresponse.Conflict("only PENDING disbursement can be deleted")
	}

	rows, err := u.repo.SoftDelete(ctx, id)
	if err != nil {
		return httpresponse.Internal("failed to delete disbursement").WithErr(err)
	}
	if rows == 0 {
		return httpresponse.Conflict("disbursement was just processed by another request, please re-check its status")
	}

	u.writeAudit(ctx, id, entity.AuditActionDeleted,
		map[string]any{"status": current.Status},
		map[string]any{"deleted": true},
	)

	logger.WithCtx(ctx, u.log).Info("disbursement deleted", zap.String("disbursement_id", id))

	return nil
}

func (u *disbursementUsecase) writeAudit(ctx context.Context, entityID, action string, before, after any) {
	auditLog := &entity.AuditLog{
		ID:        uuid.NewString(),
		EntityID:  entityID,
		Action:    action,
		Actor:     reqctx.Username(ctx),
		CreatedAt: time.Now().UTC(),
	}

	if before != nil {
		auditLog.Before, _ = json.Marshal(before)
	}
	if after != nil {
		auditLog.After, _ = json.Marshal(after)
	}

	if err := u.auditRepo.Create(context.WithoutCancel(ctx), auditLog); err != nil {
		logger.WithCtx(ctx, u.log).Error("failed to write audit log",
			zap.String("entity_id", entityID),
			zap.String("action", action),
			zap.Error(err),
		)
	}
}

func validateDateRange(dateFrom, dateTo string) error {
	var from, to time.Time
	var err error

	if dateFrom != "" {
		if from, err = time.Parse("2006-01-02", dateFrom); err != nil {
			return httpresponse.BadRequest("date_from must be in YYYY-MM-DD format")
		}
	}
	if dateTo != "" {
		if to, err = time.Parse("2006-01-02", dateTo); err != nil {
			return httpresponse.BadRequest("date_to must be in YYYY-MM-DD format")
		}
	}
	if dateFrom != "" && dateTo != "" && from.After(to) {
		return httpresponse.BadRequest("date_from cannot be later than date_to")
	}

	return nil
}

func hashRequest(req *entity.CreateDisbursementRequest) string {
	payload, _ := json.Marshal(req)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
