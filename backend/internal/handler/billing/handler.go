package billing

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/Nuttapong14/api-usage-billing/internal/api"
	"github.com/Nuttapong14/api-usage-billing/internal/pkg/auth"
	"github.com/Nuttapong14/api-usage-billing/internal/repository"
	billingsvc "github.com/Nuttapong14/api-usage-billing/internal/service/billing"
)

var (
	errMissingContext = errors.New("missing organization or customer context")
)

// Handler handles billing API requests.
type Handler struct {
	service billingsvc.Service
}

// NewHandler creates a new billing handler.
func NewHandler(service billingsvc.Service) *Handler {
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
	case errors.Is(err, repository.ErrInvoiceNotFound):
		return api.NotFound(c, "invoice not found")
	case errors.Is(err, repository.ErrPaymentNotFound):
		return api.NotFound(c, "payment not found")
	case errors.Is(err, billingsvc.ErrInvalidInvoiceFilter):
		return api.BadRequest(c, err.Error())
	case errors.Is(err, billingsvc.ErrInvoiceNotPayable):
		return api.Conflict(c, err.Error())
	case errors.Is(err, billingsvc.ErrPDFNotAvailable):
		return api.ServiceUnavailable(c, "invoice pdf unavailable")
	case errors.Is(err, billingsvc.ErrInvoiceGenerationUnavailable):
		return api.ServiceUnavailable(c, "invoice generation unavailable")
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
