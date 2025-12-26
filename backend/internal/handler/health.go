package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthHandler provides liveness and readiness endpoints.
type HealthHandler struct {
	Now         func() time.Time
	ReadyChecks []func() error
}

// NewHealthHandler returns a handler with default settings.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		Now: time.Now,
	}
}

// Register attaches health endpoints to the router.
func (h *HealthHandler) Register(router fiber.Router) {
	now := h.Now
	if now == nil {
		now = time.Now
	}

	router.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": now().UTC().Format(time.RFC3339),
		})
	})

	router.Get("/ready", func(c *fiber.Ctx) error {
		for _, check := range h.ReadyChecks {
			if check == nil {
				continue
			}
			if err := check(); err != nil {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"status":    "not_ready",
					"timestamp": now().UTC().Format(time.RFC3339),
					"error":     err.Error(),
				})
			}
		}

		return c.JSON(fiber.Map{
			"status":    "ready",
			"timestamp": now().UTC().Format(time.RFC3339),
		})
	})
}
