package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/pkg/common/httpresponse"
	"test-lintaspay/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	repo       entity.AuthRepository
	log        *zap.Logger
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthUsecase(env *viper.Viper, repo entity.AuthRepository, log *zap.Logger) entity.AuthUsecase {
	accessTTL := env.GetDuration("JWT_ACCESS_TTL")
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}

	refreshTTL := env.GetDuration("JWT_REFRESH_TTL")
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour
	}

	return &authUsecase{
		repo:       repo,
		log:        log,
		secret:     []byte(env.GetString("JWT_SECRET")),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (u *authUsecase) Login(ctx context.Context, req *entity.LoginRequest) (*entity.TokenResponse, error) {
	user, err := u.repo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, httpresponse.Unauthorized("invalid username or password")
		}
		return nil, httpresponse.Internal("failed to fetch user").WithErr(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, httpresponse.Unauthorized("invalid username or password")
	}

	accessToken, err := u.generateAccessToken(user)
	if err != nil {
		return nil, httpresponse.Internal("failed to generate access token").WithErr(err)
	}

	refreshToken, err := u.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, httpresponse.Internal("failed to generate refresh token").WithErr(err)
	}

	logger.WithCtx(ctx, u.log).Info("user logged in", zap.String("username", user.Username))

	return &entity.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.accessTTL.Seconds()),
	}, nil
}

func (u *authUsecase) Refresh(ctx context.Context, refreshToken string) (*entity.TokenResponse, error) {
	stored, err := u.repo.FindRefreshToken(ctx, hashToken(refreshToken))
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, httpresponse.Unauthorized("invalid refresh token")
		}
		return nil, httpresponse.Internal("failed to fetch refresh token").WithErr(err)
	}

	if stored.RevokedAt != nil {
		return nil, httpresponse.Unauthorized("refresh token has been revoked")
	}
	if time.Now().After(stored.ExpiresAt) {
		return nil, httpresponse.Unauthorized("refresh token has expired")
	}

	user, err := u.repo.FindUserByID(ctx, stored.UserID)
	if err != nil {
		return nil, httpresponse.Internal("failed to fetch user").WithErr(err)
	}

	accessToken, err := u.generateAccessToken(user)
	if err != nil {
		return nil, httpresponse.Internal("failed to generate access token").WithErr(err)
	}

	return &entity.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(u.accessTTL.Seconds()),
	}, nil
}

func (u *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)

	if _, err := u.repo.FindRefreshToken(ctx, tokenHash); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return httpresponse.Unauthorized("invalid refresh token")
		}
		return httpresponse.Internal("failed to fetch refresh token").WithErr(err)
	}

	if err := u.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return httpresponse.Internal("failed to revoke refresh token").WithErr(err)
	}

	logger.WithCtx(ctx, u.log).Info("refresh token revoked")

	return nil
}

func (u *authUsecase) generateAccessToken(user *entity.User) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"iat":      now.Unix(),
		"exp":      now.Add(u.accessTTL).Unix(),
		"jti":      uuid.NewString(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(u.secret)
}

func (u *authUsecase) issueRefreshToken(ctx context.Context, userID string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	tokenValue := hex.EncodeToString(raw)

	now := time.Now().UTC()
	err := u.repo.CreateRefreshToken(ctx, &entity.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hashToken(tokenValue),
		ExpiresAt: now.Add(u.refreshTTL),
		CreatedAt: now,
	})
	if err != nil {
		return "", err
	}

	return tokenValue, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
