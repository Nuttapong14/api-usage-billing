package usage

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	usageSvc "github.com/Nuttapong14/api-usage-billing/internal/service/usage"
)

// GetCurrentUsage handles GET /usage/current.
func (h *Handler) GetCurrentUsage(c *fiber.Ctx) error {
	orgID, customerID, contextAPIKeyID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	var apiKeyID *uuid.UUID
	if queryID := c.Query("api_key_id"); queryID != "" {
		parsed, err := uuid.Parse(queryID)
		if err != nil {
			return writeError(c, err)
		}
		apiKeyID = &parsed
	} else {
		apiKeyID = contextAPIKeyID
	}

	result, err := h.service.GetCurrentUsage(c.Context(), usageSvc.CurrentUsageParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		APIKeyID:       apiKeyID,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
