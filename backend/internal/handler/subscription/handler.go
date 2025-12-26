package subscription

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/auth"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
	subscriptionsvc "github.com/your-org/api-usage-billing/backend/internal/service/subscription"
)

var (
	errMissingContext = errors.New("missing organization or customer context")
)

// Handler handles subscription API requests.
type Handler struct {
	service     subscriptionsvc.Service
	tierService subscriptionsvc.TierService
}

// NewHandler creates a new subscription handler.
func NewHandler(service subscriptionsvc.Service, tierService subscriptionsvc.TierService) *Handler {
	return &Handler{service: service, tierService: tierService}
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
	var validationErr api.ValidationErrors
	switch {
	case errors.Is(err, errMissingContext):
		return api.Unauthorized(c, err.Error())
	case isUUIDParseError(err):
		return api.BadRequest(c, "invalid UUID")
	case errors.As(err, &validationErr):
		return api.ValidationFailed(c, validationErr.Errors)
	case errors.Is(err, repository.ErrSubscriptionNotFound):
		return api.NotFound(c, "subscription not found")
	case errors.Is(err, repository.ErrTierNotFound):
		return api.NotFound(c, "tier not found")
	case errors.Is(err, subscriptionsvc.ErrInvalidBillingCycle):
		return api.BadRequest(c, err.Error())
	case errors.Is(err, subscriptionsvc.ErrTierAlreadyActive):
		return api.BadRequest(c, err.Error())
	case errors.Is(err, subscriptionsvc.ErrInvalidTierChange):
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
