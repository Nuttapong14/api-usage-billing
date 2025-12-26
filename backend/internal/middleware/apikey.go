package middleware

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/Nuttapong14/api-usage-billing/internal/pkg/hash"
)

var (
	ErrMissingAPIKey     = errors.New("missing API key")
	ErrInvalidAPIKey     = errors.New("invalid API key")
	ErrAPIKeyExpired     = errors.New("API key has expired")
	ErrAPIKeyRevoked     = errors.New("API key has been revoked")
	ErrIPNotWhitelisted  = errors.New("IP address not in whitelist")
	ErrOriginNotAllowed  = errors.New("origin not allowed")
	ErrInsufficientScope = errors.New("insufficient scope for this operation")
)

// APIKeyInfo contains validated API key information
type APIKeyInfo struct {
	ID             uuid.UUID
	CustomerID     uuid.UUID
	OrganizationID uuid.UUID
	KeyPrefix      string
	Permissions    []string
	Scopes         []string
	IPWhitelist    []string
	AllowedOrigins []string
	ExpiresAt      *time.Time
	IsActive       bool
}

// APIKeyValidator interface for looking up API keys
type APIKeyValidator interface {
	ValidateKey(ctx context.Context, keyHash string) (*APIKeyInfo, error)
	RecordUsage(ctx context.Context, keyID uuid.UUID, ip string) error
}

// APIKeyMiddleware handles API key authentication
type APIKeyMiddleware struct {
	validator        APIKeyValidator
	headerName       string
	allowQueryParam  bool
	queryParamName   string
	checkIPWhitelist bool
	checkOrigin      bool
}

// APIKeyMiddlewareConfig configures the API key middleware
type APIKeyMiddlewareConfig struct {
	HeaderName       string
	AllowQueryParam  bool
	QueryParamName   string
	CheckIPWhitelist bool
	CheckOrigin      bool
}

// DefaultAPIKeyConfig returns default configuration
func DefaultAPIKeyConfig() APIKeyMiddlewareConfig {
	return APIKeyMiddlewareConfig{
		HeaderName:       "X-API-Key",
		AllowQueryParam:  false,
		QueryParamName:   "api_key",
		CheckIPWhitelist: true,
		CheckOrigin:      true,
	}
}

// NewAPIKeyMiddleware creates a new API key middleware
func NewAPIKeyMiddleware(validator APIKeyValidator, config ...APIKeyMiddlewareConfig) *APIKeyMiddleware {
	cfg := DefaultAPIKeyConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &APIKeyMiddleware{
		validator:        validator,
		headerName:       cfg.HeaderName,
		allowQueryParam:  cfg.AllowQueryParam,
		queryParamName:   cfg.QueryParamName,
		checkIPWhitelist: cfg.CheckIPWhitelist,
		checkOrigin:      cfg.CheckOrigin,
	}
}

// Handler returns the Fiber middleware handler
func (m *APIKeyMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract API key
		apiKey := c.Get(m.headerName)
		if apiKey == "" && m.allowQueryParam {
			apiKey = c.Query(m.queryParamName)
		}

		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": ErrMissingAPIKey.Error(),
			})
		}

		// Hash the API key
		keyHash := HashAPIKey(apiKey)

		// Validate key
		keyInfo, err := m.validator.ValidateKey(c.Context(), keyHash)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": ErrInvalidAPIKey.Error(),
			})
		}

		// Check if key is active
		if !keyInfo.IsActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": ErrAPIKeyRevoked.Error(),
			})
		}

		// Check expiration
		if keyInfo.ExpiresAt != nil && keyInfo.ExpiresAt.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": ErrAPIKeyExpired.Error(),
			})
		}

		// Check IP whitelist
		if m.checkIPWhitelist && len(keyInfo.IPWhitelist) > 0 {
			clientIP := c.IP()
			if !isIPAllowed(clientIP, keyInfo.IPWhitelist) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "forbidden",
					"message": ErrIPNotWhitelisted.Error(),
				})
			}
		}

		// Check origin
		if m.checkOrigin && len(keyInfo.AllowedOrigins) > 0 {
			origin := c.Get("Origin")
			if origin != "" && !isOriginAllowed(origin, keyInfo.AllowedOrigins) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "forbidden",
					"message": ErrOriginNotAllowed.Error(),
				})
			}
		}

		// Store key info in context
		c.Locals("api_key_id", keyInfo.ID)
		c.Locals("customer_id", keyInfo.CustomerID)
		c.Locals("organization_id", keyInfo.OrganizationID)
		c.Locals("api_key_info", keyInfo)
		c.Locals("permissions", keyInfo.Permissions)
		c.Locals("scopes", keyInfo.Scopes)

		// Record usage asynchronously
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = m.validator.RecordUsage(ctx, keyInfo.ID, c.IP())
		}()

		return c.Next()
	}
}

// HashAPIKey returns the SHA-256 hash of an API key
func HashAPIKey(key string) string {
	return hash.HashAPIKey(key)
}

// GenerateAPIKeyPrefix extracts the display prefix from an API key
func GenerateAPIKeyPrefix(key string) string {
	return hash.APIKeyPrefix(key)
}

func isIPAllowed(clientIP string, whitelist []string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	for _, allowed := range whitelist {
		// Check for CIDR notation
		if strings.Contains(allowed, "/") {
			_, network, err := net.ParseCIDR(allowed)
			if err == nil && network.Contains(ip) {
				return true
			}
		} else {
			// Direct IP comparison
			allowedIP := net.ParseIP(allowed)
			if allowedIP != nil && ip.Equal(allowedIP) {
				return true
			}
		}
	}

	return false
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	origin = strings.ToLower(strings.TrimSpace(origin))

	for _, allowed := range allowedOrigins {
		allowed = strings.ToLower(strings.TrimSpace(allowed))

		// Wildcard match
		if allowed == "*" {
			return true
		}

		// Exact match
		if origin == allowed {
			return true
		}

		// Wildcard subdomain match (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			suffix := allowed[1:] // Remove the *
			if strings.HasSuffix(origin, suffix) {
				return true
			}
		}
	}

	return false
}

// RequireScope creates middleware that requires specific API key scopes
func RequireScope(requiredScopes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopes, ok := c.Locals("scopes").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": ErrInsufficientScope.Error(),
			})
		}

		scopeSet := make(map[string]bool)
		for _, s := range scopes {
			scopeSet[s] = true
		}

		for _, required := range requiredScopes {
			if !scopeSet[required] {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "forbidden",
					"message": ErrInsufficientScope.Error(),
				})
			}
		}

		return c.Next()
	}
}

// RequirePermission creates middleware that requires specific permissions
func RequirePermission(requiredPerms ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions, ok := c.Locals("permissions").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
		}

		permSet := make(map[string]bool)
		for _, p := range permissions {
			permSet[p] = true
		}

		for _, required := range requiredPerms {
			if !permSet[required] {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "forbidden",
					"message": "insufficient permissions",
				})
			}
		}

		return c.Next()
	}
}
