package webhook

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	notificationsvc "github.com/your-org/api-usage-billing/backend/internal/service/notification"
)

type testWebhookRequest struct {
	EventType string `json:"event_type"`
}

// TestWebhook handles POST /webhooks/{id}/test.
func (h *Handler) TestWebhook(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	webhookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	var req testWebhookRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return api.BadRequest(c, "invalid request body")
		}
	}

	result, err := h.service.SendTestWebhook(c.Context(), notificationsvc.TestWebhookParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		WebhookID:      webhookID,
		EventType:      req.EventType,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
