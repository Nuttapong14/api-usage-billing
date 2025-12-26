package auth

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Context keys for authentication data
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// CustomerIDKey is the context key for customer ID
	CustomerIDKey contextKey = "customer_id"
	// OrganizationIDKey is the context key for organization ID
	OrganizationIDKey contextKey = "organization_id"
	// APIKeyIDKey is the context key for API key ID
	APIKeyIDKey contextKey = "api_key_id"
	// PermissionsKey is the context key for permissions
	PermissionsKey contextKey = "permissions"
	// ScopesKey is the context key for scopes
	ScopesKey contextKey = "scopes"
	// EmailKey is the context key for email
	EmailKey contextKey = "email"
	// NameKey is the context key for name
	NameKey contextKey = "name"
)

// AuthContext provides authentication context utilities
type AuthContext struct {
	UserID         uuid.UUID
	CustomerID     uuid.UUID
	OrganizationID uuid.UUID
	APIKeyID       uuid.UUID
	Email          string
	Name           string
	Permissions    []string
	Scopes         []string
	IsAPIKey       bool
	IsJWT          bool
}

// FromFiberContext extracts authentication context from Fiber context
func FromFiberContext(c *fiber.Ctx) *AuthContext {
	ctx := &AuthContext{}

	// Try JWT authentication first
	if userID, ok := c.Locals("user_uuid").(uuid.UUID); ok {
		ctx.UserID = userID
		ctx.IsJWT = true
	} else if userIDStr, ok := c.Locals("user_id").(string); ok {
		if parsed, err := uuid.Parse(userIDStr); err == nil {
			ctx.UserID = parsed
			ctx.IsJWT = true
		}
	}

	// API Key authentication
	if apiKeyID, ok := c.Locals("api_key_id").(uuid.UUID); ok {
		ctx.APIKeyID = apiKeyID
		ctx.IsAPIKey = true
	}

	// Customer ID
	if customerID, ok := c.Locals("customer_id").(uuid.UUID); ok {
		ctx.CustomerID = customerID
	}

	// Organization ID
	if orgID, ok := c.Locals("organization_id").(uuid.UUID); ok {
		ctx.OrganizationID = orgID
	}

	// Email
	if email, ok := c.Locals("email").(string); ok {
		ctx.Email = email
	}

	// Name
	if name, ok := c.Locals("name").(string); ok {
		ctx.Name = name
	}

	// Permissions
	if perms, ok := c.Locals("permissions").([]string); ok {
		ctx.Permissions = perms
	}

	// Scopes
	if scopes, ok := c.Locals("scopes").([]string); ok {
		ctx.Scopes = scopes
	}

	return ctx
}

// ToContext stores authentication context in a standard context.Context
func (a *AuthContext) ToContext(ctx context.Context) context.Context {
	if a.UserID != uuid.Nil {
		ctx = context.WithValue(ctx, UserIDKey, a.UserID)
	}
	if a.CustomerID != uuid.Nil {
		ctx = context.WithValue(ctx, CustomerIDKey, a.CustomerID)
	}
	if a.OrganizationID != uuid.Nil {
		ctx = context.WithValue(ctx, OrganizationIDKey, a.OrganizationID)
	}
	if a.APIKeyID != uuid.Nil {
		ctx = context.WithValue(ctx, APIKeyIDKey, a.APIKeyID)
	}
	if a.Email != "" {
		ctx = context.WithValue(ctx, EmailKey, a.Email)
	}
	if a.Name != "" {
		ctx = context.WithValue(ctx, NameKey, a.Name)
	}
	if len(a.Permissions) > 0 {
		ctx = context.WithValue(ctx, PermissionsKey, a.Permissions)
	}
	if len(a.Scopes) > 0 {
		ctx = context.WithValue(ctx, ScopesKey, a.Scopes)
	}
	return ctx
}

// IsAuthenticated returns true if the context has valid authentication
func (a *AuthContext) IsAuthenticated() bool {
	return a.IsJWT || a.IsAPIKey
}

// HasOrganization returns true if organization context is set
func (a *AuthContext) HasOrganization() bool {
	return a.OrganizationID != uuid.Nil
}

// HasPermission checks if the context has a specific permission
func (a *AuthContext) HasPermission(permission string) bool {
	for _, p := range a.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasScope checks if the context has a specific scope
func (a *AuthContext) HasScope(scope string) bool {
	for _, s := range a.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the context has any of the specified permissions
func (a *AuthContext) HasAnyPermission(permissions ...string) bool {
	for _, required := range permissions {
		if a.HasPermission(required) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the context has all specified permissions
func (a *AuthContext) HasAllPermissions(permissions ...string) bool {
	for _, required := range permissions {
		if !a.HasPermission(required) {
			return false
		}
	}
	return true
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}

// GetCustomerID extracts customer ID from context
func GetCustomerID(ctx context.Context) (uuid.UUID, bool) {
	if id, ok := ctx.Value(CustomerIDKey).(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}

// GetOrganizationID extracts organization ID from context
func GetOrganizationID(ctx context.Context) (uuid.UUID, bool) {
	if id, ok := ctx.Value(OrganizationIDKey).(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}

// GetAPIKeyID extracts API key ID from context
func GetAPIKeyID(ctx context.Context) (uuid.UUID, bool) {
	if id, ok := ctx.Value(APIKeyIDKey).(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}

// GetPermissions extracts permissions from context
func GetPermissions(ctx context.Context) []string {
	if perms, ok := ctx.Value(PermissionsKey).([]string); ok {
		return perms
	}
	return nil
}

// GetScopes extracts scopes from context
func GetScopes(ctx context.Context) []string {
	if scopes, ok := ctx.Value(ScopesKey).([]string); ok {
		return scopes
	}
	return nil
}

// MustGetOrganizationID extracts organization ID or panics
func MustGetOrganizationID(ctx context.Context) uuid.UUID {
	id, ok := GetOrganizationID(ctx)
	if !ok {
		panic("organization ID not found in context")
	}
	return id
}

// MustGetUserID extracts user ID or panics
func MustGetUserID(ctx context.Context) uuid.UUID {
	id, ok := GetUserID(ctx)
	if !ok {
		panic("user ID not found in context")
	}
	return id
}
