package subscription

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidBillingCycle = errors.New("invalid billing cycle")
	ErrTierAlreadyActive   = errors.New("tier already active")
	ErrInvalidTierChange   = errors.New("invalid tier change")
)

// Money represents a monetary amount.
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// TierPricing holds pricing details.
type TierPricing struct {
	Monthly Money               `json:"monthly"`
	Yearly  *Money              `json:"yearly,omitempty"`
	Overage *TierOveragePricing `json:"overage,omitempty"`
}

// TierOveragePricing holds overage pricing.
type TierOveragePricing struct {
	PerRequest *Money `json:"per_request,omitempty"`
	PerMB      *Money `json:"per_mb,omitempty"`
}

// TierQuotas defines quotas for a tier.
type TierQuotas struct {
	Requests       *int64 `json:"requests,omitempty"`
	BandwidthMB    *int64 `json:"bandwidth_mb,omitempty"`
	ComputeSeconds *int64 `json:"compute_seconds,omitempty"`
}

// TierRateLimits defines rate limits for a tier.
type TierRateLimits struct {
	PerSecond *int `json:"per_second,omitempty"`
	PerMinute *int `json:"per_minute,omitempty"`
	Burst     *int `json:"burst,omitempty"`
}

// Tier represents a subscription tier for API responses.
type Tier struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  *string                `json:"description,omitempty"`
	Pricing      TierPricing            `json:"pricing"`
	Quotas       TierQuotas             `json:"quotas"`
	RateLimits   TierRateLimits         `json:"rate_limits,omitempty"`
	Features     map[string]interface{} `json:"features,omitempty"`
	SLAPercentage *float64              `json:"sla_percentage,omitempty"`
	IsCurrent    bool                   `json:"is_current,omitempty"`
}

// TierList is the response for tier listing.
type TierList struct {
	Tiers []Tier `json:"tiers"`
}

// UsageQuota represents usage for a single quota.
type UsageQuota struct {
	Used       int64   `json:"used"`
	Limit      *int64  `json:"limit"`
	Percentage float64 `json:"percentage"`
}

// UsageSummary represents usage metrics for quotas.
type UsageSummary struct {
	Requests    UsageQuota `json:"requests"`
	BandwidthMB UsageQuota `json:"bandwidth_mb"`
}

// BillingPeriod represents a billing period.
type BillingPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// PendingChange represents a scheduled subscription change.
type PendingChange struct {
	Type         string    `json:"type"`
	NewTier      *Tier     `json:"new_tier,omitempty"`
	EffectiveDate time.Time `json:"effective_date"`
	CreatedAt    time.Time `json:"created_at"`
}

// Subscription represents a subscription response.
type Subscription struct {
	ID            uuid.UUID     `json:"id"`
	CustomerID    uuid.UUID     `json:"customer_id"`
	Tier          Tier          `json:"tier"`
	Status        string        `json:"status"`
	BillingCycle  string        `json:"billing_cycle"`
	CurrentPeriod BillingPeriod `json:"current_period"`
	Usage         *UsageSummary `json:"usage,omitempty"`
	PendingChange *PendingChange `json:"pending_change,omitempty"`
	TrialEndsAt   *time.Time    `json:"trial_ends_at,omitempty"`
	CancelledAt   *time.Time    `json:"cancelled_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// Proration represents proration details for a change.
type Proration struct {
	Credit *Money `json:"credit,omitempty"`
	Charge *Money `json:"charge,omitempty"`
	Net    *Money `json:"net,omitempty"`
}

// SubscriptionChangeResponse is returned after tier changes.
type SubscriptionChangeResponse struct {
	Subscription Subscription `json:"subscription"`
	ChangeType   string       `json:"change_type"`
	EffectiveDate time.Time    `json:"effective_date"`
	Proration    *Proration   `json:"proration,omitempty"`
	Message      string       `json:"message"`
}

// GetSubscriptionParams contains subscription lookup inputs.
type GetSubscriptionParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
}

// ListTiersParams contains tier listing inputs.
type ListTiersParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	IncludeCurrent bool
}

// ChangeSubscriptionParams contains tier change inputs.
type ChangeSubscriptionParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	TierID         uuid.UUID
	BillingCycle   string
}

// Service defines subscription operations.
type Service interface {
	GetCurrentSubscription(ctx context.Context, params GetSubscriptionParams) (*Subscription, error)
	UpgradeSubscription(ctx context.Context, params ChangeSubscriptionParams) (*SubscriptionChangeResponse, error)
	DowngradeSubscription(ctx context.Context, params ChangeSubscriptionParams) (*SubscriptionChangeResponse, error)
}

// TierService defines tier listing operations.
type TierService interface {
	ListAvailableTiers(ctx context.Context, params ListTiersParams) (*TierList, error)
}
