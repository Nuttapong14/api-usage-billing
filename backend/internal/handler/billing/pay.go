package billing

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/Nuttapong14/api-usage-billing/internal/api"
	billingsvc "github.com/Nuttapong14/api-usage-billing/internal/service/billing"
)

// PayInvoice handles POST /invoices/{id}/pay.
func (h *Handler) PayInvoice(c *fiber.Ctx) error {
	orgID, customerID, err := h.extractContext(c)
	if err != nil {
		return writeError(c, err)
	}

	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	var req billingsvc.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return api.BadRequest(c, "invalid request body")
	}
	if err := api.Validate(req); err != nil {
		return writeError(c, err)
	}

	result, err := h.service.PayInvoice(c.Context(), billingsvc.PayInvoiceParams{
		OrganizationID: orgID,
		CustomerID:     customerID,
		InvoiceID:      invoiceID,
		Request:        req,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
