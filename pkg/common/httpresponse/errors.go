package httpresponse

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// HTTPError is returned by handlers/usecases and rendered uniformly
// by the global error handler.
type HTTPError struct {
	Code    int
	Message string
	Details any
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}

func (e *HTTPError) WithDetails(details any) *HTTPError {
	e.Details = details
	return e
}

func (e *HTTPError) WithErr(err error) *HTTPError {
	e.Err = err
	return e
}

// Common error shortcuts
func BadRequest(message string) *HTTPError {
	return NewHTTPError(fiber.StatusBadRequest, message)
}

func Unauthorized(message string) *HTTPError {
	return NewHTTPError(fiber.StatusUnauthorized, message)
}

func Forbidden(message string) *HTTPError {
	return NewHTTPError(fiber.StatusForbidden, message)
}

func NotFound(message string) *HTTPError {
	return NewHTTPError(fiber.StatusNotFound, message)
}

func Conflict(message string) *HTTPError {
	return NewHTTPError(fiber.StatusConflict, message)
}

func UnprocessableEntity(message string) *HTTPError {
	return NewHTTPError(fiber.StatusUnprocessableEntity, message)
}

func Internal(message string) *HTTPError {
	return NewHTTPError(fiber.StatusInternalServerError, message)
}

func ServiceUnavailable(message string) *HTTPError {
	return NewHTTPError(fiber.StatusServiceUnavailable, message)
}

// NewErrorHandler renders every error returned from any handler into a
// uniform ErrorResponse JSON, including fiber's own errors (404, 405, ...).
func NewErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		resp := &ErrorResponse{
			Response: Response{
				Code:    fiber.StatusInternalServerError,
				Message: "internal server error",
			},
		}

		var httpErr *HTTPError
		var fiberErr *fiber.Error

		switch {
		case errors.As(err, &httpErr):
			resp.Code = httpErr.Code
			resp.Message = httpErr.Message
			resp.Error = httpErr.Details
			// never leak internal error strings on 5xx
			if resp.Error == nil && httpErr.Err != nil && httpErr.Code < 500 {
				resp.Error = httpErr.Err.Error()
			}
		case errors.As(err, &fiberErr):
			resp.Code = fiberErr.Code
			resp.Message = fiberErr.Message
		}

		if resp.Code >= 500 {
			log.Error("request failed",
				zap.Int("code", resp.Code),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Error(err),
			)
		}

		return resp.JSON(c)
	}
}
