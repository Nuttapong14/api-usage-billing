package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"

	healthhandler "github.com/your-org/api-usage-billing/backend/internal/handler"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	// Server settings
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// App settings
	AppName        string
	Environment    string
	EnablePrefork  bool
	BodyLimit      int
	CaseSensitive  bool
	StrictRouting  bool

	// Security settings
	EnableCORS        bool
	CORSAllowOrigins  string
	CORSAllowMethods  string
	CORSAllowHeaders  string
	EnableHelmet      bool
	EnableRateLimiter bool
	RateLimitMax      int
	RateLimitDuration time.Duration

	// Logging settings
	EnableLogging bool
	LogFormat     string

	// Compression settings
	EnableCompression bool
	CompressionLevel  int

	// Recovery settings
	EnableRecovery      bool
	EnableStackTrace    bool
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,

		AppName:       "API Usage & Billing",
		Environment:   "development",
		EnablePrefork: false,
		BodyLimit:     4 * 1024 * 1024, // 4MB
		CaseSensitive: false,
		StrictRouting: false,

		EnableCORS:        true,
		CORSAllowOrigins:  "http://localhost:3000,http://localhost:3001",
		CORSAllowMethods:  "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		CORSAllowHeaders:  "Origin,Content-Type,Accept,Authorization,X-API-Key,X-Tenant-ID,X-Request-ID",
		EnableHelmet:      true,
		EnableRateLimiter: true,
		RateLimitMax:      100,
		RateLimitDuration: 1 * time.Minute,

		EnableLogging: true,
		LogFormat:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",

		EnableCompression: true,
		CompressionLevel:  int(compress.LevelDefault),

		EnableRecovery:   true,
		EnableStackTrace: true,
	}
}

// Server represents the HTTP server
type Server struct {
	app    *fiber.App
	config ServerConfig
}

// NewServer creates a new Fiber server
func NewServer(config ...ServerConfig) *Server {
	cfg := DefaultServerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               cfg.AppName,
		ReadTimeout:           cfg.ReadTimeout,
		WriteTimeout:          cfg.WriteTimeout,
		IdleTimeout:           cfg.IdleTimeout,
		BodyLimit:             cfg.BodyLimit,
		CaseSensitive:         cfg.CaseSensitive,
		StrictRouting:         cfg.StrictRouting,
		Prefork:               cfg.EnablePrefork,
		DisableStartupMessage: cfg.Environment == "production",
		ErrorHandler:          customErrorHandler,
	})

	server := &Server{
		app:    app,
		config: cfg,
	}

	// Apply middleware
	server.applyMiddleware()

	return server
}

// applyMiddleware applies all configured middleware
func (s *Server) applyMiddleware() {
	// Recovery middleware (should be first)
	if s.config.EnableRecovery {
		s.app.Use(recover.New(recover.Config{
			EnableStackTrace: s.config.EnableStackTrace,
		}))
	}

	// Request ID middleware
	s.app.Use(requestid.New(requestid.Config{
		Header: "X-Request-ID",
		Generator: func() string {
			return uuid.NewString()
		},
	}))

	// Logger middleware
	if s.config.EnableLogging {
		s.app.Use(logger.New(logger.Config{
			Format:     s.config.LogFormat,
			TimeFormat: "2006-01-02 15:04:05",
			TimeZone:   "UTC",
		}))
	}

	// Helmet middleware (security headers)
	if s.config.EnableHelmet {
		s.app.Use(helmet.New())
	}

	// CORS middleware
	if s.config.EnableCORS {
		s.app.Use(cors.New(cors.Config{
			AllowOrigins:     s.config.CORSAllowOrigins,
			AllowMethods:     s.config.CORSAllowMethods,
			AllowHeaders:     s.config.CORSAllowHeaders,
		AllowCredentials: false,
			ExposeHeaders:    "X-Request-ID,X-RateLimit-Limit,X-RateLimit-Remaining,X-RateLimit-Reset",
			MaxAge:           86400, // 24 hours
		}))
	}

	// Rate limiter middleware
	if s.config.EnableRateLimiter {
		s.app.Use(limiter.New(limiter.Config{
			Max:        s.config.RateLimitMax,
			Expiration: s.config.RateLimitDuration,
			KeyGenerator: func(c *fiber.Ctx) string {
				// Use API key if present, otherwise use IP
				if apiKey := c.Get("X-API-Key"); apiKey != "" {
					return apiKey
				}
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error":   "rate_limit_exceeded",
					"message": "Too many requests, please try again later",
				})
			},
			SkipFailedRequests:     false,
			SkipSuccessfulRequests: false,
		}))
	}

	// Compression middleware
	if s.config.EnableCompression {
		s.app.Use(compress.New(compress.Config{
			Level: compress.Level(s.config.CompressionLevel),
		}))
	}
}

// customErrorHandler handles errors in a consistent format
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":      errorCodeFromStatus(code),
		"message":    message,
		"request_id": c.Locals("requestid"),
	})
}

// errorCodeFromStatus returns an error code string from HTTP status
func errorCodeFromStatus(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "bad_request"
	case fiber.StatusUnauthorized:
		return "unauthorized"
	case fiber.StatusForbidden:
		return "forbidden"
	case fiber.StatusNotFound:
		return "not_found"
	case fiber.StatusMethodNotAllowed:
		return "method_not_allowed"
	case fiber.StatusConflict:
		return "conflict"
	case fiber.StatusUnprocessableEntity:
		return "validation_error"
	case fiber.StatusTooManyRequests:
		return "rate_limit_exceeded"
	case fiber.StatusInternalServerError:
		return "internal_error"
	case fiber.StatusServiceUnavailable:
		return "service_unavailable"
	default:
		return "error"
	}
}

// App returns the underlying Fiber app for route registration
func (s *Server) App() *fiber.App {
	return s.app
}

// Start starts the server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

// ShutdownWithTimeout shuts down with the configured timeout
func (s *Server) ShutdownWithTimeout() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()
	return s.Shutdown(ctx)
}

// RegisterHealthCheck registers a health check endpoint
func (s *Server) RegisterHealthCheck() {
	handler := healthhandler.NewHealthHandler()
	handler.Register(s.app)
}

// RegisterAPIRoutes registers API route groups
func (s *Server) RegisterAPIRoutes() fiber.Router {
	return s.app.Group("/api")
}

// RegisterV1Routes registers v1 API routes
func (s *Server) RegisterV1Routes() fiber.Router {
	api := s.RegisterAPIRoutes()
	return api.Group("/v1")
}
