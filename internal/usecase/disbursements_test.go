package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/pkg/common/httpresponse"
	"test-lintaspay/pkg/common/reqctx"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type fakeDisbursementRepo struct {
	disbursements map[string]*entity.Disbursement
	idemKeys      map[string]*entity.IdempotencyKey
	forceZeroRows bool
}

func newFakeDisbursementRepo() *fakeDisbursementRepo {
	return &fakeDisbursementRepo{
		disbursements: map[string]*entity.Disbursement{},
		idemKeys:      map[string]*entity.IdempotencyKey{},
	}
}

func (f *fakeDisbursementRepo) Create(ctx context.Context, d *entity.Disbursement) error {
	f.disbursements[d.ID] = d
	return nil
}

func (f *fakeDisbursementRepo) CreateWithIdempotency(ctx context.Context, d *entity.Disbursement, idem *entity.IdempotencyKey) error {
	if _, exists := f.idemKeys[idem.Key]; exists {
		return entity.ErrIdempotencyKeyExists
	}
	f.idemKeys[idem.Key] = idem
	f.disbursements[d.ID] = d
	return nil
}

func (f *fakeDisbursementRepo) FindByID(ctx context.Context, id string) (*entity.Disbursement, error) {
	d, ok := f.disbursements[id]
	if !ok {
		return nil, entity.ErrNotFound
	}
	copied := *d
	return &copied, nil
}

func (f *fakeDisbursementRepo) FindAll(ctx context.Context, filter *entity.DisbursementFilter) ([]entity.Disbursement, int64, error) {
	var rows []entity.Disbursement
	for _, d := range f.disbursements {
		rows = append(rows, *d)
	}
	return rows, int64(len(rows)), nil
}

func (f *fakeDisbursementRepo) UpdateStatus(ctx context.Context, id, status, note, actor string) (int64, error) {
	if f.forceZeroRows {
		return 0, nil
	}

	d, ok := f.disbursements[id]
	if !ok || d.Status != entity.StatusPending {
		return 0, nil
	}

	d.Status = status
	if note != "" {
		d.Note = note
	}
	d.ApprovedBy = &actor
	return 1, nil
}

func (f *fakeDisbursementRepo) SoftDelete(ctx context.Context, id string) (int64, error) {
	d, ok := f.disbursements[id]
	if !ok || d.Status != entity.StatusPending {
		return 0, nil
	}
	delete(f.disbursements, id)
	return 1, nil
}

func (f *fakeDisbursementRepo) FindIdempotencyKey(ctx context.Context, key string) (*entity.IdempotencyKey, error) {
	idem, ok := f.idemKeys[key]
	if !ok {
		return nil, entity.ErrNotFound
	}
	return idem, nil
}

func (f *fakeDisbursementRepo) DeleteIdempotencyKey(ctx context.Context, key string) error {
	delete(f.idemKeys, key)
	return nil
}

type fakeAuditRepo struct {
	logs []*entity.AuditLog
	err  error
}

func (f *fakeAuditRepo) Create(ctx context.Context, log *entity.AuditLog) error {
	if f.err != nil {
		return f.err
	}
	f.logs = append(f.logs, log)
	return nil
}

func (f *fakeAuditRepo) FindAll(ctx context.Context, filter *entity.AuditLogFilter) ([]entity.AuditLog, int64, error) {
	var rows []entity.AuditLog
	for _, l := range f.logs {
		rows = append(rows, *l)
	}
	return rows, int64(len(rows)), nil
}

func newTestUsecase(repo *fakeDisbursementRepo, audit *fakeAuditRepo) entity.DisbursementUsecase {
	return NewDisbursementUsecase(viper.New(), repo, audit, zap.NewNop())
}

func testCtx() context.Context {
	ctx := reqctx.WithRequestID(context.Background(), "test-request-id")
	return reqctx.WithUser(ctx, "user-1", "admin", "admin")
}

func createRequest() *entity.CreateDisbursementRequest {
	return &entity.CreateDisbursementRequest{
		RecipientName: "Budi Santoso",
		AccountNumber: "1234567890",
		BankCode:      "BCA",
		Amount:        1_250_000,
		Note:          "Pembayaran supplier",
	}
}

func httpCode(t *testing.T, err error) int {
	t.Helper()

	var httpErr *httpresponse.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	return httpErr.Code
}

