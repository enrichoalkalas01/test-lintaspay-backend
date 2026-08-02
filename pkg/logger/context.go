package logger

import (
	"context"

	"test-lintaspay/pkg/common/reqctx"

	"go.uber.org/zap"
)

// WithCtx attaches request scoped fields (request_id, user) to the logger
// so every log line inside one request can be correlated.
func WithCtx(ctx context.Context, log *zap.Logger) *zap.Logger {
	fields := make([]zap.Field, 0, 2)

	if requestID := reqctx.RequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}
	if username := reqctx.Username(ctx); username != "" {
		fields = append(fields, zap.String("user", username))
	}

	if len(fields) == 0 {
		return log
	}
	return log.With(fields...)
}
