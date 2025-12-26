package analytics

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidDateRange   = errors.New("invalid date range")
	ErrInvalidGranularity = errors.New("invalid granularity")
)

// Money represents a monetary amount.
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// RevenueSnapshot represents revenue for a period.
type RevenueSnapshot struct {
	Amount       Money     `json:"amount"`
	InvoiceCount int64     `json:"invoice_count"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
}

// RevenuePoint represents a point in a revenue trend.
type RevenuePoint struct {
	Period       string    `json:"period"`
	Date         time.Time `json:"date"`
	Amount       Money     `json:"amount"`
	InvoiceCount int64     `json:"invoice_count"`
}

// RevenueReport represents a revenue trend report.
type RevenueReport struct {
	Total       Money          `json:"total"`
	PeriodStart time.Time      `json:"period_start"`
	PeriodEnd   time.Time      `json:"period_end"`
	Granularity string         `json:"granularity"`
	Data        []RevenuePoint `json:"data"`
}

// EndpointStat represents API endpoint usage statistics.
type EndpointStat struct {
	Endpoint     string  `json:"endpoint"`
	Method       string  `json:"method"`
	Requests     int64   `json:"requests"`
	ErrorRate    float64 `json:"error_rate"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
}

// CustomerMetric represents a customer analytics summary.
type CustomerMetric struct {
	CustomerID    uuid.UUID  `json:"customer_id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Revenue       Money      `json:"revenue"`
	RequestCount  int64      `json:"request_count"`
	InvoiceCount  int64      `json:"invoice_count"`
	LastInvoiceAt *time.Time `json:"last_invoice_at,omitempty"`
}

// CustomerRanking represents top customer analytics.
type CustomerRanking struct {
	PeriodStart time.Time        `json:"period_start"`
	PeriodEnd   time.Time        `json:"period_end"`
	Data        []CustomerMetric `json:"data"`
}

// Overview represents the analytics overview response.
type Overview struct {
	RevenueMTD      RevenueSnapshot `json:"revenue_mtd"`
	RequestsToday   int64           `json:"requests_today"`
	ActiveCustomers int64           `json:"active_customers"`
	TopEndpoints    []EndpointStat  `json:"top_endpoints"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// OverviewParams contains overview inputs.
type OverviewParams struct {
	OrganizationID uuid.UUID
	WindowDays     int
}

// RevenueParams contains revenue report inputs.
type RevenueParams struct {
	OrganizationID uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	Granularity    string
}

// CustomerRankingParams contains customer ranking inputs.
type CustomerRankingParams struct {
	OrganizationID uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	Limit          int
}

// Service defines analytics operations.
type Service interface {
	GetOverview(ctx context.Context, params OverviewParams) (*Overview, error)
	GetRevenue(ctx context.Context, params RevenueParams) (*RevenueReport, error)
	GetTopCustomers(ctx context.Context, params CustomerRankingParams) (*CustomerRanking, error)
}
