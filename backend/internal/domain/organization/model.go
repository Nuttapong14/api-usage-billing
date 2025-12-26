package organization

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Organization represents an API provider company (multi-tenant root entity)
type Organization struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name            string         `gorm:"size:255;not null" json:"name"`
	Slug            string         `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	LogoURL         *string        `gorm:"size:500" json:"logo_url,omitempty"`
	Settings        datatypes.JSON `gorm:"default:'{}'" json:"settings"`
	BillingEmail    *string        `gorm:"size:255" json:"billing_email,omitempty"`
	BillingAddress  datatypes.JSON `json:"billing_address,omitempty"`
	TaxID           *string        `gorm:"size:50" json:"tax_id,omitempty"`
	DefaultCurrency string         `gorm:"size:3;default:'THB'" json:"default_currency"`
	Timezone        string         `gorm:"size:50;default:'Asia/Bangkok'" json:"timezone"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (Organization) TableName() string {
	return "organizations"
}

// BeforeCreate hook to ensure slug and timezone are set
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.Timezone == "" {
		o.Timezone = "Asia/Bangkok"
	}
	if o.DefaultCurrency == "" {
		o.DefaultCurrency = "THB"
	}
	return nil
}
