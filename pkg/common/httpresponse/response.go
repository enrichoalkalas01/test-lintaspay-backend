package httpresponse

import "github.com/gofiber/fiber/v2"

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Response
	Error any `json:"error,omitempty"`
}

type DataResponse[T any] struct {
	Response
	Data T `json:"data"`
}

type Pagination struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Response
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// ============ Builders ============

func NewResponse(code int, message string) *Response {
	return &Response{Code: code, Message: message}
}

func NewDataResponse[T any](code int, message string, data T) *DataResponse[T] {
	return &DataResponse[T]{
		Response: Response{Code: code, Message: message},
		Data:     data,
	}
}

func NewPagination(total int64, page, perPage int) Pagination {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	if totalPages < 1 {
		totalPages = 1
	}

	return Pagination{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

func NewPaginatedResponse[T any](code int, message string, data []T, pagination Pagination) *PaginatedResponse[T] {
	if data == nil {
		data = []T{}
	}

	return &PaginatedResponse[T]{
		Response:   Response{Code: code, Message: message},
		Data:       data,
		Pagination: pagination,
	}
}

// ============ JSON Senders ============

func (r *Response) JSON(c *fiber.Ctx) error {
	return c.Status(r.Code).JSON(r)
}

func (r *ErrorResponse) JSON(c *fiber.Ctx) error {
	return c.Status(r.Code).JSON(r)
}

func (r *DataResponse[T]) JSON(c *fiber.Ctx) error {
	return c.Status(r.Code).JSON(r)
}

func (r *PaginatedResponse[T]) JSON(c *fiber.Ctx) error {
	return c.Status(r.Code).JSON(r)
}

// ============ Shortcuts ============

func OK[T any](c *fiber.Ctx, message string, data T) error {
	return NewDataResponse(fiber.StatusOK, message, data).JSON(c)
}

func Created[T any](c *fiber.Ctx, message string, data T) error {
	return NewDataResponse(fiber.StatusCreated, message, data).JSON(c)
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Paginated[T any](c *fiber.Ctx, message string, data []T, total int64, page, perPage int) error {
	return NewPaginatedResponse(fiber.StatusOK, message, data, NewPagination(total, page, perPage)).JSON(c)
}

func GetOffset(page, perPage int) int {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	return (page - 1) * perPage
}
