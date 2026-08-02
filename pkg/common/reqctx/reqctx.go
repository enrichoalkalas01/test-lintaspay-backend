package reqctx

import "context"

// reqctx holds request scoped values that need to travel across layers
// (handler -> usecase -> repository) without passing extra params everywhere.

type ctxKey int

const (
	requestIDKey ctxKey = iota
	userIDKey
	usernameKey
	roleKey
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}

func WithUser(ctx context.Context, userID, username, role string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, usernameKey, username)
	return context.WithValue(ctx, roleKey, role)
}

func UserID(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}

func Username(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey).(string)
	return username
}

func Role(ctx context.Context) string {
	role, _ := ctx.Value(roleKey).(string)
	return role
}
