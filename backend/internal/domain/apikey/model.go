package apikey

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// APIKey represents an API key credential for a customer.
type APIKey struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`

	KeyHash   string `gorm:"size:64;not null;uniqueIndex" json:"-"`
	KeyPrefix string `gorm:"size:12;not null;index" json:"key_prefix"`

	Name        *string `gorm:"size:100" json:"name,omitempty"`
	Description *string `gorm:"type:text" json:"description,omitempty"`

	Permissions datatypes.JSON `gorm:"default:'[\"read\"]'" json:"permissions"`
	Scopes      datatypes.JSON `gorm:"default:'[]'" json:"scopes"`

	IPWhitelist    []string `gorm:"type:text[]" json:"ip_whitelist,omitempty"`
	AllowedOrigins []string `gorm:"type:text[]" json:"allowed_origins,omitempty"`

	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP *string    `gorm:"type:inet" json:"last_used_ip,omitempty"`

	RotatedFromID      *uuid.UUID `gorm:"type:uuid" json:"rotated_from_id,omitempty"`
	RotationGraceUntil *time.Time `json:"rotation_grace_until,omitempty"`

	IsActive      bool       `gorm:"default:true" json:"is_active"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	RevokedReason *string    `json:"revoked_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (APIKey) TableName() string {
	return "api_keys"
}
