package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	ErrPermissionDenied     = errors.New("permission denied")
	ErrInvalidResource      = errors.New("invalid resource")
	ErrInvalidAction        = errors.New("invalid action")
	ErrResourceNotOwned     = errors.New("resource not owned by user")
	ErrInsufficientAccess   = errors.New("insufficient access level")
)

// Action represents an action that can be performed on a resource
type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionList   Action = "list"
	ActionManage Action = "manage" // Full access
	ActionAdmin  Action = "admin"  // Administrative operations
)

// Resource represents a resource type
type Resource string

const (
	ResourceOrganization     Resource = "organization"
	ResourceCustomer         Resource = "customer"
	ResourceAPIKey           Resource = "api_key"
	ResourceSubscription     Resource = "subscription"
	ResourceBilling          Resource = "billing"
	ResourceUsage            Resource = "usage"
	ResourceAnalytics        Resource = "analytics"
	ResourceInvoice          Resource = "invoice"
	ResourcePayment          Resource = "payment"
	ResourceSettings         Resource = "settings"
	ResourceUser             Resource = "user"
	ResourceRole             Resource = "role"
	ResourceAuditLog         Resource = "audit_log"
	ResourceWebhook          Resource = "webhook"
	ResourceIntegration      Resource = "integration"
)

// Permission represents a specific permission
type Permission struct {
	Resource Resource
	Action   Action
	Scope    string // Optional scope restriction (e.g., "own", "org", "all")
}

// String returns the string representation of a permission
func (p Permission) String() string {
	if p.Scope != "" {
		return string(p.Resource) + ":" + string(p.Action) + ":" + p.Scope
	}
	return string(p.Resource) + ":" + string(p.Action)
}

// ParsePermission parses a permission string
func ParsePermission(s string) (Permission, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return Permission{}, ErrInvalidResource
	}

	p := Permission{
		Resource: Resource(parts[0]),
		Action:   Action(parts[1]),
	}

	if len(parts) > 2 {
		p.Scope = parts[2]
	}

	return p, nil
}

// PermissionChecker provides permission checking capabilities
type PermissionChecker struct {
	// RolePermissions maps roles to their permissions
	rolePermissions map[string][]Permission

	// ResourceOwnershipResolver resolves resource ownership
	ownershipResolver ResourceOwnershipResolver
}

