package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Context key for request ID
type requestIDKey struct{}

// RequestIDConfig configures the request ID middleware
type RequestIDConfig struct {
	// Header is the header name to read/write the request ID
	Header string

	// ContextKey is the key to store request ID in context
	ContextKey string

	// Generator is the function to generate request IDs
	Generator func() string

	// TrustProxy determines whether to trust incoming request ID headers
	TrustProxy bool
}

// DefaultRequestIDConfig returns default configuration
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header:     "X-Request-ID",
		ContextKey: "request_id",
		Generator: func() string {
			return uuid.NewString()
		},
		TrustProxy: true,
	}
}

// RequestIDMiddleware handles request ID generation and propagation
type RequestIDMiddleware struct {
	config RequestIDConfig
}

// NewRequestIDMiddleware creates a new request ID middleware
func NewRequestIDMiddleware(config ...RequestIDConfig) *RequestIDMiddleware {
	cfg := DefaultRequestIDConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Generator == nil {
		cfg.Generator = func() string {
			return uuid.NewString()
		}
	}

	return &RequestIDMiddleware{
		config: cfg,
	}
}

// Handler returns the Fiber middleware handler
func (m *RequestIDMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestID string

		// Try to get request ID from header if proxy is trusted
		if m.config.TrustProxy {
			requestID = c.Get(m.config.Header)
		}

		// Generate new request ID if not found
		if requestID == "" {
			requestID = m.config.Generator()
		}

		// Store in Fiber context
		c.Locals(m.config.ContextKey, requestID)
		c.Locals("requestid", requestID) // Standard key for response builder

		// Set response header
		c.Set(m.config.Header, requestID)

		return c.Next()
	}
}

// GetRequestID extracts request ID from Fiber context
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("request_id").(string); ok {
		return id
	}
	if id, ok := c.Locals("requestid").(string); ok {
		return id
	}
	return ""
}

// WithRequestID adds request ID to a context.Context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDFromContext extracts request ID from context.Context
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// PropagateRequestID creates a context with the request ID from Fiber context
func PropagateRequestID(c *fiber.Ctx) context.Context {
	requestID := GetRequestID(c)
	return WithRequestID(c.Context(), requestID)
}

// CorrelationIDConfig configures correlation ID handling
type CorrelationIDConfig struct {
	// Header is the header name for correlation ID
	Header string

	// RequestIDHeader is the header to use as fallback
	RequestIDHeader string

	// ContextKey is the key to store correlation ID in context
	ContextKey string
}

// DefaultCorrelationIDConfig returns default configuration
func DefaultCorrelationIDConfig() CorrelationIDConfig {
	return CorrelationIDConfig{
		Header:          "X-Correlation-ID",
		RequestIDHeader: "X-Request-ID",
		ContextKey:      "correlation_id",
	}
}

// CorrelationIDMiddleware handles correlation ID for distributed tracing
type CorrelationIDMiddleware struct {
	config CorrelationIDConfig
}

// NewCorrelationIDMiddleware creates a new correlation ID middleware
func NewCorrelationIDMiddleware(config ...CorrelationIDConfig) *CorrelationIDMiddleware {
	cfg := DefaultCorrelationIDConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &CorrelationIDMiddleware{
		config: cfg,
	}
}

// Handler returns the Fiber middleware handler
func (m *CorrelationIDMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try to get correlation ID from header
		correlationID := c.Get(m.config.Header)

		// Fall back to request ID if no correlation ID
		if correlationID == "" {
			correlationID = c.Get(m.config.RequestIDHeader)
		}

		// Fall back to existing request ID local
		if correlationID == "" {
			if id, ok := c.Locals("request_id").(string); ok {
				correlationID = id
			}
		}

		// Generate new if still empty
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		// Store in context
		c.Locals(m.config.ContextKey, correlationID)

		// Set response header
		c.Set(m.config.Header, correlationID)

		return c.Next()
	}
}

// GetCorrelationID extracts correlation ID from Fiber context
func GetCorrelationID(c *fiber.Ctx) string {
	if id, ok := c.Locals("correlation_id").(string); ok {
		return id
	}
	return GetRequestID(c)
}

// TraceContext holds trace-related IDs
type TraceContext struct {
	RequestID     string
	CorrelationID string
	TraceID       string
	SpanID        string
	ParentSpanID  string
}

// GetTraceContext extracts all trace-related IDs from Fiber context
func GetTraceContext(c *fiber.Ctx) TraceContext {
	tc := TraceContext{
		RequestID:     GetRequestID(c),
		CorrelationID: GetCorrelationID(c),
	}

	if traceID, ok := c.Locals("trace_id").(string); ok {
		tc.TraceID = traceID
	}
	if spanID, ok := c.Locals("span_id").(string); ok {
		tc.SpanID = spanID
	}
	if parentSpanID, ok := c.Locals("parent_span_id").(string); ok {
		tc.ParentSpanID = parentSpanID
	}

	return tc
}

// ToMap converts TraceContext to a map for logging
func (tc TraceContext) ToMap() map[string]string {
	m := make(map[string]string)
	if tc.RequestID != "" {
		m["request_id"] = tc.RequestID
	}
	if tc.CorrelationID != "" {
		m["correlation_id"] = tc.CorrelationID
	}
	if tc.TraceID != "" {
		m["trace_id"] = tc.TraceID
	}
	if tc.SpanID != "" {
		m["span_id"] = tc.SpanID
	}
	if tc.ParentSpanID != "" {
		m["parent_span_id"] = tc.ParentSpanID
	}
	return m
}
