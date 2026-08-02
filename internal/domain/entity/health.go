package entity

import "context"

type HealthStatus struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

type HealthRepository interface {
	PingDB(ctx context.Context) error
}

type HealthUsecase interface {
	Status(ctx context.Context) *HealthStatus
}
