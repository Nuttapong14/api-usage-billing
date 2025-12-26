package subscription

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SubscriptionTier represents a configurable pricing plan per organization
type SubscriptionTier struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Slug           string         `gorm:"size:100;not null;uniqueIndex:idx_tier_org_slug" json:"slug"`
	Description    *string        `json:"description,omitempty"`
	DisplayOrder   int            `gorm:"default:0" json:"display_order"`

	// Pricing
	PriceMonthly decimal.Decimal  `gorm:"type:decimal(12,2);not null" json:"price_monthly"`
	PriceYearly  *decimal.Decimal `gorm:"type:decimal(12,2)" json:"price_yearly,omitempty"`
	Currency     string           `gorm:"size:3;default:'THB'" json:"currency"`

	// Quotas (nil = unlimited)
	QuotaRequests       *int64 `json:"quota_requests,omitempty"`
	QuotaBandwidthMB    *int64 `json:"quota_bandwidth_mb,omitempty"`
	QuotaComputeSeconds *int64 `json:"quota_compute_seconds,omitempty"`

	// Rate Limits
	RateLimitPerSecond *int `json:"rate_limit_per_second,omitempty"`
	RateLimitPerMinute *int `json:"rate_limit_per_minute,omitempty"`
	RateLimitBurst     *int `json:"rate_limit_burst,omitempty"`

	// Overage
	OverageEnabled        bool             `gorm:"default:false" json:"overage_enabled"`
	OverageRatePerRequest *decimal.Decimal `gorm:"type:decimal(10,6)" json:"overage_rate_per_request,omitempty"`
	OverageRatePerMB      *decimal.Decimal `gorm:"type:decimal(10,6)" json:"overage_rate_per_mb,omitempty"`

	// Config
	Features      datatypes.JSON   `gorm:"default:'{}'" json:"features"`
	SLAPercentage *decimal.Decimal `gorm:"type:decimal(5,2)" json:"sla_percentage,omitempty"`
	IsPublic      bool             `gorm:"default:true" json:"is_public"`
	IsActive      bool             `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (SubscriptionTier) TableName() string {
	return "subscription_tiers"
}

// BeforeCreate hook to set defaults
func (t *SubscriptionTier) BeforeCreate(tx *gorm.DB) error {
	if t.Currency == "" {
		t.Currency = "THB"
	}
	return nil
}

// HasUnlimitedRequests checks if the tier has no request quota
func (t *SubscriptionTier) HasUnlimitedRequests() bool {
	return t.QuotaRequests == nil
}

// HasUnlimitedBandwidth checks if the tier has no bandwidth quota
func (t *SubscriptionTier) HasUnlimitedBandwidth() bool {
	return t.QuotaBandwidthMB == nil
}

// SupportsOverage checks if overage billing is enabled for this tier
func (t *SubscriptionTier) SupportsOverage() bool {
	return t.OverageEnabled
}

// GetEffectiveRateLimit returns the appropriate rate limit per second
func (t *SubscriptionTier) GetEffectiveRateLimit() int {
	if t.RateLimitPerSecond != nil {
		return *t.RateLimitPerSecond
	}
	if t.RateLimitPerMinute != nil {
		return *t.RateLimitPerMinute / 60
	}
	return 0 // No rate limit
}
