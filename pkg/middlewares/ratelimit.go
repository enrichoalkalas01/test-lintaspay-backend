package middlewares

import (
	"time"

	"test-lintaspay/pkg/common/httpresponse"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// NewRateLimiter limits requests per user (falls back to IP when anonymous).
// Storage is in-memory, enough for single instance deployment.
func NewRateLimiter(max int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			if username, ok := c.Locals("username").(string); ok && username != "" {
				return username
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return httpresponse.NewHTTPError(fiber.StatusTooManyRequests, "too many requests, slow down")
		},
	})
}
