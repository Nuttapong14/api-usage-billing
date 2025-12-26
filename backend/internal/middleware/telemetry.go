package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryConfig configures the telemetry middleware
type TelemetryConfig struct {
	// ServiceName is the name of the service
	ServiceName string

	// TracerProvider is the OpenTelemetry tracer provider
	TracerProvider trace.TracerProvider

	// MeterProvider is the OpenTelemetry meter provider
	MeterProvider metric.MeterProvider

	// Propagators is the text map propagator for context propagation
	Propagators propagation.TextMapPropagator

	// SkipPaths are paths to skip telemetry
	SkipPaths []string

	// RecordRequestBody determines if request body should be recorded
	RecordRequestBody bool

	// RecordResponseBody determines if response body should be recorded
	RecordResponseBody bool

	// MaxBodySize is the maximum body size to record
	MaxBodySize int
}

// DefaultTelemetryConfig returns default configuration
func DefaultTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    "api-usage-billing",
		TracerProvider: otel.GetTracerProvider(),
		MeterProvider:  otel.GetMeterProvider(),
		Propagators:    otel.GetTextMapPropagator(),
		SkipPaths:      []string{"/health", "/ready", "/metrics"},
		MaxBodySize:    1024,
	}
}

// TelemetryMiddleware provides OpenTelemetry tracing and metrics
type TelemetryMiddleware struct {
	config          TelemetryConfig
	tracer          trace.Tracer
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	requestSize     metric.Int64Histogram
	responseSize    metric.Int64Histogram
}

// NewTelemetryMiddleware creates a new telemetry middleware
func NewTelemetryMiddleware(config ...TelemetryConfig) (*TelemetryMiddleware, error) {
	cfg := DefaultTelemetryConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	tracer := cfg.TracerProvider.Tracer(
		cfg.ServiceName,
		trace.WithInstrumentationVersion("1.0.0"),
	)

	meter := cfg.MeterProvider.Meter(
		cfg.ServiceName,
		metric.WithInstrumentationVersion("1.0.0"),
	)

	// Create metrics
	requestCounter, err := meter.Int64Counter(
		"http.server.request_count",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request counter: %w", err)
	}

	requestDuration, err := meter.Float64Histogram(
		"http.server.duration",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create duration histogram: %w", err)
	}

	requestSize, err := meter.Int64Histogram(
		"http.server.request_size",
		metric.WithDescription("HTTP request body size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request size histogram: %w", err)
	}

	responseSize, err := meter.Int64Histogram(
		"http.server.response_size",
		metric.WithDescription("HTTP response body size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create response size histogram: %w", err)
	}

	return &TelemetryMiddleware{
		config:          cfg,
		tracer:          tracer,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
		requestSize:     requestSize,
		responseSize:    responseSize,
	}, nil
}

// Handler returns the Fiber middleware handler
func (m *TelemetryMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if path should be skipped
		path := c.Path()
		for _, skip := range m.config.SkipPaths {
			if path == skip {
				return c.Next()
			}
		}

		// Extract trace context from headers
		ctx := m.config.Propagators.Extract(c.Context(), propagation.HeaderCarrier(c.GetReqHeaders()))

		// Start span
		spanName := fmt.Sprintf("%s %s", c.Method(), c.Route().Path)
		ctx, span := m.tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.route", c.Route().Path),
				attribute.String("http.host", c.Hostname()),
				attribute.String("http.scheme", c.Protocol()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
				attribute.String("http.client_ip", c.IP()),
				attribute.Int("http.request_content_length", len(c.Body())),
			),
		)
		defer span.End()

		// Store trace context in Fiber context
		c.Locals("trace_id", span.SpanContext().TraceID().String())
		c.Locals("span_id", span.SpanContext().SpanID().String())
		c.SetUserContext(ctx)

		// Record request start time
		startTime := time.Now()

		// Record request size
		requestBodySize := int64(len(c.Body()))
		if requestBodySize > 0 {
			m.requestSize.Record(ctx, requestBodySize, metric.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.route", c.Route().Path),
			))
		}

		// Add request ID to span
		if requestID := GetRequestID(c); requestID != "" {
			span.SetAttributes(attribute.String("request_id", requestID))
		}

		// Add tenant info to span
		if tenantID, ok := GetTenantID(c); ok {
			span.SetAttributes(attribute.String("tenant_id", tenantID.String()))
		}

		// Process request
		err := c.Next()

		// Record response metrics
		duration := time.Since(startTime).Seconds()
		statusCode := c.Response().StatusCode()
		responseBodySize := int64(len(c.Response().Body()))

		// Set span attributes
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int("http.response_content_length", int(responseBodySize)),
		)

		// Set span status based on HTTP status code
		if statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
			if err != nil {
				span.RecordError(err)
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Record metrics
		commonAttrs := []attribute.KeyValue{
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", c.Route().Path),
			attribute.Int("http.status_code", statusCode),
		}

		m.requestCounter.Add(ctx, 1, metric.WithAttributes(commonAttrs...))
		m.requestDuration.Record(ctx, duration, metric.WithAttributes(commonAttrs...))
		m.responseSize.Record(ctx, responseBodySize, metric.WithAttributes(commonAttrs...))

		return err
	}
}

// TraceMiddleware provides simple span creation without full metrics
func TraceMiddleware(tracer trace.Tracer, skipPaths ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		for _, skip := range skipPaths {
			if path == skip {
				return c.Next()
			}
		}

		spanName := fmt.Sprintf("%s %s", c.Method(), c.Route().Path)
		ctx, span := tracer.Start(c.Context(), spanName,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		c.Locals("trace_id", span.SpanContext().TraceID().String())
		c.Locals("span_id", span.SpanContext().SpanID().String())
		c.SetUserContext(ctx)

		err := c.Next()

		if err != nil || c.Response().StatusCode() >= 400 {
			span.SetStatus(codes.Error, "")
			if err != nil {
				span.RecordError(err)
			}
		}

		return err
	}
}

// SpanFromContext extracts span from Fiber context
func SpanFromContext(c *fiber.Ctx) trace.Span {
	return trace.SpanFromContext(c.UserContext())
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(c *fiber.Ctx, name string, attrs ...attribute.KeyValue) {
	span := SpanFromContext(c)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanError records an error on the current span
func SetSpanError(c *fiber.Ctx, err error) {
	span := SpanFromContext(c)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanAttribute sets an attribute on the current span
func SetSpanAttribute(c *fiber.Ctx, key string, value interface{}) {
	span := SpanFromContext(c)
	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// StartChildSpan starts a child span for the current request
func StartChildSpan(c *fiber.Ctx, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("")
	return tracer.Start(c.UserContext(), name, opts...)
}
