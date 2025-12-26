package billing

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	billingsvc "github.com/your-org/api-usage-billing/backend/internal/service/billing"
)

// GetInvoicePDF handles GET /invoices/{id}/pdf.
func (h *Handler) GetInvoicePDF(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	language := c.Query("language", "th")

	result, err := h.service.GetInvoicePDF(c.Context(), billingsvc.GetInvoicePDFParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		InvoiceID:      invoiceID,
		Language:       language,
	})
	if err != nil {
		return writeError(c, err)
	}

	c.Set("Content-Type", result.ContentType)
	return c.Status(fiber.StatusOK).Send(result.Data)
}
