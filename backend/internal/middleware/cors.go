package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig defines CORS settings for HTTP handlers.
type CORSConfig struct {
	AllowOrigins     string
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns baseline CORS settings.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key,X-Tenant-ID,X-Request-ID",
		ExposeHeaders:    "X-Request-ID,X-RateLimit-Limit,X-RateLimit-Remaining,X-RateLimit-Reset",
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// NewCORSMiddleware creates a Fiber CORS middleware handler.
func NewCORSMiddleware(config ...CORSConfig) fiber.Handler {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
