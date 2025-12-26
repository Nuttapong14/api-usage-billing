package analytics

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/auth"
	analyticsvc "github.com/your-org/api-usage-billing/backend/internal/service/analytics"
)

var (
	errMissingContext = errors.New("missing organization context")
)

// Handler handles analytics API requests.
type Handler struct {
	service analyticsvc.Service
}

// NewHandler creates a new analytics handler.
func NewHandler(service analyticsvc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) extractContext(c *fiber.Ctx) (uuid.UUID, error) {
	ctx := auth.FromFiberContext(c)
	orgID := ctx.OrganizationID

	if orgID == uuid.Nil {
		if header := c.Get("X-Organization-ID"); header != "" {
			parsed, err := uuid.Parse(header)
			if err != nil {
				return uuid.Nil, err
			}
			orgID = parsed
		}
	}

	if orgID == uuid.Nil {
		return uuid.Nil, errMissingContext
	}

	return orgID, nil
}

func writeError(c *fiber.Ctx, err error) error {
	var timeErr *time.ParseError

	switch {
	case errors.Is(err, errMissingContext):
		return api.Unauthorized(c, err.Error())
	case isUUIDParseError(err):
		return api.BadRequest(c, "invalid UUID")
	case errors.As(err, &timeErr):
		return api.BadRequest(c, "invalid date format")
	case errors.Is(err, analyticsvc.ErrInvalidDateRange):
		return api.BadRequest(c, err.Error())
	case errors.Is(err, analyticsvc.ErrInvalidGranularity):
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

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func parseOptionalDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := parseDate(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseIntParam(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func normalizeDateRange(start, end *time.Time, defaultDays int) (time.Time, time.Time) {
	now := time.Now().UTC()
	defaultEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	defaultStart := defaultEnd.AddDate(0, 0, -defaultDays)

	if start == nil && end == nil {
		return defaultStart, defaultEnd
	}
	if start != nil && end == nil {
		return *start, defaultEnd
	}
	if start == nil && end != nil {
		endValue := *end
		return endValue.AddDate(0, 0, -defaultDays), endValue
	}
	return *start, *end
}

func resolveGranularity(value string, start, end time.Time) string {
	if value != "" {
		return strings.ToLower(value)
	}
	if end.Sub(start) > 90*24*time.Hour {
		return "month"
	}
	return "day"
}
