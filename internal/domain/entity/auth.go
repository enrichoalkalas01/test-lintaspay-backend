package entity

import (
	"context"
	"time"
)

const (
	RoleOperator   = "operator"
	RoleAdmin      = "admin"
	RoleSuperadmin = "superadmin"
)

type User struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	Username     string    `json:"username" gorm:"size:50;uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"size:100"`
	Role         string    `json:"role" gorm:"size:20"`
	CreatedAt    time.Time `json:"created_at"`
}

// RefreshToken is stored hashed (sha256), never the raw value.
type RefreshToken struct {
	ID        string `gorm:"primaryKey;size:36"`
	UserID    string `gorm:"size:36;index"`
	TokenHash string `gorm:"size:64;uniqueIndex"`
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AuthRepository interface {
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}

type AuthUsecase interface {
	Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}
