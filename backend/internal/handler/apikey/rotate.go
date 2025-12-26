package apikey

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	apikeysvc "github.com/your-org/api-usage-billing/backend/internal/service/apikey"
)

type rotateAPIKeyRequest struct {
	GracePeriodHours *int `json:"grace_period_hours,omitempty" validate:"omitempty,min=1,max=168"`
}

// RotateAPIKey handles POST /api-keys/{id}/rotate.
func (h *Handler) RotateAPIKey(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	keyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	var req rotateAPIKeyRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return api.BadRequest(c, "invalid request body")
		}
		if err := api.Validate(req); err != nil {
			return writeError(c, err)
		}
	}

	graceHours := 0
	if req.GracePeriodHours != nil {
		graceHours = *req.GracePeriodHours
	}

	result, err := h.service.RotateAPIKey(c.Context(), apikeysvc.RotateAPIKeyParams{
		OrganizationID:   orgID,
		CustomerID:       customerID,
		APIKeyID:         keyID,
		GracePeriodHours: graceHours,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
