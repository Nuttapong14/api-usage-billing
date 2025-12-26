package usage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// UsageRecord represents a single API request event.
type UsageRecord struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	CustomerID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_usage_customer_time" json:"customer_id"`
	APIKeyID       *uuid.UUID `gorm:"type:uuid" json:"api_key_id,omitempty"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index:idx_usage_org_time" json:"organization_id"`

	RequestID  string `gorm:"size:64;not null;index" json:"request_id"`
	Endpoint   string `gorm:"size:500;not null" json:"endpoint"`
	Method     string `gorm:"size:10;not null" json:"method"`
	StatusCode int16  `gorm:"not null" json:"status_code"`

	RequestSizeBytes  int `gorm:"default:0" json:"request_size_bytes"`
	ResponseSizeBytes int `gorm:"default:0" json:"response_size_bytes"`
	LatencyMs         int `gorm:"not null" json:"latency_ms"`

	RecordedAt time.Time `gorm:"not null;default:now();index:idx_usage_customer_time" json:"recorded_at"`

	UserAgent *string        `gorm:"size:500" json:"user_agent,omitempty"`
	ClientIP  *string        `gorm:"type:inet" json:"client_ip,omitempty"`
	Metadata  datatypes.JSON `gorm:"default:'{}'" json:"metadata"`
}

// TableName specifies the table name for GORM.
func (UsageRecord) TableName() string {
	return "usage_records"
}

// TotalBytes returns the total bytes transferred for the request.
func (r *UsageRecord) TotalBytes() int64 {
	return int64(r.RequestSizeBytes + r.ResponseSizeBytes)
}
