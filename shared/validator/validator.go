package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/microservices-go/shared/errors"
)

// Validator wraps go-playground validator
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom validations here if needed

	return &Validator{validate: v}
}

// ValidateStruct validates a struct
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(validationErrors)
		}
		return errors.Wrap(err, errors.ErrValidationFailed, "Validation failed")
	}
	return nil
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	if err := v.validate.Var(field, tag); err != nil {
		return errors.Wrap(err, errors.ErrValidationFailed, "Validation failed")
	}
	return nil
}

// formatValidationErrors converts validator errors to AppError
func formatValidationErrors(errs validator.ValidationErrors) error {
	details := make(map[string]string)
	for _, err := range errs {
		field := err.Field()
		tag := err.Tag()
		param := err.Param()

		msg := getErrorMessage(tag, param)
		details[field] = msg
	}

	return errors.New(errors.ErrValidationFailed, "Validation failed").WithDetails(formatDetails(details))
}

func getErrorMessage(tag, param string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short, minimum " + param
	case "max":
		return "Value is too long, maximum " + param
	case "gte":
		return "Value must be greater than or equal to " + param
	case "lte":
		return "Value must be less than or equal to " + param
	case "gt":
		return "Value must be greater than " + param
	case "lt":
		return "Value must be less than " + param
	case "len":
		return "Value must have length of " + param
	case "alphanum":
		return "Value must be alphanumeric"
	case "numeric":
		return "Value must be numeric"
	case "uuid":
		return "Invalid UUID format"
	case "url":
		return "Invalid URL format"
	default:
		return "Invalid value"
	}
}

func formatDetails(details map[string]string) string {
	result := ""
	for field, msg := range details {
		if result != "" {
			result += "; "
		}
		result += field + ": " + msg
	}
	return result
}

// Singleton instance
var defaultValidator = New()

// Validate is a convenience function for struct validation
func Validate(s interface{}) error {
	return defaultValidator.ValidateStruct(s)
}
