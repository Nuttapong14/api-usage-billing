package subscription

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	subscriptionsvc "github.com/your-org/api-usage-billing/backend/internal/service/subscription"
)

type upgradeRequest struct {
	TierID       string `json:"tier_id" validate:"required,uuid"`
	BillingCycle string `json:"billing_cycle,omitempty"`
}

// UpgradeSubscription handles POST /subscription/upgrade.
func (h *Handler) UpgradeSubscription(c *fiber.Ctx) error {
	orgID, customerID, _, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	var req upgradeRequest
	if err := c.BodyParser(&req); err != nil {
		return api.BadRequest(c, "invalid request body")
	}
	if err := api.Validate(req); err != nil {
		return writeError(c, err)
	}

	tierID, err := uuid.Parse(req.TierID)
	if err != nil {
		return writeError(c, err)
	}

	result, err := h.service.UpgradeSubscription(c.Context(), subscriptionsvc.ChangeSubscriptionParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		TierID:         tierID,
		BillingCycle:   req.BillingCycle,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
