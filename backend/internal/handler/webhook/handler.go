package webhook

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/auth"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
	notificationsvc "github.com/your-org/api-usage-billing/backend/internal/service/notification"
)

var (
	errMissingContext = errors.New("missing organization or customer context")
)

// Handler handles webhook management requests.
type Handler struct {
	service notificationsvc.Service
}

// NewHandler creates a new webhook handler.
func NewHandler(service notificationsvc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) extractContext(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	ctx := auth.FromFiberContext(c)
	orgID := ctx.OrganizationID
	customerID := ctx.CustomerID

	if orgID == uuid.Nil {
		if header := c.Get("X-Organization-ID"); header != "" {
			parsed, err := uuid.Parse(header)
			if err != nil {
				return uuid.Nil, uuid.Nil, err
			}
			orgID = parsed
		}
	}

	if customerID == uuid.Nil {
		if header := c.Get("X-Customer-ID"); header != "" {
			parsed, err := uuid.Parse(header)
			if err != nil {
				return uuid.Nil, uuid.Nil, err
			}
			customerID = parsed
		}
	}

	if orgID == uuid.Nil || customerID == uuid.Nil {
		return uuid.Nil, uuid.Nil, errMissingContext
	}

	return orgID, customerID, nil
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
	case errors.Is(err, notificationsvc.ErrInvalidContext):
		return api.Unauthorized(c, err.Error())
	case errors.Is(err, repository.ErrWebhookNotFound):
		return api.NotFound(c, "webhook not found")
	case errors.Is(err, notificationsvc.ErrWebhookLimitReached):
		return api.Forbidden(c, err.Error())
	case errors.Is(err, notificationsvc.ErrWebhookInactive):
		return api.BadRequest(c, err.Error())
	case errors.Is(err, notificationsvc.ErrInvalidWebhookEvent):
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
