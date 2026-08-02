package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct melakukan validasi terhadap struct
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// FormatValidationError memformat error validasi menjadi lebih readable
func FormatValidationError(err error) []string {
	var errors []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, formatFieldError(e))
		}
	}

	return errors
}

func formatFieldError(e validator.FieldError) string {
	field := strings.ToLower(e.Field())

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
