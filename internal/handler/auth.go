package handler

import (
	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/pkg/common/httpresponse"
	"test-lintaspay/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	usecase entity.AuthUsecase
}

func NewAuthHandler(usecase entity.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(entity.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return httpresponse.BadRequest("invalid request body")
	}
	if err := utils.ValidateStruct(req); err != nil {
		return httpresponse.BadRequest("validation failed").WithDetails(utils.FormatValidationError(err))
	}

	result, err := h.usecase.Login(c.UserContext(), req)
	if err != nil {
		return err
	}

	return httpresponse.OK(c, "login success", result)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	req := new(entity.RefreshRequest)
	if err := c.BodyParser(req); err != nil {
		return httpresponse.BadRequest("invalid request body")
	}
	if err := utils.ValidateStruct(req); err != nil {
		return httpresponse.BadRequest("validation failed").WithDetails(utils.FormatValidationError(err))
	}

	result, err := h.usecase.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		return err
	}

	return httpresponse.OK(c, "access token refreshed", result)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	req := new(entity.RefreshRequest)
	if err := c.BodyParser(req); err != nil {
		return httpresponse.BadRequest("invalid request body")
	}
	if err := utils.ValidateStruct(req); err != nil {
		return httpresponse.BadRequest("validation failed").WithDetails(utils.FormatValidationError(err))
	}

	if err := h.usecase.Logout(c.UserContext(), req.RefreshToken); err != nil {
		return err
	}

	return httpresponse.NewResponse(fiber.StatusOK, "logout success").JSON(c)
}