// ResourceOwnershipResolver resolves ownership of resources
type ResourceOwnershipResolver interface {
	// IsOwner checks if the user owns the resource
	IsOwner(ctx context.Context, userID uuid.UUID, resource Resource, resourceID uuid.UUID) (bool, error)

	// GetResourceOrgID gets the organization ID for a resource
	GetResourceOrgID(ctx context.Context, resource Resource, resourceID uuid.UUID) (uuid.UUID, error)
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(resolver ResourceOwnershipResolver) *PermissionChecker {
	pc := &PermissionChecker{
		rolePermissions:   make(map[string][]Permission),
		ownershipResolver: resolver,
	}

	// Initialize default role permissions
	pc.initializeDefaultRoles()

	return pc
}

// initializeDefaultRoles sets up default role permissions
func (pc *PermissionChecker) initializeDefaultRoles() {
	// Platform admin - full access to everything
	pc.rolePermissions["platform_admin"] = []Permission{
		{Resource: "*", Action: ActionManage},
	}

	// Organization owner - full access within organization
	pc.rolePermissions["org_owner"] = []Permission{
		{Resource: ResourceOrganization, Action: ActionManage, Scope: "org"},
		{Resource: ResourceCustomer, Action: ActionManage, Scope: "org"},
		{Resource: ResourceAPIKey, Action: ActionManage, Scope: "org"},
		{Resource: ResourceSubscription, Action: ActionManage, Scope: "org"},
		{Resource: ResourceBilling, Action: ActionManage, Scope: "org"},
		{Resource: ResourceUsage, Action: ActionRead, Scope: "org"},
		{Resource: ResourceAnalytics, Action: ActionRead, Scope: "org"},
		{Resource: ResourceInvoice, Action: ActionManage, Scope: "org"},
		{Resource: ResourceSettings, Action: ActionManage, Scope: "org"},
		{Resource: ResourceUser, Action: ActionManage, Scope: "org"},
		{Resource: ResourceRole, Action: ActionManage, Scope: "org"},
		{Resource: ResourceAuditLog, Action: ActionRead, Scope: "org"},
		{Resource: ResourceWebhook, Action: ActionManage, Scope: "org"},
		{Resource: ResourceIntegration, Action: ActionManage, Scope: "org"},
	}

	// Organization admin - most access within organization
	pc.rolePermissions["org_admin"] = []Permission{
		{Resource: ResourceOrganization, Action: ActionRead, Scope: "org"},
		{Resource: ResourceOrganization, Action: ActionUpdate, Scope: "org"},
		{Resource: ResourceCustomer, Action: ActionManage, Scope: "org"},
		{Resource: ResourceAPIKey, Action: ActionManage, Scope: "org"},
		{Resource: ResourceSubscription, Action: ActionRead, Scope: "org"},
		{Resource: ResourceBilling, Action: ActionRead, Scope: "org"},
		{Resource: ResourceUsage, Action: ActionRead, Scope: "org"},
		{Resource: ResourceAnalytics, Action: ActionRead, Scope: "org"},
		{Resource: ResourceInvoice, Action: ActionRead, Scope: "org"},
		{Resource: ResourceSettings, Action: ActionUpdate, Scope: "org"},
		{Resource: ResourceUser, Action: ActionManage, Scope: "org"},
		{Resource: ResourceAuditLog, Action: ActionRead, Scope: "org"},
		{Resource: ResourceWebhook, Action: ActionManage, Scope: "org"},
	}

	// Developer - API key and usage focused
	pc.rolePermissions["developer"] = []Permission{
		{Resource: ResourceAPIKey, Action: ActionCreate, Scope: "own"},
		{Resource: ResourceAPIKey, Action: ActionRead, Scope: "own"},
		{Resource: ResourceAPIKey, Action: ActionUpdate, Scope: "own"},
		{Resource: ResourceAPIKey, Action: ActionDelete, Scope: "own"},
		{Resource: ResourceAPIKey, Action: ActionList, Scope: "own"},
		{Resource: ResourceUsage, Action: ActionRead, Scope: "own"},
		{Resource: ResourceAnalytics, Action: ActionRead, Scope: "own"},
	}

	// Billing admin - billing and payment focused
	pc.rolePermissions["billing_admin"] = []Permission{
		{Resource: ResourceBilling, Action: ActionManage, Scope: "org"},
		{Resource: ResourceSubscription, Action: ActionManage, Scope: "org"},
		{Resource: ResourceInvoice, Action: ActionManage, Scope: "org"},
		{Resource: ResourcePayment, Action: ActionManage, Scope: "org"},
		{Resource: ResourceUsage, Action: ActionRead, Scope: "org"},
	}

	// Read-only viewer
	pc.rolePermissions["viewer"] = []Permission{
		{Resource: ResourceOrganization, Action: ActionRead, Scope: "org"},
		{Resource: ResourceCustomer, Action: ActionRead, Scope: "org"},
		{Resource: ResourceAPIKey, Action: ActionRead, Scope: "org"},
		{Resource: ResourceSubscription, Action: ActionRead, Scope: "org"},
		{Resource: ResourceUsage, Action: ActionRead, Scope: "org"},
		{Resource: ResourceAnalytics, Action: ActionRead, Scope: "org"},
		{Resource: ResourceInvoice, Action: ActionRead, Scope: "org"},
	}

	// Customer (end-user of the API)
	pc.rolePermissions["customer"] = []Permission{
		{Resource: ResourceAPIKey, Action: ActionCreate, Scope: "own"},
		{Resource: ResourceAPIKey, Action: ActionRead, Scope: "own"},
		{Resource: ResourceAPIKey, Action: ActionDelete, Scope: "own"},
		{Resource: ResourceAPIKey, Action: ActionList, Scope: "own"},
		{Resource: ResourceUsage, Action: ActionRead, Scope: "own"},
		{Resource: ResourceSubscription, Action: ActionRead, Scope: "own"},
		{Resource: ResourceInvoice, Action: ActionRead, Scope: "own"},
	}
}

// AddRolePermissions adds permissions to a role
func (pc *PermissionChecker) AddRolePermissions(role string, permissions []Permission) {
	pc.rolePermissions[role] = append(pc.rolePermissions[role], permissions...)
}

// GetRolePermissions returns permissions for a role
func (pc *PermissionChecker) GetRolePermissions(role string) []Permission {
	return pc.rolePermissions[role]
}

// Check verifies if the given permissions allow the action on the resource
func (pc *PermissionChecker) Check(
	ctx context.Context,
	authCtx *AuthContext,
	resource Resource,
	action Action,
	resourceID *uuid.UUID,
) error {
	// Combine all permissions from user's permissions list
	allPermissions := make([]Permission, 0)
	for _, permStr := range authCtx.Permissions {
		// Check if it's a role reference
		if rolePerms, ok := pc.rolePermissions[permStr]; ok {
			allPermissions = append(allPermissions, rolePerms...)
		} else {
			// Try to parse as a direct permission
			if perm, err := ParsePermission(permStr); err == nil {
				allPermissions = append(allPermissions, perm)
			}
		}
	}

	// Check each permission
	for _, perm := range allPermissions {
		if pc.permissionMatches(ctx, authCtx, perm, resource, action, resourceID) {
			return nil
		}
	}

	return ErrPermissionDenied
}

// permissionMatches checks if a permission grants access
func (pc *PermissionChecker) permissionMatches(
	ctx context.Context,
	authCtx *AuthContext,
	perm Permission,
	resource Resource,
	action Action,
	resourceID *uuid.UUID,
) bool {
	// Check for wildcard resource
	if perm.Resource != "*" && perm.Resource != resource {
		return false
	}

	// Check action
	if perm.Action != ActionManage && perm.Action != action {
		// Manage action grants all other actions
		return false
	}

	// Check scope
	switch perm.Scope {
	case "", "all":
		return true
	case "org":
		// Must be in the same organization
		if resourceID != nil && pc.ownershipResolver != nil {
			orgID, err := pc.ownershipResolver.GetResourceOrgID(ctx, resource, *resourceID)
			if err != nil {
				return false
			}
			return orgID == authCtx.OrganizationID
		}
		return authCtx.HasOrganization()
	case "own":
		// Must own the resource
		if resourceID != nil && pc.ownershipResolver != nil {
			userID := authCtx.UserID
			if userID == uuid.Nil {
				userID = authCtx.CustomerID
			}
			isOwner, err := pc.ownershipResolver.IsOwner(ctx, userID, resource, *resourceID)
			if err != nil {
				return false
			}
			return isOwner
		}
		return false
	default:
		return false
	}
}

// RequirePermission creates Fiber middleware that requires a specific permission
func RequirePermission(checker *PermissionChecker, resource Resource, action Action) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx := FromFiberContext(c)
		if !authCtx.IsAuthenticated() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "authentication required",
			})
		}

		// Try to get resource ID from path
		var resourceID *uuid.UUID
		if idStr := c.Params("id"); idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				resourceID = &id
			}
		}

		if err := checker.Check(c.Context(), authCtx, resource, action, resourceID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": err.Error(),
			})
		}

		return c.Next()
	}
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func RequireAnyPermission(checker *PermissionChecker, permissions ...Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx := FromFiberContext(c)
		if !authCtx.IsAuthenticated() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "authentication required",
			})
		}

		var resourceID *uuid.UUID
		if idStr := c.Params("id"); idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				resourceID = &id
			}
		}

		for _, perm := range permissions {
			if err := checker.Check(c.Context(), authCtx, perm.Resource, perm.Action, resourceID); err == nil {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "forbidden",
			"message": ErrPermissionDenied.Error(),
		})
	}
}

// RequireOwnership creates middleware that requires ownership of a resource
func RequireOwnership(resolver ResourceOwnershipResolver, resource Resource) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx := FromFiberContext(c)
		if !authCtx.IsAuthenticated() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "authentication required",
			})
		}

		idStr := c.Params("id")
		if idStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "bad_request",
				"message": "resource ID required",
			})
		}

		resourceID, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "bad_request",
				"message": "invalid resource ID",
			})
		}

		userID := authCtx.UserID
		if userID == uuid.Nil {
			userID = authCtx.CustomerID
		}

		isOwner, err := resolver.IsOwner(c.Context(), userID, resource, resourceID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "failed to verify ownership",
			})
		}

		if !isOwner {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": ErrResourceNotOwned.Error(),
			})
		}

		return c.Next()
	}
}
