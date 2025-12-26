package analytics

import (
	"github.com/gofiber/fiber/v2"

	analyticsvc "github.com/your-org/api-usage-billing/backend/internal/service/analytics"
)

// GetTopCustomers handles GET /analytics/customers.
func (h *Handler) GetTopCustomers(c *fiber.Ctx) error {
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
	limit := parseIntParam(c.Query("limit", "10"), 10)

	result, err := h.service.GetTopCustomers(c.Context(), analyticsvc.CustomerRankingParams{
		OrganizationID: orgID,
		StartDate:      start,
		EndDate:        end,
		Limit:          limit,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