func TestCalculateAdminFee(t *testing.T) {
	tests := []struct {
		amount int64
		want   int64
	}{
		{10_000, 2500},
		{4_999_999, 2500},
		{5_000_000, 5000},
		{12_345_678, 5000},
	}

	for _, tt := range tests {
		if got := CalculateAdminFee(tt.amount); got != tt.want {
			t.Errorf("CalculateAdminFee(%d) = %d, want %d", tt.amount, got, tt.want)
		}
	}
}

func TestCreateWithoutIdempotencyKey(t *testing.T) {
	repo := newFakeDisbursementRepo()
	audit := &fakeAuditRepo{}
	uc := newTestUsecase(repo, audit)

	d, replayed, err := uc.Create(testCtx(), createRequest(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if replayed {
		t.Error("fresh create should not be marked as replayed")
	}
	if d.Status != entity.StatusPending {
		t.Errorf("status = %s, want PENDING", d.Status)
	}
	if d.AdminFee != 2500 {
		t.Errorf("admin_fee = %d, want 2500", d.AdminFee)
	}
	if d.CreatedBy != "admin" {
		t.Errorf("created_by = %s, want admin (from context)", d.CreatedBy)
	}
	if len(audit.logs) != 1 || audit.logs[0].Action != entity.AuditActionCreated {
		t.Error("expected one 'created' audit log entry")
	}
}

func TestCreateIdempotentReplay(t *testing.T) {
	repo := newFakeDisbursementRepo()
	uc := newTestUsecase(repo, &fakeAuditRepo{})
	key := uuid.NewString()

	first, replayed, err := uc.Create(testCtx(), createRequest(), key)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if replayed {
		t.Fatal("first request must not be a replay")
	}

	second, replayed, err := uc.Create(testCtx(), createRequest(), key)
	if err != nil {
		t.Fatalf("second create failed: %v", err)
	}
	if !replayed {
		t.Error("second request with same key must be a replay")
	}
	if second.ID != first.ID {
		t.Errorf("replayed response has different id: %s vs %s", second.ID, first.ID)
	}
	if len(repo.disbursements) != 1 {
		t.Errorf("disbursement created twice, count = %d", len(repo.disbursements))
	}
}

func TestCreateIdempotencyKeyReusedWithDifferentPayload(t *testing.T) {
	repo := newFakeDisbursementRepo()
	uc := newTestUsecase(repo, &fakeAuditRepo{})
	key := uuid.NewString()

	if _, _, err := uc.Create(testCtx(), createRequest(), key); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	other := createRequest()
	other.Amount = 7_000_000

	_, _, err := uc.Create(testCtx(), other, key)
	if err == nil {
		t.Fatal("expected conflict error for reused key with different payload")
	}
	if code := httpCode(t, err); code != 409 {
		t.Errorf("status code = %d, want 409", code)
	}
}

func TestCreateIdempotencyKeyExpired(t *testing.T) {
	repo := newFakeDisbursementRepo()
	uc := newTestUsecase(repo, &fakeAuditRepo{})
	key := uuid.NewString()

	if _, _, err := uc.Create(testCtx(), createRequest(), key); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	repo.idemKeys[key].ExpiresAt = time.Now().UTC().Add(-time.Hour)

	d, replayed, err := uc.Create(testCtx(), createRequest(), key)
	if err != nil {
		t.Fatalf("create after expiry failed: %v", err)
	}
	if replayed {
		t.Error("expired key must be processed as a new request, not replayed")
	}
	if d == nil || len(repo.disbursements) != 2 {
		t.Errorf("expected a second disbursement, count = %d", len(repo.disbursements))
	}
}

func TestCreateInvalidIdempotencyKeyFormat(t *testing.T) {
	uc := newTestUsecase(newFakeDisbursementRepo(), &fakeAuditRepo{})

	_, _, err := uc.Create(testCtx(), createRequest(), "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for malformed idempotency key")
	}
	if code := httpCode(t, err); code != 400 {
		t.Errorf("status code = %d, want 400", code)
	}
}

func TestCreateAuditFailureDoesNotBlock(t *testing.T) {
	repo := newFakeDisbursementRepo()
	audit := &fakeAuditRepo{err: errors.New("audit table is on fire")}
	uc := newTestUsecase(repo, audit)

	if _, _, err := uc.Create(testCtx(), createRequest(), ""); err != nil {
		t.Fatalf("create must succeed even when audit log write fails: %v", err)
	}
}

func seedPending(repo *fakeDisbursementRepo) *entity.Disbursement {
	d := &entity.Disbursement{
		ID:            uuid.NewString(),
		RecipientName: "Budi Santoso",
		AccountNumber: "1234567890",
		BankCode:      "BCA",
		Amount:        1_250_000,
		AdminFee:      2500,
		Status:        entity.StatusPending,
		CreatedBy:     "operator",
	}
	repo.disbursements[d.ID] = d
	return d
}

func TestUpdateStatusSuccess(t *testing.T) {
	repo := newFakeDisbursementRepo()
	audit := &fakeAuditRepo{}
	uc := newTestUsecase(repo, audit)
	d := seedPending(repo)

	updated, err := uc.UpdateStatus(testCtx(), d.ID, &entity.UpdateStatusRequest{
		Status: entity.StatusApproved,
		Note:   "Sudah diverifikasi",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entity.StatusApproved {
		t.Errorf("status = %s, want APPROVED", updated.Status)
	}
	if updated.ApprovedBy == nil || *updated.ApprovedBy != "admin" {
		t.Error("approved_by must be filled from the JWT context")
	}
	if len(audit.logs) != 1 || audit.logs[0].Action != entity.AuditActionStatusChanged {
		t.Error("expected one 'status_changed' audit log entry")
	}
}

func TestUpdateStatusAlreadyProcessed(t *testing.T) {
	repo := newFakeDisbursementRepo()
	uc := newTestUsecase(repo, &fakeAuditRepo{})
	d := seedPending(repo)
	repo.disbursements[d.ID].Status = entity.StatusApproved

	_, err := uc.UpdateStatus(testCtx(), d.ID, &entity.UpdateStatusRequest{Status: entity.StatusRejected})
	if err == nil {
		t.Fatal("expected conflict for already processed disbursement")
	}
	if code := httpCode(t, err); code != 409 {
		t.Errorf("status code = %d, want 409", code)
	}
	if !strings.Contains(err.Error(), "APPROVED") {
		t.Errorf("error message should mention current status, got: %v", err)
	}
}

func TestUpdateStatusLostRace(t *testing.T) {

	repo := newFakeDisbursementRepo()
	uc := newTestUsecase(repo, &fakeAuditRepo{})
	d := seedPending(repo)
	repo.forceZeroRows = true

	_, err := uc.UpdateStatus(testCtx(), d.ID, &entity.UpdateStatusRequest{Status: entity.StatusApproved})
	if err == nil {
		t.Fatal("expected conflict when losing the concurrent update race")
	}
	if code := httpCode(t, err); code != 409 {
		t.Errorf("status code = %d, want 409", code)
	}
}

func TestUpdateStatusInvalidStatus(t *testing.T) {
	repo := newFakeDisbursementRepo()
	uc := newTestUsecase(repo, &fakeAuditRepo{})
	d := seedPending(repo)

	_, err := uc.UpdateStatus(testCtx(), d.ID, &entity.UpdateStatusRequest{Status: "PENDING"})
	if err == nil {
		t.Fatal("expected error for invalid target status")
	}
	if code := httpCode(t, err); code != 400 {
		t.Errorf("status code = %d, want 400", code)
	}
}

func TestUpdateStatusNotFound(t *testing.T) {
	uc := newTestUsecase(newFakeDisbursementRepo(), &fakeAuditRepo{})

	_, err := uc.UpdateStatus(testCtx(), uuid.NewString(), &entity.UpdateStatusRequest{Status: entity.StatusApproved})
	if err == nil {
		t.Fatal("expected not found error")
	}
	if code := httpCode(t, err); code != 404 {
		t.Errorf("status code = %d, want 404", code)
	}
}

func TestDeleteNonPending(t *testing.T) {
	repo := newFakeDisbursementRepo()
	uc := newTestUsecase(repo, &fakeAuditRepo{})
	d := seedPending(repo)
	repo.disbursements[d.ID].Status = entity.StatusApproved

	err := uc.Delete(testCtx(), d.ID)
	if err == nil {
		t.Fatal("expected conflict when deleting non PENDING disbursement")
	}
	if code := httpCode(t, err); code != 409 {
		t.Errorf("status code = %d, want 409", code)
	}
}
