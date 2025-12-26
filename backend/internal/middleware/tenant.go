package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	ErrMissingTenant     = errors.New("tenant context is required")
	ErrInvalidTenant     = errors.New("invalid tenant")
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrTenantMismatch    = errors.New("tenant mismatch")
	ErrTenantSuspended   = errors.New("tenant is suspended")
	ErrCrossTenantAccess = errors.New("cross-tenant access not allowed")
)

// TenantInfo contains tenant information
type TenantInfo struct {
	ID        uuid.UUID
	Slug      string
	Name      string
	IsActive  bool
	Settings  map[string]interface{}
}

// TenantResolver resolves tenant from various sources
type TenantResolver interface {
	// ResolveByID looks up tenant by ID
	ResolveByID(ctx context.Context, id uuid.UUID) (*TenantInfo, error)

	// ResolveBySlug looks up tenant by slug
	ResolveBySlug(ctx context.Context, slug string) (*TenantInfo, error)

	// ResolveByDomain looks up tenant by domain
	ResolveByDomain(ctx context.Context, domain string) (*TenantInfo, error)
}

// TenantMiddlewareConfig configures the tenant middleware
type TenantMiddlewareConfig struct {
	// HeaderName is the header to look for tenant ID
	HeaderName string

	// QueryParamName is the query parameter for tenant ID
	QueryParamName string

	// PathParamName is the path parameter for tenant slug
	PathParamName string

	// AllowSubdomain enables subdomain-based tenant resolution
	AllowSubdomain bool

	// SubdomainSuffix is the base domain (e.g., ".example.com")
	SubdomainSuffix string

	// RequireTenant fails requests without tenant context
	RequireTenant bool

	// EnforceTenantIsolation validates tenant access
	EnforceTenantIsolation bool

	// AllowedCrossTenantRoles are roles that can access multiple tenants
	AllowedCrossTenantRoles []string
}

// DefaultTenantConfig returns default configuration
func DefaultTenantConfig() TenantMiddlewareConfig {
	return TenantMiddlewareConfig{
		HeaderName:              "X-Tenant-ID",
		QueryParamName:          "tenant_id",
		PathParamName:           "org",
		AllowSubdomain:          false,
		RequireTenant:           true,
		EnforceTenantIsolation:  true,
		AllowedCrossTenantRoles: []string{"super_admin", "platform_admin"},
	}
}

// TenantMiddleware handles multi-tenant context
type TenantMiddleware struct {
	resolver TenantResolver
	config   TenantMiddlewareConfig
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(resolver TenantResolver, config ...TenantMiddlewareConfig) *TenantMiddleware {
	cfg := DefaultTenantConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &TenantMiddleware{
		resolver: resolver,
		config:   cfg,
	}
}

// Handler returns the Fiber middleware handler
func (m *TenantMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tenant *TenantInfo
		var err error

		// Try to resolve tenant from various sources (priority order)

		// 1. From header
		if tenantIDStr := c.Get(m.config.HeaderName); tenantIDStr != "" {
			tenantID, parseErr := uuid.Parse(tenantIDStr)
			if parseErr == nil {
				tenant, err = m.resolver.ResolveByID(c.Context(), tenantID)
			}
		}

		// 2. From path parameter (e.g., /api/v1/orgs/:org/...)
		if tenant == nil {
			if slug := c.Params(m.config.PathParamName); slug != "" {
				tenant, err = m.resolver.ResolveBySlug(c.Context(), slug)
			}
		}

		// 3. From query parameter
		if tenant == nil {
			if tenantIDStr := c.Query(m.config.QueryParamName); tenantIDStr != "" {
				tenantID, parseErr := uuid.Parse(tenantIDStr)
				if parseErr == nil {
					tenant, err = m.resolver.ResolveByID(c.Context(), tenantID)
				}
			}
		}

		// 4. From subdomain
		if tenant == nil && m.config.AllowSubdomain && m.config.SubdomainSuffix != "" {
			host := c.Hostname()
			if strings.HasSuffix(host, m.config.SubdomainSuffix) {
				subdomain := strings.TrimSuffix(host, m.config.SubdomainSuffix)
				if subdomain != "" && subdomain != "www" && subdomain != "api" {
					tenant, err = m.resolver.ResolveBySlug(c.Context(), subdomain)
				}
			}
		}

		// 5. From authenticated user's organization
		if tenant == nil {
			if orgID, ok := c.Locals("organization_id").(uuid.UUID); ok && orgID != uuid.Nil {
				tenant, err = m.resolver.ResolveByID(c.Context(), orgID)
			}
		}

		// Handle resolution errors
		if err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "not_found",
					"message": "tenant not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "failed to resolve tenant",
			})
		}

		// Check if tenant is required
		if tenant == nil && m.config.RequireTenant {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "bad_request",
				"message": ErrMissingTenant.Error(),
			})
		}

		// Check if tenant is active
		if tenant != nil && !tenant.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": ErrTenantSuspended.Error(),
			})
		}

		// Enforce tenant isolation
		if tenant != nil && m.config.EnforceTenantIsolation {
			if err := m.enforceTenantIsolation(c, tenant); err != nil {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "forbidden",
					"message": err.Error(),
				})
			}
		}

		// Store tenant info in context
		if tenant != nil {
			c.Locals("tenant_id", tenant.ID)
			c.Locals("tenant_slug", tenant.Slug)
			c.Locals("tenant_name", tenant.Name)
			c.Locals("tenant_info", tenant)
			c.Locals("organization_id", tenant.ID) // Alias for consistency
		}

		return c.Next()
	}
}

