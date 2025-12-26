package apikey

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	apikeysvc "github.com/your-org/api-usage-billing/backend/internal/service/apikey"
)

// ListAPIKeys handles GET /api-keys.
func (h *Handler) ListAPIKeys(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	limit := parseIntParam(c.Query("limit", "20"), 20)
	offset := parseIntParam(c.Query("offset", "0"), 0)

	var isActive *bool
	if raw := c.Query("is_active"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return api.BadRequest(c, "invalid is_active")
		}
		isActive = &parsed
	}

	result, err := h.service.ListAPIKeys(c.Context(), apikeysvc.ListAPIKeysParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		IsActive:       isActive,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func parseIntParam(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
