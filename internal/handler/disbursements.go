package handler

import (
	"fmt"

	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/pkg/common/httpresponse"
	"test-lintaspay/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type DisbursementHandler struct {
	usecase entity.DisbursementUsecase
}

func NewDisbursementHandler(usecase entity.DisbursementUsecase) *DisbursementHandler {
	return &DisbursementHandler{usecase: usecase}
}

func (h *DisbursementHandler) List(c *fiber.Ctx) error {
	filter := new(entity.DisbursementFilter)
	if err := c.QueryParser(filter); err != nil {
		return httpresponse.BadRequest("invalid query parameters")
	}

	rows, total, err := h.usecase.List(c.UserContext(), filter)
	if err != nil {
		return err
	}

	return httpresponse.Paginated(c, "disbursements retrieved", rows, total, filter.Page, filter.Limit)
}

func (h *DisbursementHandler) Detail(c *fiber.Ctx) error {
	result, err := h.usecase.Detail(c.UserContext(), c.Params("id"))
	if err != nil {
		return err
	}

	return httpresponse.OK(c, "disbursement retrieved", result)
}

func (h *DisbursementHandler) Create(c *fiber.Ctx) error {
	req := new(entity.CreateDisbursementRequest)
	if err := c.BodyParser(req); err != nil {
		return httpresponse.BadRequest("invalid request body")
	}
	if err := utils.ValidateStruct(req); err != nil {
		return httpresponse.BadRequest("validation failed").WithDetails(utils.FormatValidationError(err))
	}

	result, replayed, err := h.usecase.Create(c.UserContext(), req, c.Get("Idempotency-Key"))
	if err != nil {
		return err
	}

	if replayed {
		c.Set("X-Idempotent-Replayed", "true")
	}

	return httpresponse.Created(c, "disbursement created", result)
}

func (h *DisbursementHandler) CreateBatch(c *fiber.Ctx) error {
	req := new(entity.BatchCreateDisbursementRequest)
	if err := c.BodyParser(req); err != nil {
		return httpresponse.BadRequest("invalid request body")
	}
	if err := utils.ValidateStruct(req); err != nil {
		return httpresponse.BadRequest("validation failed").WithDetails(utils.FormatValidationError(err))
	}

	results, err := h.usecase.CreateBatch(c.UserContext(), req)
	if err != nil {
		return err
	}

	succeeded := 0
	for _, r := range results {
		if r.Success {
			succeeded++
		}
	}

	code := fiber.StatusCreated
	switch {
	case succeeded == 0:
		code = fiber.StatusBadRequest
	case succeeded < len(results):
		code = fiber.StatusMultiStatus
	}

	message := fmt.Sprintf("%d of %d disbursements created", succeeded, len(results))
	return httpresponse.NewDataResponse(code, message, results).JSON(c)
}

func (h *DisbursementHandler) UpdateStatus(c *fiber.Ctx) error {
	req := new(entity.UpdateStatusRequest)
	if err := c.BodyParser(req); err != nil {
		return httpresponse.BadRequest("invalid request body")
	}
	if err := utils.ValidateStruct(req); err != nil {
		return httpresponse.BadRequest("validation failed").WithDetails(utils.FormatValidationError(err))
	}

	result, err := h.usecase.UpdateStatus(c.UserContext(), c.Params("id"), req)
	if err != nil {
		return err
	}

	return httpresponse.OK(c, "disbursement status updated", result)
}

func (h *DisbursementHandler) Delete(c *fiber.Ctx) error {
	if err := h.usecase.Delete(c.UserContext(), c.Params("id")); err != nil {
		return err
	}

	return httpresponse.NewResponse(fiber.StatusOK, "disbursement deleted").JSON(c)
}
