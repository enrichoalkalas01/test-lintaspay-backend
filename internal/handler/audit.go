package handler

import (
	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/pkg/common/httpresponse"

	"github.com/gofiber/fiber/v2"
)

type AuditLogHandler struct {
	usecase entity.AuditLogUsecase
}

func NewAuditLogHandler(usecase entity.AuditLogUsecase) *AuditLogHandler {
	return &AuditLogHandler{usecase: usecase}
}

func (h *AuditLogHandler) List(c *fiber.Ctx) error {
	filter := new(entity.AuditLogFilter)
	if err := c.QueryParser(filter); err != nil {
		return httpresponse.BadRequest("invalid query parameters")
	}

	rows, total, err := h.usecase.List(c.UserContext(), filter)
	if err != nil {
		return err
	}

	return httpresponse.Paginated(c, "audit logs retrieved", rows, total, filter.Page, filter.Limit)
}
