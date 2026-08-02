package middlewares

import (
	"test-lintaspay/pkg/common/reqctx"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

func (m *RequestIDMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(fiber.HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set(fiber.HeaderXRequestID, requestID)
		c.Locals("request_id", requestID)

		// propagate into context so usecase/repository logs can pick it up
		c.SetUserContext(reqctx.WithRequestID(c.UserContext(), requestID))

		return c.Next()
	}
}
