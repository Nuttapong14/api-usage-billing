package analytics

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultOverviewWindowDays = 7
)

var revenueStatuses = []string{"pending", "paid", "overdue"}

// ServiceImpl implements analytics operations.
type ServiceImpl struct {
	db    *gorm.DB
	clock func() time.Time
}

// NewService creates a new analytics service.
func NewService(db *gorm.DB) *ServiceImpl {
	return &ServiceImpl{db: db, clock: time.Now}
}

// GetOverview returns the analytics overview.
func (s *ServiceImpl) GetOverview(ctx context.Context, params OverviewParams) (*Overview, error) {
	if params.OrganizationID == uuid.Nil {
		return nil, ErrInvalidDateRange
	}

	now := s.clock().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	revenueSnapshot, err := s.revenueSnapshot(ctx, params.OrganizationID, monthStart, now)
	if err != nil {
		return nil, err
	}

	requestsToday, err := s.requestCount(ctx, params.OrganizationID, startOfDay(now), now)
	if err != nil {
		return nil, err
	}

	activeCustomers, err := s.activeCustomerCount(ctx, params.OrganizationID)
	if err != nil {
		return nil, err
	}

	windowDays := params.WindowDays
	if windowDays <= 0 {
		windowDays = defaultOverviewWindowDays
	}
	windowStart := now.AddDate(0, 0, -windowDays)
	endpoints, err := s.topEndpoints(ctx, params.OrganizationID, windowStart, now, 5)
	if err != nil {
		return nil, err
	}

	return &Overview{
		RevenueMTD:      revenueSnapshot,
		RequestsToday:   requestsToday,
		ActiveCustomers: activeCustomers,
		TopEndpoints:    endpoints,
		UpdatedAt:       now,
	}, nil
}

// GetRevenue returns a revenue trend report.
func (s *ServiceImpl) GetRevenue(ctx context.Context, params RevenueParams) (*RevenueReport, error) {
	if params.OrganizationID == uuid.Nil {
		return nil, ErrInvalidDateRange
	}

	if params.EndDate.Before(params.StartDate) {
		return nil, ErrInvalidDateRange
	}

	granularity := normalizeGranularity(params.Granularity)
	if granularity == "" {
		return nil, ErrInvalidGranularity
	}

	return s.revenueReport(ctx, params.OrganizationID, params.StartDate, params.EndDate, granularity)
}

// GetTopCustomers returns the top customers by revenue.
func (s *ServiceImpl) GetTopCustomers(ctx context.Context, params CustomerRankingParams) (*CustomerRanking, error) {
	if params.OrganizationID == uuid.Nil {
		return nil, ErrInvalidDateRange
	}
	if params.EndDate.Before(params.StartDate) {
		return nil, ErrInvalidDateRange
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	return s.customerRanking(ctx, params.OrganizationID, params.StartDate, params.EndDate, limit)
}

func (s *ServiceImpl) requestCount(ctx context.Context, orgID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	query := `
SELECT COALESCE(COUNT(*), 0)
FROM usage_records
WHERE organization_id = ?
  AND recorded_at >= ?
  AND recorded_at < ?`

	if err := s.db.WithContext(ctx).Raw(query, orgID, start, end).Scan(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *ServiceImpl) activeCustomerCount(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var count int64
	query := `
SELECT COALESCE(COUNT(*), 0)
FROM customers
WHERE organization_id = ?
  AND is_active = true`

	if err := s.db.WithContext(ctx).Raw(query, orgID).Scan(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *ServiceImpl) revenueSnapshot(ctx context.Context, orgID uuid.UUID, start, end time.Time) (RevenueSnapshot, error) {
	row := struct {
		Total        sql.NullFloat64
		InvoiceCount int64
		Currency     sql.NullString
	}{}

	query := `
SELECT
  COALESCE(SUM(total), 0)::float8 AS total,
  COUNT(*) AS invoice_count,
  COALESCE(MAX(currency), 'THB') AS currency
FROM invoices
WHERE organization_id = ?
  AND issue_date >= ?
  AND issue_date <= ?
  AND status IN ?`

	if err := s.db.WithContext(ctx).Raw(query, orgID, start, end, revenueStatuses).Scan(&row).Error; err != nil {
		return RevenueSnapshot{}, err
	}

	amount := 0.0
	if row.Total.Valid {
		amount = row.Total.Float64
	}

	currency := "THB"
	if row.Currency.Valid && row.Currency.String != "" {
		currency = row.Currency.String
	}

	return RevenueSnapshot{
		Amount: Money{
			Amount:   amount,
			Currency: currency,
		},
		InvoiceCount: row.InvoiceCount,
		PeriodStart:  start,
		PeriodEnd:    end,
	}, nil
}

func normalizeGranularity(value string) string {
	switch value {
	case "day", "daily":
		return "day"
	case "month", "monthly":
		return "month"
	case "":
		return ""
	default:
		return ""
	}
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
