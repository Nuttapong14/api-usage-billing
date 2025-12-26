package middleware

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"

	"github.com/your-org/api-usage-billing/backend/internal/api"
)

// SanitizeConfig configures request sanitization checks.
type SanitizeConfig struct {
	RejectNullBytes  bool
	RejectInvalidUTF8 bool
	MaxQueryBytes    int
	MaxBodyBytes     int
}

// DefaultSanitizeConfig returns baseline sanitization settings.
func DefaultSanitizeConfig() SanitizeConfig {
	return SanitizeConfig{
		RejectNullBytes:  true,
		RejectInvalidUTF8: true,
		MaxQueryBytes:    4096,
		MaxBodyBytes:     0,
	}
}

// NewSanitizeMiddleware validates incoming request payloads.
func NewSanitizeMiddleware(config ...SanitizeConfig) fiber.Handler {
	cfg := DefaultSanitizeConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *fiber.Ctx) error {
		if cfg.MaxQueryBytes > 0 && len(c.OriginalURL()) > cfg.MaxQueryBytes {
			return api.BadRequest(c, "request URL too large")
		}

		if cfg.RejectNullBytes && containsNullByte(c.OriginalURL()) {
			return api.BadRequest(c, "request URL contains invalid characters")
		}

		body := c.Body()
		if len(body) > 0 {
			if cfg.MaxBodyBytes > 0 && len(body) > cfg.MaxBodyBytes {
				return api.BadRequest(c, "request body too large")
			}
			if cfg.RejectNullBytes && bytes.IndexByte(body, 0) >= 0 {
				return api.BadRequest(c, "request body contains invalid characters")
			}
			if cfg.RejectInvalidUTF8 && !utf8.Valid(body) {
				return api.BadRequest(c, "request body must be valid UTF-8")
			}
		}

		return c.Next()
	}
}

func containsNullByte(value string) bool {
	return strings.IndexByte(value, 0) >= 0
}
