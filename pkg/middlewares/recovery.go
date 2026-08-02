package middlewares

import (
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func NewRecoveryMiddleware(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}

				// Log panic dengan stack trace
				log.Error("PANIC RECOVERED",
					zap.String("error", err.Error()),
					zap.String("stack", string(debug.Stack())),
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
				)

				// Return error response
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Internal Server Error",
					"message": "Something went wrong",
				})
			}
		}()

		return c.Next()
	}
}
