package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLog represents an immutable audit entry.
type AuditLog struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index:idx_audit_org,priority:1" json:"organization_id"`

	ActorType  string     `gorm:"size:50;not null" json:"actor_type"`
	ActorID    *uuid.UUID `gorm:"type:uuid" json:"actor_id,omitempty"`
	ActorEmail *string    `gorm:"size:255" json:"actor_email,omitempty"`
	ActorIP    *string    `gorm:"type:inet" json:"actor_ip,omitempty"`

	Action       string     `gorm:"size:100;not null" json:"action"`
	ResourceType string     `gorm:"size:50;not null" json:"resource_type"`
	ResourceID   *uuid.UUID `gorm:"type:uuid" json:"resource_id,omitempty"`

	OldValues datatypes.JSON `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues datatypes.JSON `gorm:"type:jsonb" json:"new_values,omitempty"`
	Metadata  datatypes.JSON `gorm:"default:'{}'" json:"metadata"`

	CreatedAt time.Time `gorm:"not null;default:now();index:idx_audit_org,priority:2" json:"created_at"`
}

// TableName specifies the table name for GORM.
func (AuditLog) TableName() string {
	return "audit_logs"
}
