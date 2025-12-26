package customer

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Status represents customer account status
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusClosed    Status = "closed"
)

// Customer represents an API consumer subscribing to an organization's services
type Customer struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	KeycloakUserID    uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"keycloak_user_id"`
	CompanyName       *string        `gorm:"size:255" json:"company_name,omitempty"`
	DisplayName       string         `gorm:"size:255;not null" json:"display_name"`
	Email             string         `gorm:"size:255;not null" json:"email"`
	Phone             *string        `gorm:"size:50" json:"phone,omitempty"`
	TaxID             *string        `gorm:"size:50" json:"tax_id,omitempty"`
	BillingAddress    datatypes.JSON `json:"billing_address,omitempty"`
	Settings          datatypes.JSON `gorm:"default:'{}'" json:"settings"`
	Metadata          datatypes.JSON `gorm:"default:'{}'" json:"metadata"`
	PreferredLanguage string         `gorm:"size:10;default:'th'" json:"preferred_language"`
	PreferredCurrency string         `gorm:"size:3;default:'THB'" json:"preferred_currency"`
	Status            Status         `gorm:"size:20;default:'active'" json:"status"`
	SuspendedReason   *string        `json:"suspended_reason,omitempty"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (Customer) TableName() string {
	return "customers"
}

// BeforeCreate hook to set defaults
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.PreferredLanguage == "" {
		c.PreferredLanguage = "th"
	}
	if c.PreferredCurrency == "" {
		c.PreferredCurrency = "THB"
	}
	if c.Status == "" {
		c.Status = StatusActive
	}
	return nil
}

// Suspend marks the customer as suspended with a reason
func (c *Customer) Suspend(reason string) {
	c.Status = StatusSuspended
	c.SuspendedReason = &reason
	c.IsActive = false
}

// Activate reactivates a suspended customer
func (c *Customer) Activate() {
	c.Status = StatusActive
	c.SuspendedReason = nil
	c.IsActive = true
}

// Close permanently closes the customer account
func (c *Customer) Close() {
	c.Status = StatusClosed
	c.IsActive = false
}
