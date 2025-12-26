package billing

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	billingsvc "github.com/Nuttapong14/api-usage-billing/internal/service/billing"
)

// ListInvoices handles GET /invoices.
func (h *Handler) ListInvoices(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	limit := parseIntParam(c.Query("limit", "20"), 20)
	offset := parseIntParam(c.Query("offset", "0"), 0)
	status := c.Query("status", "")
	sortBy := c.Query("sort_by", "issue_date")
	sortDir := c.Query("sort_dir", "desc")

	startDate, err := parseOptionalDate(c.Query("start_date"))
	if err != nil {
		return writeError(c, err)
	}
	endDate, err := parseOptionalDate(c.Query("end_date"))
	if err != nil {
		return writeError(c, err)
	}

	result, err := h.service.ListInvoices(c.Context(), billingsvc.ListInvoicesParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		Status:         status,
		StartDate:      startDate,
		EndDate:        endDate,
		Limit:          limit,
		Offset:         offset,
		SortBy:         sortBy,
		SortDir:        sortDir,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func parseOptionalDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, billingsvc.ErrInvalidInvoiceFilter
	}
	return &parsed, nil
}

func parseIntParam(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
