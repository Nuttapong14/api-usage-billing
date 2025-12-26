package usage

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// UsageDaily represents aggregated daily usage metrics.
type UsageDaily struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	CustomerID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_usage_daily_unique,priority:1" json:"customer_id"`
	APIKeyID       *uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_usage_daily_unique,priority:2" json:"api_key_id,omitempty"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	Date           time.Time  `gorm:"type:date;not null;uniqueIndex:idx_usage_daily_unique,priority:3" json:"date"`

	TotalRequests      int64 `gorm:"default:0" json:"total_requests"`
	SuccessfulRequests int64 `gorm:"default:0" json:"successful_requests"`
	FailedRequests     int64 `gorm:"default:0" json:"failed_requests"`
	TotalRequestBytes  int64 `gorm:"default:0" json:"total_request_bytes"`
	TotalResponseBytes int64 `gorm:"default:0" json:"total_response_bytes"`

	AvgLatencyMs *decimal.Decimal `gorm:"type:decimal(10,2)" json:"avg_latency_ms,omitempty"`
	MinLatencyMs *int             `json:"min_latency_ms,omitempty"`
	MaxLatencyMs *int             `json:"max_latency_ms,omitempty"`
	P50LatencyMs *decimal.Decimal `gorm:"type:decimal(10,2)" json:"p50_latency_ms,omitempty"`
	P95LatencyMs *decimal.Decimal `gorm:"type:decimal(10,2)" json:"p95_latency_ms,omitempty"`
	P99LatencyMs *decimal.Decimal `gorm:"type:decimal(10,2)" json:"p99_latency_ms,omitempty"`

	EndpointsBreakdown   datatypes.JSON `gorm:"default:'{}'" json:"endpoints_breakdown"`
	StatusCodesBreakdown datatypes.JSON `gorm:"default:'{}'" json:"status_codes_breakdown"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (UsageDaily) TableName() string {
	return "usage_daily"
}
