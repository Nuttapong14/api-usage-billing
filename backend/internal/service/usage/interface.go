package usage

import (
	"context"
	"time"

	"github.com/google/uuid"

	usageevent "github.com/your-org/api-usage-billing/backend/internal/event/usage"
)

// UsageMetrics represents usage statistics.
type UsageMetrics struct {
	TotalRequests       int64    `json:"total_requests"`
	SuccessfulRequests  int64    `json:"successful_requests,omitempty"`
	FailedRequests      int64    `json:"failed_requests,omitempty"`
	TotalBandwidthBytes int64    `json:"total_bandwidth_bytes,omitempty"`
	AvgLatencyMs        *float64 `json:"avg_latency_ms,omitempty"`
	P95LatencyMs        *float64 `json:"p95_latency_ms,omitempty"`
	P99LatencyMs        *float64 `json:"p99_latency_ms,omitempty"`
}

// QuotaItem represents quota usage for a resource.
type QuotaItem struct {
	Limit         *int64  `json:"limit"`
	Used          int64   `json:"used"`
	Remaining     *int64  `json:"remaining"`
	PercentageUsed float64 `json:"percentage_used"`
}

// QuotaStatus represents quota usage for the subscription.
type QuotaStatus struct {
	Requests    QuotaItem `json:"requests"`
	BandwidthMB QuotaItem `json:"bandwidth_mb"`
}

// BillingPeriod represents the active billing period.
type BillingPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// CurrentUsage is the response for current usage.
type CurrentUsage struct {
	CustomerID   uuid.UUID   `json:"customer_id"`
	BillingPeriod BillingPeriod `json:"billing_period"`
	Usage        UsageMetrics `json:"usage"`
	Quota        QuotaStatus  `json:"quota"`
	LastUpdated  time.Time    `json:"last_updated"`
}

// UsageDataPoint represents usage metrics for a period.
type UsageDataPoint struct {
	Timestamp time.Time    `json:"timestamp"`
	Period    string       `json:"period"`
	Metrics   UsageMetrics `json:"metrics"`
}

// Pagination represents limit/offset pagination.
type Pagination struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

// UsageHistory is the response for usage history.
type UsageHistory struct {
	Data       []UsageDataPoint `json:"data"`
	Pagination Pagination       `json:"pagination"`
}

// UsageBreakdownItem represents usage metrics for a grouped key.
type UsageBreakdownItem struct {
	Key        string       `json:"key"`
	Metrics    UsageMetrics `json:"metrics"`
	Percentage float64      `json:"percentage"`
}

// UsageBreakdown is the response for usage breakdown.
type UsageBreakdown struct {
	Total     UsageMetrics        `json:"total"`
	Breakdown []UsageBreakdownItem `json:"breakdown"`
}

// CurrentUsageParams describes inputs for current usage retrieval.
type CurrentUsageParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	APIKeyID       *uuid.UUID
}

// UsageHistoryParams describes inputs for usage history retrieval.
type UsageHistoryParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	APIKeyID       *uuid.UUID
	Start          time.Time
	End            time.Time
	Granularity    string
	Limit          int
	Offset         int
}

// UsageBreakdownParams describes inputs for usage breakdown retrieval.
type UsageBreakdownParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	APIKeyID       *uuid.UUID
	Start          time.Time
	End            time.Time
	GroupBy        string
}

// Service defines usage operations.
type Service interface {
	RecordUsage(ctx context.Context, event *usageevent.Event) error
	GetCurrentUsage(ctx context.Context, params CurrentUsageParams) (*CurrentUsage, error)
	GetUsageHistory(ctx context.Context, params UsageHistoryParams) (*UsageHistory, error)
	GetUsageBreakdown(ctx context.Context, params UsageBreakdownParams) (*UsageBreakdown, error)
}
