package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// APIError represents an API error
type APIError struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
	Err     error             `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.Err
}

// WithError adds an underlying error
func (e *APIError) WithError(err error) *APIError {
	e.Err = err
	return e
}

// WithDetails adds validation details
func (e *APIError) WithDetails(details []ValidationError) *APIError {
	e.Details = details
	return e
}

// Error constructors

// NewBadRequestError creates a 400 Bad Request error
func NewBadRequestError(message string) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "bad_request",
		Message: message,
	}
}

// NewUnauthorizedError creates a 401 Unauthorized error
func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		Status:  http.StatusUnauthorized,
		Code:    "unauthorized",
		Message: message,
	}
}

// NewForbiddenError creates a 403 Forbidden error
func NewForbiddenError(message string) *APIError {
	return &APIError{
		Status:  http.StatusForbidden,
		Code:    "forbidden",
		Message: message,
	}
}

// NewNotFoundError creates a 404 Not Found error
func NewNotFoundError(message string) *APIError {
	return &APIError{
		Status:  http.StatusNotFound,
		Code:    "not_found",
		Message: message,
	}
}

// NewConflictError creates a 409 Conflict error
func NewConflictError(message string) *APIError {
	return &APIError{
		Status:  http.StatusConflict,
		Code:    "conflict",
		Message: message,
	}
}

// NewValidationError creates a 422 Unprocessable Entity error
func NewValidationError(details []ValidationError) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    "validation_error",
		Message: "Validation failed",
		Details: details,
	}
}

// NewTooManyRequestsError creates a 429 Too Many Requests error
func NewTooManyRequestsError(message string) *APIError {
	return &APIError{
		Status:  http.StatusTooManyRequests,
		Code:    "rate_limit_exceeded",
		Message: message,
	}
}

// NewInternalError creates a 500 Internal Server Error
func NewInternalError(message string) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: message,
	}
}

// NewServiceUnavailableError creates a 503 Service Unavailable error
func NewServiceUnavailableError(message string) *APIError {
	return &APIError{
		Status:  http.StatusServiceUnavailable,
		Code:    "service_unavailable",
		Message: message,
	}
}

// NewQuotaExceededError creates a quota exceeded error
func NewQuotaExceededError(message string) *APIError {
	return &APIError{
		Status:  http.StatusPaymentRequired,
		Code:    "quota_exceeded",
		Message: message,
	}
}

// NewPaymentRequiredError creates a 402 Payment Required error
func NewPaymentRequiredError(message string) *APIError {
	return &APIError{
		Status:  http.StatusPaymentRequired,
		Code:    "payment_required",
		Message: message,
	}
}

// NewGoneError creates a 410 Gone error
func NewGoneError(message string) *APIError {
	return &APIError{
		Status:  http.StatusGone,
		Code:    "gone",
		Message: message,
	}
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// GetAPIError extracts an APIError from an error
func GetAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// ErrorHandler is the global Fiber error handler
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default error
	code := fiber.StatusInternalServerError
	errorCode := "internal_error"
	message := "An unexpected error occurred"
	var details []ValidationError

	// Check for APIError
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		code = apiErr.Status
		errorCode = apiErr.Code
		message = apiErr.Message
		details = apiErr.Details
	} else if e, ok := err.(*fiber.Error); ok {
		// Fiber error
		code = e.Code
		errorCode = errorCodeFromHTTPStatus(code)
		message = e.Message
	}

	// Get request ID
	requestID := ""
	if id := c.Locals("requestid"); id != nil {
		requestID = id.(string)
	}

	// Build error response
	response := fiber.Map{
		"success":    false,
		"error":      errorCode,
		"message":    message,
		"request_id": requestID,
	}

	if len(details) > 0 {
		response["details"] = details
	}

	return c.Status(code).JSON(response)
}

// errorCodeFromHTTPStatus returns an error code from HTTP status
func errorCodeFromHTTPStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusPaymentRequired:
		return "payment_required"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	case http.StatusConflict:
		return "conflict"
	case http.StatusGone:
		return "gone"
	case http.StatusUnprocessableEntity:
		return "validation_error"
	case http.StatusTooManyRequests:
		return "rate_limit_exceeded"
	case http.StatusInternalServerError:
		return "internal_error"
	case http.StatusBadGateway:
		return "bad_gateway"
	case http.StatusServiceUnavailable:
		return "service_unavailable"
	case http.StatusGatewayTimeout:
		return "gateway_timeout"
	default:
		return "error"
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.WithError(fmt.Errorf("%s: %w", message, apiErr.Err))
	}
	return fmt.Errorf("%s: %w", message, err)
}

// HandleError is a helper to handle errors in handlers
func HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	return FromError(c, err)
}

// MustNoError panics if there's an error (for use in initialization)
func MustNoError(err error) {
	if err != nil {
		panic(err)
	}
}

// RecoverMiddleware creates a panic recovery middleware
func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				if err, ok = r.(error); !ok {
					err = fmt.Errorf("%v", r)
				}
				err = NewInternalError("Internal server error").WithError(err)
			}
		}()
		return c.Next()
	}
}

// Common domain errors

var (
	ErrResourceNotFound    = NewNotFoundError("Resource not found")
	ErrUnauthorized        = NewUnauthorizedError("Authentication required")
	ErrForbidden           = NewForbiddenError("Access denied")
	ErrInvalidInput        = NewBadRequestError("Invalid input")
	ErrDuplicateResource   = NewConflictError("Resource already exists")
	ErrRateLimitExceeded   = NewTooManyRequestsError("Rate limit exceeded")
	ErrQuotaExceeded       = NewQuotaExceededError("Usage quota exceeded")
	ErrServiceUnavailable  = NewServiceUnavailableError("Service temporarily unavailable")
)
