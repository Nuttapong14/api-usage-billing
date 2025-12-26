package api

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	validate *validator.Validate
)

func init() {
	validate = validator.New()

	// Register custom tag name function to use json tags
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	registerCustomValidators()
}

// registerCustomValidators registers custom validation rules
func registerCustomValidators() {
	// UUID validation
	validate.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		_, err := uuid.Parse(fl.Field().String())
		return err == nil
	})

	// Slug validation (lowercase alphanumeric with hyphens)
	validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
		return slugRegex.MatchString(fl.Field().String())
	})

	// Phone number validation (basic international format)
	validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		return phoneRegex.MatchString(fl.Field().String())
	})

	// Currency code validation (ISO 4217)
	validate.RegisterValidation("currency", func(fl validator.FieldLevel) bool {
		currencyRegex := regexp.MustCompile(`^[A-Z]{3}$`)
		return currencyRegex.MatchString(fl.Field().String())
	})

	// Country code validation (ISO 3166-1 alpha-2)
	validate.RegisterValidation("country_code", func(fl validator.FieldLevel) bool {
		countryRegex := regexp.MustCompile(`^[A-Z]{2}$`)
		return countryRegex.MatchString(fl.Field().String())
	})

	// Timezone validation
	validate.RegisterValidation("timezone", func(fl validator.FieldLevel) bool {
		tzRegex := regexp.MustCompile(`^[A-Za-z_]+/[A-Za-z_]+$`)
		val := fl.Field().String()
		return val == "UTC" || tzRegex.MatchString(val)
	})

	// API key prefix validation
	validate.RegisterValidation("api_key_prefix", func(fl validator.FieldLevel) bool {
		prefixRegex := regexp.MustCompile(`^[a-zA-Z0-9]{8,16}$`)
		return prefixRegex.MatchString(fl.Field().String())
	})

	// Safe string (no HTML/script injection)
	validate.RegisterValidation("safe_string", func(fl validator.FieldLevel) bool {
		dangerousChars := regexp.MustCompile(`[<>'";&]`)
		return !dangerousChars.MatchString(fl.Field().String())
	})

	// Positive decimal
	validate.RegisterValidation("positive_decimal", func(fl validator.FieldLevel) bool {
		switch fl.Field().Kind() {
		case reflect.Float32, reflect.Float64:
			return fl.Field().Float() > 0
		case reflect.Int, reflect.Int32, reflect.Int64:
			return fl.Field().Int() > 0
		}
		return false
	})

	// Non-negative decimal
	validate.RegisterValidation("non_negative", func(fl validator.FieldLevel) bool {
		switch fl.Field().Kind() {
		case reflect.Float32, reflect.Float64:
			return fl.Field().Float() >= 0
		case reflect.Int, reflect.Int32, reflect.Int64:
			return fl.Field().Int() >= 0
		}
		return false
	})
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	var sb strings.Builder
	for i, e := range ve.Errors {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return sb.String()
}

// Validate validates a struct and returns validation errors
func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors := ValidationErrors{}

	for _, err := range err.(validator.ValidationErrors) {
		ve := ValidationError{
			Field:   err.Field(),
			Tag:     err.Tag(),
			Message: formatValidationError(err),
		}

		// Include value for non-sensitive fields
		if !isSensitiveField(err.Field()) {
			ve.Value = fmt.Sprintf("%v", err.Value())
		}

		validationErrors.Errors = append(validationErrors.Errors, ve)
	}

	return validationErrors
}

// formatValidationError creates a human-readable error message
func formatValidationError(fe validator.FieldError) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "slug":
		return fmt.Sprintf("%s must be a valid slug (lowercase letters, numbers, and hyphens)", field)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "currency":
		return fmt.Sprintf("%s must be a valid ISO 4217 currency code", field)
	case "country_code":
		return fmt.Sprintf("%s must be a valid ISO 3166-1 alpha-2 country code", field)
	case "timezone":
		return fmt.Sprintf("%s must be a valid timezone", field)
	case "safe_string":
		return fmt.Sprintf("%s contains invalid characters", field)
	case "positive_decimal":
		return fmt.Sprintf("%s must be a positive number", field)
	case "non_negative":
		return fmt.Sprintf("%s must be a non-negative number", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "numeric":
		return fmt.Sprintf("%s must contain only numbers", field)
	default:
		return fmt.Sprintf("%s failed %s validation", field, fe.Tag())
	}
}

// isSensitiveField checks if a field contains sensitive data
func isSensitiveField(field string) bool {
	sensitiveFields := []string{
		"password", "token", "secret", "key", "api_key",
		"credit_card", "card_number", "cvv", "ssn",
	}

	lowerField := strings.ToLower(field)
	for _, s := range sensitiveFields {
		if strings.Contains(lowerField, s) {
			return true
		}
	}
	return false
}

// ParseAndValidate parses the request body and validates it
func ParseAndValidate(c *fiber.Ctx, dest interface{}) error {
	// Parse body
	if err := c.BodyParser(dest); err != nil {
		return NewBadRequestError("Invalid request body: " + err.Error())
	}

	// Validate
	if err := Validate(dest); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			return NewValidationError(ve.Errors)
		}
		return NewBadRequestError(err.Error())
	}

	return nil
}

// ParseQueryAndValidate parses query parameters and validates them
func ParseQueryAndValidate(c *fiber.Ctx, dest interface{}) error {
	// Parse query parameters
	if err := c.QueryParser(dest); err != nil {
		return NewBadRequestError("Invalid query parameters: " + err.Error())
	}

	// Validate
	if err := Validate(dest); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			return NewValidationError(ve.Errors)
		}
		return NewBadRequestError(err.Error())
	}

	return nil
}

// ParseParamsAndValidate parses path parameters and validates them
func ParseParamsAndValidate(c *fiber.Ctx, dest interface{}) error {
	// Parse path parameters
	if err := c.ParamsParser(dest); err != nil {
		return NewBadRequestError("Invalid path parameters: " + err.Error())
	}

	// Validate
	if err := Validate(dest); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			return NewValidationError(ve.Errors)
		}
		return NewBadRequestError(err.Error())
	}

	return nil
}

// ValidateUUID validates and parses a UUID string
func ValidateUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, NewBadRequestError("Invalid UUID format")
	}
	return id, nil
}

// ValidateUUIDParam validates and parses a UUID path parameter
func ValidateUUIDParam(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	idStr := c.Params(paramName)
	if idStr == "" {
		return uuid.Nil, NewBadRequestError(fmt.Sprintf("Missing %s parameter", paramName))
	}
	return ValidateUUID(idStr)
}

// GetValidator returns the validator instance for custom use
func GetValidator() *validator.Validate {
	return validate
}
