package middlewares

import (
	"errors"
	"time"

	"test-lintaspay/pkg/common/httpresponse"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// NewRequestLoggerMiddleware emits one structured log line per request.
func NewRequestLoggerMiddleware(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		// when the handler returns an error, the response has not been
		// written yet, so resolve the status code from the error itself
		status := c.Response().StatusCode()
		if err != nil {
			var httpErr *httpresponse.HTTPError
			var fiberErr *fiber.Error

			switch {
			case errors.As(err, &httpErr):
				status = httpErr.Code
			case errors.As(err, &fiberErr):
				status = fiberErr.Code
			default:
				status = fiber.StatusInternalServerError
			}
		}

		requestID, _ := c.Locals("request_id").(string)
		username, _ := c.Locals("username").(string)

		log.Info("request completed",
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status_code", status),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
			zap.String("user", username),
		)

		return err
	}
}
