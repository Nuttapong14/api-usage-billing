package subscription

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	subscriptionmodel "github.com/your-org/api-usage-billing/backend/internal/domain/subscription"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
)

// RateLimit defines a rate limit policy.
type RateLimit struct {
	Limit  int
	Window time.Duration
	Burst  int
}

// RateLimitParams contains lookup inputs for rate limits.
type RateLimitParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	APIKeyID       *uuid.UUID
}

// RateLimitResolver resolves rate limits for a request.
type RateLimitResolver interface {
	Resolve(ctx context.Context, params RateLimitParams) (*RateLimit, error)
}

// TierRateLimitResolver resolves rate limits from subscription tiers.
type TierRateLimitResolver struct {
	subscriptions repository.SubscriptionRepository
	tiers         repository.TierRepository
}

// NewTierRateLimitResolver creates a new tier-based resolver.
func NewTierRateLimitResolver(
	subscriptions repository.SubscriptionRepository,
	tiers repository.TierRepository,
) *TierRateLimitResolver {
	return &TierRateLimitResolver{
		subscriptions: subscriptions,
		tiers:         tiers,
	}
}

// Resolve returns the effective rate limit for the customer.
func (r *TierRateLimitResolver) Resolve(ctx context.Context, params RateLimitParams) (*RateLimit, error) {
	if params.CustomerID == uuid.Nil {
		return nil, nil
	}

	sub, err := r.subscriptions.GetByCustomerID(ctx, params.CustomerID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			return nil, nil
		}
		return nil, err
	}

	tier, err := r.tiers.GetByID(ctx, sub.TierID)
	if err != nil {
		if errors.Is(err, repository.ErrTierNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if params.OrganizationID != uuid.Nil && tier.OrganizationID != params.OrganizationID {
		return nil, ErrInvalidTierChange
	}

	return resolveTierRateLimit(tier), nil
}

func resolveTierRateLimit(tier *subscriptionmodel.SubscriptionTier) *RateLimit {
	if tier == nil {
		return nil
	}

	if tier.RateLimitPerSecond != nil && *tier.RateLimitPerSecond > 0 {
		return &RateLimit{
			Limit:  *tier.RateLimitPerSecond,
			Window: time.Second,
			Burst:  intValue(tier.RateLimitBurst),
		}
	}

	if tier.RateLimitPerMinute != nil && *tier.RateLimitPerMinute > 0 {
		return &RateLimit{
			Limit:  *tier.RateLimitPerMinute,
			Window: time.Minute,
			Burst:  intValue(tier.RateLimitBurst),
		}
	}

	return nil
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
