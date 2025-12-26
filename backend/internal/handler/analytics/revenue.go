package analytics

import (
	"github.com/gofiber/fiber/v2"

	analyticsvc "github.com/Nuttapong14/api-usage-billing/internal/service/analytics"
)

// GetRevenue handles GET /analytics/revenue.
func (h *Handler) GetRevenue(c *fiber.Ctx) error {
	orgID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	startDate, err := parseOptionalDate(c.Query("start_date"))
	if err != nil {
		return writeError(c, err)
	}
	endDate, err := parseOptionalDate(c.Query("end_date"))
	if err != nil {
		return writeError(c, err)
	}

	start, end := normalizeDateRange(startDate, endDate, 30)
	granularity := resolveGranularity(c.Query("granularity"), start, end)

	result, err := h.service.GetRevenue(c.Context(), analyticsvc.RevenueParams{
		OrganizationID: orgID,
		StartDate:      start,
		EndDate:        end,
		Granularity:    granularity,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
