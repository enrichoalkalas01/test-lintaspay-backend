package entity

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

const (
	AuditActionCreated       = "created"
	AuditActionStatusChanged = "status_changed"
	AuditActionDeleted       = "deleted"
)

type AuditLog struct {
	ID        string          `json:"id" gorm:"primaryKey;size:36"`
	EntityID  string          `json:"entity_id" gorm:"size:36;index"`
	Action    string          `json:"action" gorm:"size:30;index"`
	Actor     string          `json:"actor" gorm:"size:50"`
	Before    json.RawMessage `json:"before" gorm:"type:json"`
	After     json.RawMessage `json:"after" gorm:"type:json"`
	CreatedAt time.Time       `json:"created_at" gorm:"index"`
}

type AuditLogFilter struct {
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
	EntityID string `query:"entity_id"`
	Action   string `query:"action"`
	DateFrom string `query:"date_from"`
	DateTo   string `query:"date_to"`
}

func (f *AuditLogFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}

	f.EntityID = strings.TrimSpace(f.EntityID)
	f.Action = strings.TrimSpace(f.Action)
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	FindAll(ctx context.Context, f *AuditLogFilter) ([]AuditLog, int64, error)
}

type AuditLogUsecase interface {
	List(ctx context.Context, f *AuditLogFilter) ([]AuditLog, int64, error)
}
