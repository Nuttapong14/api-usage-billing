package webhook

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Nuttapong14/api-usage-billing/internal/api"
	notificationsvc "github.com/Nuttapong14/api-usage-billing/internal/service/notification"
)

type createWebhookRequest struct {
	URL         string   `json:"url" validate:"required,url"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=255"`
	Events      []string `json:"events" validate:"required,min=1,dive,required"`
}

// CreateWebhook handles POST /webhooks.
func (h *Handler) CreateWebhook(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	var req createWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return api.BadRequest(c, "invalid request body")
	}
	if err := api.Validate(req); err != nil {
		return writeError(c, err)
	}

	result, err := h.service.CreateWebhook(c.Context(), notificationsvc.CreateWebhookParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		URL:            req.URL,
		Description:    req.Description,
		Events:         req.Events,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}