// enforceTenantIsolation validates that the authenticated user has access to the tenant
func (m *TenantMiddleware) enforceTenantIsolation(c *fiber.Ctx, tenant *TenantInfo) error {
	// Check if user has cross-tenant role
	if m.hasCrossTenantRole(c) {
		return nil
	}

	// For API key authentication, check if key belongs to tenant
	if keyOrgID, ok := c.Locals("organization_id").(uuid.UUID); ok {
		if keyOrgID != uuid.Nil && keyOrgID != tenant.ID {
			return ErrCrossTenantAccess
		}
	}

	// For JWT authentication, user's organization must match
	// This assumes user's org is stored during JWT validation
	if userOrgID, ok := c.Locals("user_organization_id").(uuid.UUID); ok {
		if userOrgID != uuid.Nil && userOrgID != tenant.ID {
			return ErrCrossTenantAccess
		}
	}

	return nil
}

// hasCrossTenantRole checks if user has a role that allows cross-tenant access
func (m *TenantMiddleware) hasCrossTenantRole(c *fiber.Ctx) bool {
	permissions, ok := c.Locals("permissions").([]string)
	if !ok {
		return false
	}

	for _, perm := range permissions {
		for _, allowed := range m.config.AllowedCrossTenantRoles {
			if perm == allowed {
				return true
			}
		}
	}
	return false
}

// RequireTenantContext creates middleware that requires tenant context to be set
func RequireTenantContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID, ok := c.Locals("tenant_id").(uuid.UUID)
		if !ok || tenantID == uuid.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "bad_request",
				"message": ErrMissingTenant.Error(),
			})
		}
		return c.Next()
	}
}

// GetTenantID extracts tenant ID from Fiber context
func GetTenantID(c *fiber.Ctx) (uuid.UUID, bool) {
	if id, ok := c.Locals("tenant_id").(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}

// GetTenantSlug extracts tenant slug from Fiber context
func GetTenantSlug(c *fiber.Ctx) (string, bool) {
	if slug, ok := c.Locals("tenant_slug").(string); ok {
		return slug, true
	}
	return "", false
}

// GetTenantInfo extracts full tenant info from Fiber context
func GetTenantInfo(c *fiber.Ctx) (*TenantInfo, bool) {
	if info, ok := c.Locals("tenant_info").(*TenantInfo); ok {
		return info, true
	}
	return nil, false
}

// MustGetTenantID extracts tenant ID or panics
func MustGetTenantID(c *fiber.Ctx) uuid.UUID {
	id, ok := GetTenantID(c)
	if !ok {
		panic("tenant ID not found in context")
	}
	return id
}
