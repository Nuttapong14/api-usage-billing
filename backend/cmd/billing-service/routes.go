package main

import (
	"github.com/gofiber/fiber/v2"

	billinghandler "github.com/Nuttapong14/api-usage-billing/internal/handler/billing"
)

// RegisterRoutes wires HTTP routes for the billing service.
func RegisterRoutes(app *fiber.App, handler *billinghandler.Handler) {
	v1 := app.Group("/v1")
	invoices := v1.Group("/invoices")

	invoices.Get("/", handler.ListInvoices)
	invoices.Get("/:id/pdf", handler.GetInvoicePDF)
	invoices.Post("/:id/pay", handler.PayInvoice)
}
