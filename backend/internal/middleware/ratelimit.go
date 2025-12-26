package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/your-org/api-usage-billing/backend/internal/api"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/auth"
	subscriptionsvc "github.com/your-org/api-usage-billing/backend/internal/service/subscription"
)

// RateLimitConfig configures the rate limit middleware.
type RateLimitConfig struct {
	Redis     redis.UniversalClient
	Resolver  subscriptionsvc.RateLimitResolver
	KeyPrefix string
	Clock     func() time.Time
}

// DefaultRateLimitConfig returns default configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		KeyPrefix: "ratelimit",
		Clock:     time.Now,
	}
}

// RateLimitMiddleware enforces tier-based rate limits.
type RateLimitMiddleware struct {
	config RateLimitConfig
}

// NewRateLimitMiddleware creates a new rate limit middleware.
func NewRateLimitMiddleware(config ...RateLimitConfig) *RateLimitMiddleware {
	cfg := DefaultRateLimitConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	return &RateLimitMiddleware{config: cfg}
}

// Handler returns the Fiber middleware handler.
func (m *RateLimitMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if m.config.Redis == nil || m.config.Resolver == nil {
			return c.Next()
		}

		orgID, customerID, apiKeyID := extractRateLimitContext(c)
		if orgID == uuid.Nil || customerID == uuid.Nil {
			return c.Next()
		}

		limit, err := m.config.Resolver.Resolve(c.Context(), subscriptionsvc.RateLimitParams{
			OrganizationID: orgID,
			CustomerID:     customerID,
			APIKeyID:       apiKeyID,
		})
		if err != nil {
			return api.ServiceUnavailable(c, "rate limit unavailable")
		}
		if limit == nil || limit.Limit <= 0 || limit.Window <= 0 {
			return c.Next()
		}

		member := rateLimitMemberFromContext(c)
		allowed, remaining, resetAt, exceeded, err := m.checkLimit(c.Context(), limit, orgID, customerID, apiKeyID, member)
		if err != nil {
			return api.ServiceUnavailable(c, "rate limit unavailable")
		}

		applyRateLimitHeaders(c, allowed, remaining, resetAt)
		if exceeded {
			return api.TooManyRequests(c, "rate limit exceeded")
		}

		return c.Next()
	}
}

func (m *RateLimitMiddleware) checkLimit(
	ctx context.Context,
	limit *subscriptionsvc.RateLimit,
	orgID, customerID uuid.UUID,
	apiKeyID *uuid.UUID,
	member string,
) (int, int, time.Time, bool, error) {
	allowed := limit.Limit + limit.Burst
	if allowed <= 0 {
		return 0, 0, time.Time{}, false, nil
	}

	key := buildRateLimitKey(m.config.KeyPrefix, orgID, customerID, apiKeyID, limit.Window)
	now := m.config.Clock().UTC()
	nowMs := now.UnixMilli()
	cutoff := nowMs - limit.Window.Milliseconds()

	pipe := m.config.Redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", cutoff))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowMs), Member: member})
	countCmd := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, limit.Window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return allowed, 0, time.Time{}, false, err
	}

	count := int(countCmd.Val())
	if count > allowed {
		_ = m.config.Redis.ZRem(ctx, key, member).Err()
		resetAt := m.resetTime(ctx, key, limit.Window, now)
		return allowed, 0, resetAt, true, nil
	}

	remaining := allowed - count
	resetAt := m.resetTime(ctx, key, limit.Window, now)
	return allowed, remaining, resetAt, false, nil
}

func (m *RateLimitMiddleware) resetTime(ctx context.Context, key string, window time.Duration, now time.Time) time.Time {
	oldest, err := m.config.Redis.ZRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil || len(oldest) == 0 {
		return now.Add(window)
	}
	resetMs := int64(oldest[0].Score) + window.Milliseconds()
	return time.UnixMilli(resetMs).UTC()
}

func buildRateLimitKey(prefix string, orgID, customerID uuid.UUID, apiKeyID *uuid.UUID, window time.Duration) string {
	windowKey := fmt.Sprintf("%ds", int(window.Seconds()))
	key := fmt.Sprintf("%s:%s:%s:%s", prefix, orgID.String(), customerID.String(), windowKey)
	if apiKeyID != nil {
		key = fmt.Sprintf("%s:%s:%s:%s:%s", prefix, orgID.String(), customerID.String(), apiKeyID.String(), windowKey)
	}
	return key
}

func applyRateLimitHeaders(c *fiber.Ctx, limit int, remaining int, resetAt time.Time) {
	c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	if !resetAt.IsZero() {
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))
	}
}

func rateLimitMemberFromContext(c *fiber.Ctx) string {
	if id := c.Locals("requestid"); id != nil {
		if value, ok := id.(string); ok && value != "" {
			return fmt.Sprintf("%s:%s", value, uuid.NewString())
		}
	}
	return uuid.NewString()
}

func extractRateLimitContext(c *fiber.Ctx) (uuid.UUID, uuid.UUID, *uuid.UUID) {
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
			if parsed, err := uuid.Parse(header); err == nil {
				orgID = parsed
			}
		}
	}

	if customerID == uuid.Nil {
		if header := c.Get("X-Customer-ID"); header != "" {
			if parsed, err := uuid.Parse(header); err == nil {
				customerID = parsed
			}
		}
	}

	if apiKeyID == nil {
		if header := c.Get("X-API-Key-ID"); header != "" {
			if parsed, err := uuid.Parse(header); err == nil {
				apiKeyID = &parsed
			}
		}
	}

	return orgID, customerID, apiKeyID
}
