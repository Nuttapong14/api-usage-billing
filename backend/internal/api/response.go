package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []ValidationError  `json:"details,omitempty"`
}

// Meta contains response metadata
type Meta struct {
	// Pagination
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalItems int `json:"total_items,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`

	// Additional metadata
	Version    string `json:"version,omitempty"`
	Deprecated bool   `json:"deprecated,omitempty"`
}

// ResponseBuilder builds API responses
type ResponseBuilder struct {
	c         *fiber.Ctx
	status    int
	data      interface{}
	error     *ErrorInfo
	meta      *Meta
	headers   map[string]string
}

// NewResponse creates a new response builder
func NewResponse(c *fiber.Ctx) *ResponseBuilder {
	return &ResponseBuilder{
		c:       c,
		status:  fiber.StatusOK,
		headers: make(map[string]string),
	}
}

// Status sets the HTTP status code
func (r *ResponseBuilder) Status(status int) *ResponseBuilder {
	r.status = status
	return r
}

// Data sets the response data
func (r *ResponseBuilder) Data(data interface{}) *ResponseBuilder {
	r.data = data
	return r
}

// Error sets the error info
func (r *ResponseBuilder) Error(code, message string) *ResponseBuilder {
	r.error = &ErrorInfo{
		Code:    code,
		Message: message,
	}
	return r
}

// ErrorWithDetails sets the error with validation details
func (r *ResponseBuilder) ErrorWithDetails(code, message string, details []ValidationError) *ResponseBuilder {
	r.error = &ErrorInfo{
		Code:    code,
		Message: message,
		Details: details,
	}
	return r
}

// Meta sets the response metadata
func (r *ResponseBuilder) Meta(meta *Meta) *ResponseBuilder {
	r.meta = meta
	return r
}

// Pagination sets pagination metadata
func (r *ResponseBuilder) Pagination(page, pageSize, totalItems, totalPages int) *ResponseBuilder {
	if r.meta == nil {
		r.meta = &Meta{}
	}
	r.meta.Page = page
	r.meta.PageSize = pageSize
	r.meta.TotalItems = totalItems
	r.meta.TotalPages = totalPages
	return r
}

// Header adds a response header
func (r *ResponseBuilder) Header(key, value string) *ResponseBuilder {
	r.headers[key] = value
	return r
}

// Deprecated marks the response as deprecated
func (r *ResponseBuilder) Deprecated() *ResponseBuilder {
	if r.meta == nil {
		r.meta = &Meta{}
	}
	r.meta.Deprecated = true
	r.headers["Deprecation"] = "true"
	return r
}

// Version sets the API version in metadata
func (r *ResponseBuilder) Version(version string) *ResponseBuilder {
	if r.meta == nil {
		r.meta = &Meta{}
	}
	r.meta.Version = version
	return r
}

// Send sends the response
func (r *ResponseBuilder) Send() error {
	// Set headers
	for key, value := range r.headers {
		r.c.Set(key, value)
	}

	// Get request ID
	requestID := ""
	if id := r.c.Locals("requestid"); id != nil {
		requestID = id.(string)
	}

	// Build response
	response := Response{
		Success:   r.error == nil,
		Data:      r.data,
		Error:     r.error,
		Meta:      r.meta,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return r.c.Status(r.status).JSON(response)
}

// SendRaw sends data without the standard wrapper
func (r *ResponseBuilder) SendRaw() error {
	for key, value := range r.headers {
		r.c.Set(key, value)
	}
	return r.c.Status(r.status).JSON(r.data)
}

// Helper functions for common responses

// OK sends a 200 OK response with data
func OK(c *fiber.Ctx, data interface{}) error {
	return NewResponse(c).Status(fiber.StatusOK).Data(data).Send()
}

// Created sends a 201 Created response with data
func Created(c *fiber.Ctx, data interface{}) error {
	return NewResponse(c).Status(fiber.StatusCreated).Data(data).Send()
}

// Accepted sends a 202 Accepted response
func Accepted(c *fiber.Ctx, data interface{}) error {
	return NewResponse(c).Status(fiber.StatusAccepted).Data(data).Send()
}

// NoContent sends a 204 No Content response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// List sends a paginated list response
func List(c *fiber.Ctx, items interface{}, page, pageSize, totalItems, totalPages int) error {
	return NewResponse(c).
		Status(fiber.StatusOK).
		Data(items).
		Pagination(page, pageSize, totalItems, totalPages).
		Send()
}

// BadRequest sends a 400 Bad Request error
func BadRequest(c *fiber.Ctx, message string) error {
	return NewResponse(c).
		Status(fiber.StatusBadRequest).
		Error("bad_request", message).
		Send()
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(c *fiber.Ctx, message string) error {
	return NewResponse(c).
		Status(fiber.StatusUnauthorized).
		Error("unauthorized", message).
		Send()
}

// Forbidden sends a 403 Forbidden error
func Forbidden(c *fiber.Ctx, message string) error {
	return NewResponse(c).
		Status(fiber.StatusForbidden).
		Error("forbidden", message).
		Send()
}

// NotFound sends a 404 Not Found error
func NotFound(c *fiber.Ctx, message string) error {
	return NewResponse(c).
		Status(fiber.StatusNotFound).
		Error("not_found", message).
		Send()
}

// Conflict sends a 409 Conflict error
func Conflict(c *fiber.Ctx, message string) error {
	return NewResponse(c).
		Status(fiber.StatusConflict).
		Error("conflict", message).
		Send()
}

// ValidationFailed sends a 422 Unprocessable Entity error with validation details
func ValidationFailed(c *fiber.Ctx, errors []ValidationError) error {
	return NewResponse(c).
		Status(fiber.StatusUnprocessableEntity).
		ErrorWithDetails("validation_error", "Validation failed", errors).
		Send()
}

// TooManyRequests sends a 429 Too Many Requests error
func TooManyRequests(c *fiber.Ctx, message string) error {
	return NewResponse(c).
		Status(fiber.StatusTooManyRequests).
		Error("rate_limit_exceeded", message).
		Send()
}

// InternalError sends a 500 Internal Server Error
func InternalError(c *fiber.Ctx, message string) error {
	return NewResponse(c).
		Status(fiber.StatusInternalServerError).
		Error("internal_error", message).
		Send()
}

// ServiceUnavailable sends a 503 Service Unavailable error
func ServiceUnavailable(c *fiber.Ctx, message string) error {
	return NewResponse(c).
		Status(fiber.StatusServiceUnavailable).
		Error("service_unavailable", message).
		Send()
}

// FromError sends an appropriate error response based on the error type
func FromError(c *fiber.Ctx, err error) error {
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.Details != nil {
			return NewResponse(c).
				Status(apiErr.Status).
				ErrorWithDetails(apiErr.Code, apiErr.Message, apiErr.Details).
				Send()
		}
		return NewResponse(c).
			Status(apiErr.Status).
			Error(apiErr.Code, apiErr.Message).
			Send()
	}

	// Default to internal error
	return InternalError(c, "An unexpected error occurred")
}
