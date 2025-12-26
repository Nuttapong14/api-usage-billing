package main

import (
	"github.com/gofiber/fiber/v2"

	subscriptionhandler "github.com/Nuttapong14/api-usage-billing/internal/handler/subscription"
)

// RegisterRoutes wires HTTP routes for the subscription service.
func RegisterRoutes(app *fiber.App, handler *subscriptionhandler.Handler) {
	v1 := app.Group("/v1")
	subscription := v1.Group("/subscription")

	subscription.Get("/", handler.GetCurrentSubscription)
	subscription.Get("/tiers", handler.ListAvailableTiers)
	subscription.Post("/upgrade", handler.UpgradeSubscription)
	subscription.Post("/downgrade", handler.DowngradeSubscription)
}
