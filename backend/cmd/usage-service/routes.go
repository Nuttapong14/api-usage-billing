package main

import (
	"github.com/gofiber/fiber/v2"

	apikeyhandler "github.com/Nuttapong14/api-usage-billing/internal/handler/apikey"
	usagehandler "github.com/Nuttapong14/api-usage-billing/internal/handler/usage"
)

// RegisterRoutes wires HTTP routes for the usage service.
func RegisterRoutes(app *fiber.App, handler *usagehandler.Handler, apiKeyHandler *apikeyhandler.Handler) {
	v1 := app.Group("/v1")
	usage := v1.Group("/usage")
	apiKeys := v1.Group("/api-keys")

	usage.Get("/current", handler.GetCurrentUsage)
	usage.Get("/history", handler.GetUsageHistory)
	usage.Get("/breakdown", handler.GetUsageBreakdown)

	apiKeys.Get("/", apiKeyHandler.ListAPIKeys)
	apiKeys.Post("/", apiKeyHandler.CreateAPIKey)
	apiKeys.Post("/:id/rotate", apiKeyHandler.RotateAPIKey)
	apiKeys.Delete("/:id", apiKeyHandler.RevokeAPIKey)
}
