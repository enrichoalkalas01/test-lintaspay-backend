package entity

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	StatusPending  = "PENDING"
	StatusApproved = "APPROVED"
	StatusRejected = "REJECTED"
)

type Disbursement struct {
	ID            string         `json:"id" gorm:"primaryKey;size:36"`
	RecipientName string         `json:"recipient_name" gorm:"size:255;index"`
	AccountNumber string         `json:"account_number" gorm:"size:50"`
	BankCode      string         `json:"bank_code" gorm:"size:20"`
	Amount        int64          `json:"amount"`
	AdminFee      int64          `json:"admin_fee"`
	Note          string         `json:"note" gorm:"size:500"`
	Status        string         `json:"status" gorm:"size:20;index"`
	CreatedBy     string         `json:"created_by" gorm:"size:50"`
	ApprovedBy    *string        `json:"approved_by" gorm:"size:50"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type IdempotencyKey struct {
	Key          string `gorm:"primaryKey;size:36"`
	UserID       string `gorm:"size:36"`
	RequestHash  string `gorm:"size:64"`
	ResponseCode int
	ResponseBody json.RawMessage `gorm:"type:json"`
	ExpiresAt    time.Time       `gorm:"index"`
	CreatedAt    time.Time
}

type CreateDisbursementRequest struct {
	RecipientName string `json:"recipient_name" validate:"required"`
	AccountNumber string `json:"account_number" validate:"required"`
	BankCode      string `json:"bank_code" validate:"required"`
	Amount        int64  `json:"amount" validate:"required,min=10000"`
	Note          string `json:"note"`
}

type BatchCreateDisbursementRequest struct {
	// items are validated one by one in the usecase so a single bad item
	// does not fail the whole batch (partial success)
	Items []CreateDisbursementRequest `json:"items" validate:"required,min=1,max=100"`
}

type BatchItemResult struct {
	Index   int           `json:"index"`
	Success bool          `json:"success"`
	Data    *Disbursement `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=APPROVED REJECTED"`
	Note   string `json:"note"`
}

type DisbursementFilter struct {
	Page      int    `query:"page"`
	Limit     int    `query:"limit"`
	Search    string `query:"search"`
	Status    string `query:"status"`
	DateFrom  string `query:"date_from"`
	DateTo    string `query:"date_to"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
}

func (f *DisbursementFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}

	if f.SortBy != "amount" {
		f.SortBy = "created_at"
	}

	f.SortOrder = strings.ToLower(strings.TrimSpace(f.SortOrder))
	if f.SortOrder != "asc" {
		f.SortOrder = "desc"
	}

	f.Search = strings.TrimSpace(f.Search)
	f.Status = strings.ToUpper(strings.TrimSpace(f.Status))
}

type DisbursementRepository interface {
	Create(ctx context.Context, d *Disbursement) error
	CreateWithIdempotency(ctx context.Context, d *Disbursement, idem *IdempotencyKey) error
	FindByID(ctx context.Context, id string) (*Disbursement, error)
	FindAll(ctx context.Context, f *DisbursementFilter) ([]Disbursement, int64, error)
	UpdateStatus(ctx context.Context, id, status, note, actor string) (int64, error)
	SoftDelete(ctx context.Context, id string) (int64, error)
	FindIdempotencyKey(ctx context.Context, key string) (*IdempotencyKey, error)
	DeleteIdempotencyKey(ctx context.Context, key string) error
}

type DisbursementUsecase interface {
	Create(ctx context.Context, req *CreateDisbursementRequest, idempotencyKey string) (*Disbursement, bool, error)
	CreateBatch(ctx context.Context, req *BatchCreateDisbursementRequest) ([]BatchItemResult, error)
	List(ctx context.Context, f *DisbursementFilter) ([]Disbursement, int64, error)
	Detail(ctx context.Context, id string) (*Disbursement, error)
	UpdateStatus(ctx context.Context, id string, req *UpdateStatusRequest) (*Disbursement, error)
	Delete(ctx context.Context, id string) error
}
