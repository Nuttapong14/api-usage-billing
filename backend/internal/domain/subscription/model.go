package subscription

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Status represents subscription lifecycle status.
type Status string

const (
	StatusActive    Status = "active"
	StatusPaused    Status = "paused"
	StatusCancelled Status = "cancelled"
	StatusExpired   Status = "expired"
)

// BillingCycle represents the subscription billing cycle.
type BillingCycle string

const (
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleYearly  BillingCycle = "yearly"
)

// Subscription represents a customer's active subscription to a tier.
type Subscription struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"customer_id"`
	TierID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tier_id"`

	Status       Status       `gorm:"size:20;default:'active'" json:"status"`
	BillingCycle BillingCycle `gorm:"size:20;default:'monthly'" json:"billing_cycle"`

	CurrentPeriodStart time.Time `gorm:"type:date;not null" json:"current_period_start"`
	CurrentPeriodEnd   time.Time `gorm:"type:date;not null" json:"current_period_end"`

	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	PausedAt           *time.Time `json:"paused_at,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason *string    `json:"cancellation_reason,omitempty"`
	UsageResetAt       *time.Time `json:"usage_reset_at,omitempty"`

	Metadata  datatypes.JSON `gorm:"default:'{}'" json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (Subscription) TableName() string {
	return "subscriptions"
}

// BeforeCreate hook to set defaults.
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.Status == "" {
		s.Status = StatusActive
	}
	if s.BillingCycle == "" {
		s.BillingCycle = BillingCycleMonthly
	}
	return nil
}

// IsActive returns true if the subscription is active.
func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive
}
