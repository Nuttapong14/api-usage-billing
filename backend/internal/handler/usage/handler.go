package usage

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/auth"
	usageSvc "github.com/your-org/api-usage-billing/backend/internal/service/usage"
)

var (
	errMissingContext = errors.New("missing organization or customer context")
)

// Handler handles usage API requests.
type Handler struct {
	service usageSvc.Service
}

// NewHandler creates a new usage handler.
func NewHandler(service usageSvc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) extractContext(c *fiber.Ctx) (uuid.UUID, uuid.UUID, *uuid.UUID, error) {
	ctx := auth.FromFiberContext(c)
	orgID := ctx.OrganizationID
	customerID := ctx.CustomerID
	var apiKeyID *uuid.UUID
	if ctx.APIKeyID != uuid.Nil {
		id := ctx.APIKeyID
		apiKeyID = &id
	}

	if orgID == uuid.Nil {
		if header := c.Get("X-Organization-ID"); header != "" {
			parsed, err := uuid.Parse(header)
			if err != nil {
				return uuid.Nil, uuid.Nil, nil, err
			}
			orgID = parsed
		}
	}

	if customerID == uuid.Nil {
		if header := c.Get("X-Customer-ID"); header != "" {
			parsed, err := uuid.Parse(header)
			if err != nil {
				return uuid.Nil, uuid.Nil, nil, err
			}
			customerID = parsed
		}
	}

	if apiKeyID == nil {
		if header := c.Get("X-API-Key-ID"); header != "" {
			parsed, err := uuid.Parse(header)
			if err != nil {
				return uuid.Nil, uuid.Nil, nil, err
			}
			apiKeyID = &parsed
		}
	}

	if orgID == uuid.Nil || customerID == uuid.Nil {
		return uuid.Nil, uuid.Nil, nil, errMissingContext
	}

	return orgID, customerID, apiKeyID, nil
}

func writeError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errMissingContext):
		return api.Unauthorized(c, err.Error())
	case isUUIDParseError(err):
		return api.BadRequest(c, "invalid UUID")
	case errors.Is(err, usageSvc.ErrInvalidDateRange):
		return api.BadRequest(c, err.Error())
	default:
		return api.InternalError(c, "Failed to process request")
	}
}

func isUUIDParseError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "invalid uuid")
}
