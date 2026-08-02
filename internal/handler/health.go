package handler

import (
	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/pkg/common/httpresponse"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	usecase entity.HealthUsecase
}

func NewHealthHandler(usecase entity.HealthUsecase) *HealthHandler {
	return &HealthHandler{usecase: usecase}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	status := h.usecase.Status(c.UserContext())

	if status.Status != "ok" {
		return httpresponse.ServiceUnavailable("service degraded").
			WithDetails(fiber.Map{"database": status.Database})
	}

	return httpresponse.OK(c, "service healthy", status)
}
