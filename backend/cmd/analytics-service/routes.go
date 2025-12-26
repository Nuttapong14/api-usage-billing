package main

import (
	"github.com/gofiber/fiber/v2"

	analyticshandler "github.com/Nuttapong14/api-usage-billing/internal/handler/analytics"
)

// RegisterRoutes wires HTTP routes for the analytics service.
func RegisterRoutes(app *fiber.App, handler *analyticshandler.Handler) {
	v1 := app.Group("/v1")
	analytics := v1.Group("/analytics")

	analytics.Get("/overview", handler.GetOverview)
	analytics.Get("/revenue", handler.GetRevenue)
	analytics.Get("/customers", handler.GetTopCustomers)
}
