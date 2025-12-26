package usage

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	usageSvc "github.com/your-org/api-usage-billing/backend/internal/service/usage"
)

// GetUsageBreakdown handles GET /usage/breakdown.
func (h *Handler) GetUsageBreakdown(c *fiber.Ctx) error {
	orgID, customerID, contextAPIKeyID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	startDate, err := parseDateParam(c, "start_date")
	if err != nil {
		return writeError(c, err)
	}
	endDate, err := parseDateParam(c, "end_date")
	if err != nil {
		return writeError(c, err)
	}

	groupBy := c.Query("group_by", "endpoint")
	if !isValidGroupBy(groupBy) {
		return writeError(c, usageSvc.ErrInvalidDateRange)
	}

	var apiKeyID *uuid.UUID
	if queryID := c.Query("api_key_id"); queryID != "" {
		parsed, err := uuid.Parse(queryID)
		if err != nil {
			return writeError(c, err)
		}
		apiKeyID = &parsed
	} else {
		apiKeyID = contextAPIKeyID
	}

	result, err := h.service.GetUsageBreakdown(c.Context(), usageSvc.UsageBreakdownParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		APIKeyID:       apiKeyID,
		Start:          startDate,
		End:            endDate,
		GroupBy:        groupBy,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func isValidGroupBy(value string) bool {
	switch value {
	case "endpoint", "method", "status_code", "api_key":
		return true
	default:
		return false
	}
}
