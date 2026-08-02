package response

import "time"

// ============ Response Structure ============

type ResponseStatus string

const (
	StatusSuccess ResponseStatus = "success"
	StatusError   ResponseStatus = "error"
	StatusFailed  ResponseStatus = "failed"
)

// Base structure response
type BaseResponse[T any] struct {
	Status    ResponseStatus `json:"status"`
	Message   string         `json:"message"`
	Data      T              `json:"data"`
	Timestamp time.Time      `json:"timestamp"`
	Path      string         `json:"path,omitempty"`
	ErrorCode string         `json:"error_code,omitempty"`
}

// Paginated response for list response
type PaginatedResponse[T any] struct {
	Status     ResponseStatus `json:"status"`
	Message    string         `json:"message"`
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	Timestamp  time.Time      `json:"timestamp"`
	Path       string         `json:"path,omitempty"`
}

// Metadata for pagination
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// Error response for error handler
type ErrorResponse struct {
	Status    ResponseStatus `json:"status"`
	Message   string         `json:"message"`
	ErrorCode string         `json:"error_code"`
	Details   map[string]any `json:"details,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Path      string         `json:"path,omitempty"`
}

// ============ Response Builders ============
func NewSuccessResponse[T any](data T, message string) *BaseResponse[T] {
	return &BaseResponse[T]{
		Status:    StatusSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func NewSuccessResponseWithPath[T any](data T, message, path string) *BaseResponse[T] {
	return &BaseResponse[T]{
		Status:    StatusSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		Path:      path,
	}
}

func NewPaginatedResponse[T any](data []T, message string, pagination PaginationMeta) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Status:     StatusSuccess,
		Message:    message,
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now(),
	}
}

func NewPaginatedResponseWithPath[T any](
	data []T,
	message, path string,
	pagination PaginationMeta,
) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Status:     StatusSuccess,
		Message:    message,
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now(),
		Path:       path,
	}
}

func NewErrorResponse(message, errorCode string) *ErrorResponse {
	return &ErrorResponse{
		Status:    StatusError,
		Message:   message,
		ErrorCode: errorCode,
		Timestamp: time.Now(),
	}
}

func NewErrorResponseWithDetails(
	message, errorCode string,
	details map[string]any,
) *ErrorResponse {
	return &ErrorResponse{
		Status:    StatusError,
		Message:   message,
		ErrorCode: errorCode,
		Details:   details,
		Timestamp: time.Now(),
	}
}

func NewErrorResponseWithPath(
	message, errorCode, path string,
) *ErrorResponse {
	return &ErrorResponse{
		Status:    StatusError,
		Message:   message,
		ErrorCode: errorCode,
		Timestamp: time.Now(),
		Path:      path,
	}
}

// ============ Pagination Helper ============
func CalculatePaginationMeta(page, limit int, total int64) PaginationMeta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

func GetOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return (page - 1) * limit
}
