package apikey

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Nuttapong14/api-usage-billing/internal/api"
	apikeysvc "github.com/Nuttapong14/api-usage-billing/internal/service/apikey"
)

type createAPIKeyRequest struct {
	Name           string   `json:"name" validate:"required,min=1,max=100"`
	Description    *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Permissions    []string `json:"permissions,omitempty" validate:"omitempty,dive,oneof=read write admin"`
	Scopes         []string `json:"scopes,omitempty"`
	IPWhitelist    []string `json:"ip_whitelist,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
}

// CreateAPIKey handles POST /api-keys.
func (h *Handler) CreateAPIKey(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	var req createAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return api.BadRequest(c, "invalid request body")
	}
	if err := api.Validate(req); err != nil {
		return writeError(c, err)
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return api.BadRequest(c, "name is required")
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return api.BadRequest(c, "invalid expires_at")
		}
		expiresAt = &parsed
	}

	result, err := h.service.CreateAPIKey(c.Context(), apikeysvc.CreateAPIKeyParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		Name:           name,
		Description:    req.Description,
		Permissions:    req.Permissions,
		Scopes:         req.Scopes,
		IPWhitelist:    req.IPWhitelist,
		AllowedOrigins: req.AllowedOrigins,
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}
