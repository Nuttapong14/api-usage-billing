package subscription

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	subscriptionsvc "github.com/your-org/api-usage-billing/backend/internal/service/subscription"
)

// ListAvailableTiers handles GET /subscription/tiers.
func (h *Handler) ListAvailableTiers(c *fiber.Ctx) error {
	orgID, customerID, _, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	includeCurrent := true
	if value := c.Query("include_current"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return api.BadRequest(c, "invalid include_current")
		}
		includeCurrent = parsed
	}

	result, err := h.tierService.ListAvailableTiers(c.Context(), subscriptionsvc.ListTiersParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		IncludeCurrent: includeCurrent,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
