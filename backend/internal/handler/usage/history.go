package usage

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	usageSvc "github.com/your-org/api-usage-billing/backend/internal/service/usage"
)

// GetUsageHistory handles GET /usage/history.
func (h *Handler) GetUsageHistory(c *fiber.Ctx) error {
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

	granularity := c.Query("granularity", "daily")
	if !isValidGranularity(granularity) {
		return writeError(c, usageSvc.ErrInvalidDateRange)
	}

	limit := parseIntParam(c.Query("limit", "100"), 100)
	offset := parseIntParam(c.Query("offset", "0"), 0)

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

	result, err := h.service.GetUsageHistory(c.Context(), usageSvc.UsageHistoryParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		APIKeyID:       apiKeyID,
		Start:          startDate,
		End:            endDate,
		Granularity:    granularity,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func parseDateParam(c *fiber.Ctx, name string) (time.Time, error) {
	value := c.Query(name, "")
	if value == "" {
		return time.Time{}, usageSvc.ErrInvalidDateRange
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, usageSvc.ErrInvalidDateRange
	}
	return parsed, nil
}

func parseIntParam(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func isValidGranularity(value string) bool {
	switch value {
	case "hourly", "daily", "weekly", "monthly":
		return true
	default:
		return false
	}
}
