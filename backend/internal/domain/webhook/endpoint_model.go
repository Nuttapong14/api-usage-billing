package webhook

import (
	"time"

	"github.com/google/uuid"
)

// WebhookEndpoint represents a customer-configured webhook endpoint.
type WebhookEndpoint struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerID     uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`

	URL    string   `gorm:"size:500;not null" json:"url"`
	Secret string   `gorm:"size:100;not null" json:"-"`
	Events []string `gorm:"type:text[];not null" json:"events"`

	Description *string `gorm:"size:255" json:"description,omitempty"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`

	LastTriggeredAt     *time.Time `json:"last_triggered_at,omitempty"`
	ConsecutiveFailures int        `gorm:"default:0" json:"consecutive_failures"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (WebhookEndpoint) TableName() string {
	return "webhook_endpoints"
}
