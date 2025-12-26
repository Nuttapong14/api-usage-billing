package main

import (
	"github.com/gofiber/fiber/v2"

	webhookhandler "github.com/your-org/api-usage-billing/backend/internal/handler/webhook"
)

// RegisterRoutes wires HTTP routes for the notification service.
func RegisterRoutes(app *fiber.App, handler *webhookhandler.Handler) {
	v1 := app.Group("/v1")
	webhooks := v1.Group("/webhooks")

	webhooks.Get("/", handler.ListWebhooks)
	webhooks.Post("/", handler.CreateWebhook)
	webhooks.Post("/:id/test", handler.TestWebhook)
}
