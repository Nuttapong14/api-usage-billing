package webhook

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// DeliveryStatus represents the delivery lifecycle status.
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
)

// WebhookDelivery represents a webhook delivery attempt.
type WebhookDelivery struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EndpointID     uuid.UUID `gorm:"type:uuid;not null;index" json:"endpoint_id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`

	EventType string    `gorm:"size:100;not null" json:"event_type"`
	EventID   uuid.UUID `gorm:"type:uuid;not null" json:"event_id"`
	Payload   datatypes.JSON `gorm:"type:jsonb;not null" json:"payload"`

	ResponseStatus *int    `json:"response_status,omitempty"`
	ResponseBody   *string `json:"response_body,omitempty"`
	ResponseTimeMs *int    `json:"response_time_ms,omitempty"`

	Attempts    int        `gorm:"default:1" json:"attempts"`
	MaxAttempts int        `gorm:"default:5" json:"max_attempts"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	Status        DeliveryStatus `gorm:"size:20;default:'pending'" json:"status"`
	DeliveredAt   *time.Time     `json:"delivered_at,omitempty"`
	FailedAt      *time.Time     `json:"failed_at,omitempty"`
	FailureReason *string        `json:"failure_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for GORM.
func (WebhookDelivery) TableName() string {
	return "webhook_deliveries"
}
