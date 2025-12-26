package subscription

import (
	"github.com/gofiber/fiber/v2"

	subscriptionsvc "github.com/your-org/api-usage-billing/backend/internal/service/subscription"
)

// GetCurrentSubscription handles GET /subscription.
func (h *Handler) GetCurrentSubscription(c *fiber.Ctx) error {
	orgID, customerID, _, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	result, err := h.service.GetCurrentSubscription(c.Context(), subscriptionsvc.GetSubscriptionParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
