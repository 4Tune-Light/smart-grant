package validator

import (
	"errors"
	"strings"

	goValidator "github.com/go-playground/validator/v10"
)

var validate *goValidator.Validate

func init() {
	validate = goValidator.New()
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Struct(s interface{}) []FieldError {
	if err := validate.Struct(s); err != nil {
		var ve goValidator.ValidationErrors
		if errors.As(err, &ve) {
			fields := make([]FieldError, len(ve))
			for i, fe := range ve {
				fields[i] = FieldError{
					Field: toSnake(fe.Field()),
					Message: messageForTag(fe.Tag(), fe.Param()),
				}
			}
			return fields
		}
	}
	return nil
}

func messageForTag(tag string, param string) string {
	switch tag {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + param + " characters"
	case "max":
		return "must not exceed " + param + " characters"
	case "oneof":
		return "must be one of: " + param
	case "gt":
		return "must be greater than " + param
	case "gte":
		return "must be greater than or equal to " + param
	case "lt":
		return "must be less than " + param
	case "lte":
		return "must be less than or equal to " + param
	default:
		return "validation failed on " + tag
	}
}

func toSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
