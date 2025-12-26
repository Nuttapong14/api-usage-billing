package apikey

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/Nuttapong14/api-usage-billing/internal/api"
	apikeysvc "github.com/Nuttapong14/api-usage-billing/internal/service/apikey"
)

type revokeAPIKeyRequest struct {
	Reason *string `json:"reason,omitempty" validate:"omitempty,max=200"`
}

// RevokeAPIKey handles DELETE /api-keys/{id}.
func (h *Handler) RevokeAPIKey(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	keyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	var req revokeAPIKeyRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return api.BadRequest(c, "invalid request body")
		}
		if err := api.Validate(req); err != nil {
			return writeError(c, err)
		}
	}

	result, err := h.service.RevokeAPIKey(c.Context(), apikeysvc.RevokeAPIKeyParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		APIKeyID:       keyID,
		Reason:         req.Reason,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
