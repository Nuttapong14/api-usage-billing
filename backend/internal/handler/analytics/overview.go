package analytics

import (
	"github.com/gofiber/fiber/v2"

	analyticsvc "github.com/your-org/api-usage-billing/backend/internal/service/analytics"
)

// GetOverview handles GET /analytics/overview.
func (h *Handler) GetOverview(c *fiber.Ctx) error {
	orgID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	windowDays := parseIntParam(c.Query("window_days", "7"), 7)

	result, err := h.service.GetOverview(c.Context(), analyticsvc.OverviewParams{
		OrganizationID: orgID,
		WindowDays:     windowDays,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
