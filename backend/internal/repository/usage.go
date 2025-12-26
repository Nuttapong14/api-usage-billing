package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/api-usage-billing/backend/internal/domain/subscription"
	"github.com/your-org/api-usage-billing/backend/internal/domain/usage"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrTierNotFound         = errors.New("subscription tier not found")
)

// UsageMetrics holds aggregated usage metrics.
type UsageMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalBandwidthBytes int64
	AvgLatencyMs       *float64
	P95LatencyMs       *float64
	P99LatencyMs       *float64
}

// UsageDataPoint represents a usage summary for a period.
type UsageDataPoint struct {
	Period  time.Time
	Metrics UsageMetrics
}

// UsageBreakdownItem represents grouped usage metrics.
type UsageBreakdownItem struct {
	Key     string
	Metrics UsageMetrics
}

// UsageRepository handles usage data access.
type UsageRepository interface {
	InsertUsageRecords(ctx context.Context, records []usage.UsageRecord) error
	GetUsageSummary(ctx context.Context, orgID, customerID uuid.UUID, apiKeyID *uuid.UUID, start, end time.Time) (UsageMetrics, error)
	GetUsageHistory(ctx context.Context, orgID, customerID uuid.UUID, apiKeyID *uuid.UUID, start, end time.Time, granularity string, limit, offset int) ([]UsageDataPoint, int64, error)
	GetUsageBreakdown(ctx context.Context, orgID, customerID uuid.UUID, apiKeyID *uuid.UUID, start, end time.Time, groupBy string) (UsageMetrics, []UsageBreakdownItem, error)
	GetCurrentSubscription(ctx context.Context, customerID uuid.UUID) (*subscription.Subscription, error)
	GetTierByID(ctx context.Context, tierID uuid.UUID) (*subscription.SubscriptionTier, error)
}

type usageRepository struct {
	*BaseRepository
}

// NewUsageRepository creates a new usage repository.
func NewUsageRepository(db *gorm.DB) UsageRepository {
	return &usageRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *usageRepository) InsertUsageRecords(ctx context.Context, records []usage.UsageRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.DB().WithContext(ctx).CreateInBatches(records, 1000).Error
}

func (r *usageRepository) GetUsageSummary(
	ctx context.Context,
	orgID, customerID uuid.UUID,
	apiKeyID *uuid.UUID,
	start, end time.Time,
) (UsageMetrics, error) {
	query := `
SELECT
	COALESCE(COUNT(*), 0) AS total_requests,
	COALESCE(SUM(CASE WHEN status_code < 400 THEN 1 ELSE 0 END), 0) AS successful_requests,
	COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0) AS failed_requests,
	COALESCE(SUM(request_size_bytes + response_size_bytes), 0) AS total_bandwidth_bytes,
	AVG(latency_ms)::float8 AS avg_latency_ms,
	PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms)::float8 AS p95_latency_ms,
	PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms)::float8 AS p99_latency_ms
FROM usage_records
WHERE organization_id = ?
	AND customer_id = ?
	AND recorded_at >= ?
	AND recorded_at < ?`

	args := []interface{}{orgID, customerID, start, end}
	if apiKeyID != nil {
		query += " AND api_key_id = ?"
		args = append(args, *apiKeyID)
	}

	var row struct {
		TotalRequests      int64
		SuccessfulRequests int64
		FailedRequests     int64
		TotalBandwidthBytes int64
		AvgLatencyMs       sql.NullFloat64
		P95LatencyMs       sql.NullFloat64
		P99LatencyMs       sql.NullFloat64
	}

	if err := r.DB().WithContext(ctx).Raw(query, args...).Scan(&row).Error; err != nil {
		return UsageMetrics{}, err
	}

	return UsageMetrics{
		TotalRequests:      row.TotalRequests,
		SuccessfulRequests: row.SuccessfulRequests,
		FailedRequests:     row.FailedRequests,
		TotalBandwidthBytes: row.TotalBandwidthBytes,
		AvgLatencyMs:       nullFloatToPtr(row.AvgLatencyMs),
		P95LatencyMs:       nullFloatToPtr(row.P95LatencyMs),
		P99LatencyMs:       nullFloatToPtr(row.P99LatencyMs),
	}, nil
}

func (r *usageRepository) GetUsageHistory(
	ctx context.Context,
	orgID, customerID uuid.UUID,
	apiKeyID *uuid.UUID,
	start, end time.Time,
	granularity string,
	limit, offset int,
) ([]UsageDataPoint, int64, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	useRecords := granularity == "hourly"
	periodExpr := "date"
	dateFilter := "date"
	if granularity == "weekly" {
		periodExpr = "date_trunc('week', date)"
	} else if granularity == "monthly" {
		periodExpr = "date_trunc('month', date)"
	}
	if useRecords {
		periodExpr = "date_trunc('hour', recorded_at)"
		dateFilter = "recorded_at"
	}

	var (
		rows []struct {
			Period            time.Time
			TotalRequests     int64
			SuccessfulRequests int64
			FailedRequests    int64
			TotalBandwidthBytes int64
			AvgLatencyMs      sql.NullFloat64
			P95LatencyMs      sql.NullFloat64
			P99LatencyMs      sql.NullFloat64
		}
		total int64
	)

	if useRecords {
		query := fmt.Sprintf(`
SELECT
	%s AS period,
	COALESCE(COUNT(*), 0) AS total_requests,
	COALESCE(SUM(CASE WHEN status_code < 400 THEN 1 ELSE 0 END), 0) AS successful_requests,
	COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0) AS failed_requests,
	COALESCE(SUM(request_size_bytes + response_size_bytes), 0) AS total_bandwidth_bytes,
	AVG(latency_ms)::float8 AS avg_latency_ms,
	PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms)::float8 AS p95_latency_ms,
	PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms)::float8 AS p99_latency_ms
FROM usage_records
WHERE organization_id = ?
	AND customer_id = ?
	AND %s >= ?
	AND %s < ?`, periodExpr, dateFilter, dateFilter)

		args := []interface{}{orgID, customerID, start, end}
		if apiKeyID != nil {
			query += " AND api_key_id = ?"
			args = append(args, *apiKeyID)
		}

		query += " GROUP BY period ORDER BY period LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		if err := r.DB().WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
			return nil, 0, err
		}

		countQuery := fmt.Sprintf(`
SELECT COUNT(*) FROM (
	SELECT 1
	FROM usage_records
	WHERE organization_id = ?
		AND customer_id = ?
		AND %s >= ?
		AND %s < ?`, dateFilter, dateFilter)

		countArgs := []interface{}{orgID, customerID, start, end}
		if apiKeyID != nil {
			countQuery += " AND api_key_id = ?"
			countArgs = append(countArgs, *apiKeyID)
		}

		countQuery += fmt.Sprintf(" GROUP BY %s) AS periods", periodExpr)

		if err := r.DB().WithContext(ctx).Raw(countQuery, countArgs...).Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		query := fmt.Sprintf(`
SELECT
	%s AS period,
	COALESCE(SUM(total_requests), 0) AS total_requests,
	COALESCE(SUM(successful_requests), 0) AS successful_requests,
	COALESCE(SUM(failed_requests), 0) AS failed_requests,
	COALESCE(SUM(total_request_bytes + total_response_bytes), 0) AS total_bandwidth_bytes,
	CASE WHEN SUM(total_requests) > 0 THEN
		SUM(COALESCE(avg_latency_ms, 0) * total_requests) / SUM(total_requests)
	END AS avg_latency_ms,
	NULL::float8 AS p95_latency_ms,
	NULL::float8 AS p99_latency_ms
FROM usage_daily
WHERE organization_id = ?
	AND customer_id = ?
	AND %s >= ?
	AND %s <= ?`, periodExpr, dateFilter, dateFilter)

		args := []interface{}{orgID, customerID, start, end}
		if apiKeyID != nil {
			query += " AND api_key_id = ?"
			args = append(args, *apiKeyID)
		}

		query += " GROUP BY period ORDER BY period LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		if err := r.DB().WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
			return nil, 0, err
		}

		countQuery := fmt.Sprintf(`
SELECT COUNT(*) FROM (
	SELECT 1
	FROM usage_daily
	WHERE organization_id = ?
		AND customer_id = ?
		AND %s >= ?
		AND %s <= ?`, dateFilter, dateFilter)

		countArgs := []interface{}{orgID, customerID, start, end}
		if apiKeyID != nil {
			countQuery += " AND api_key_id = ?"
			countArgs = append(countArgs, *apiKeyID)
		}

		countQuery += fmt.Sprintf(" GROUP BY %s) AS periods", periodExpr)

		if err := r.DB().WithContext(ctx).Raw(countQuery, countArgs...).Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	points := make([]UsageDataPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, UsageDataPoint{
			Period: row.Period,
			Metrics: UsageMetrics{
				TotalRequests:      row.TotalRequests,
				SuccessfulRequests: row.SuccessfulRequests,
				FailedRequests:     row.FailedRequests,
				TotalBandwidthBytes: row.TotalBandwidthBytes,
				AvgLatencyMs:       nullFloatToPtr(row.AvgLatencyMs),
				P95LatencyMs:       nullFloatToPtr(row.P95LatencyMs),
				P99LatencyMs:       nullFloatToPtr(row.P99LatencyMs),
			},
		})
	}

	return points, total, nil
}

func (r *usageRepository) GetUsageBreakdown(
	ctx context.Context,
	orgID, customerID uuid.UUID,
	apiKeyID *uuid.UUID,
	start, end time.Time,
	groupBy string,
) (UsageMetrics, []UsageBreakdownItem, error) {
	groupExpr := "endpoint"
	switch groupBy {
	case "method":
		groupExpr = "method"
	case "status_code":
		groupExpr = "status_code::text"
	case "api_key":
		groupExpr = "COALESCE(api_key_id::text, 'unknown')"
	}

	baseQuery := fmt.Sprintf(`
FROM usage_records
WHERE organization_id = ?
	AND customer_id = ?
	AND recorded_at >= ?
	AND recorded_at < ?`)

	args := []interface{}{orgID, customerID, start, end}
	if apiKeyID != nil {
		baseQuery += " AND api_key_id = ?"
		args = append(args, *apiKeyID)
	}

	breakdownQuery := fmt.Sprintf(`
SELECT
	%s AS group_key,
	COALESCE(COUNT(*), 0) AS total_requests,
	COALESCE(SUM(CASE WHEN status_code < 400 THEN 1 ELSE 0 END), 0) AS successful_requests,
	COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0) AS failed_requests,
	COALESCE(SUM(request_size_bytes + response_size_bytes), 0) AS total_bandwidth_bytes,
	AVG(latency_ms)::float8 AS avg_latency_ms,
	NULL::float8 AS p95_latency_ms,
	NULL::float8 AS p99_latency_ms
%s
GROUP BY group_key
ORDER BY total_requests DESC`, groupExpr, baseQuery)

	var rows []struct {
		GroupKey           string
		TotalRequests      int64
		SuccessfulRequests int64
		FailedRequests     int64
		TotalBandwidthBytes int64
		AvgLatencyMs       sql.NullFloat64
		P95LatencyMs       sql.NullFloat64
		P99LatencyMs       sql.NullFloat64
	}

	if err := r.DB().WithContext(ctx).Raw(breakdownQuery, args...).Scan(&rows).Error; err != nil {
		return UsageMetrics{}, nil, err
	}

	totalMetrics, err := r.GetUsageSummary(ctx, orgID, customerID, apiKeyID, start, end)
	if err != nil {
		return UsageMetrics{}, nil, err
	}

	items := make([]UsageBreakdownItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, UsageBreakdownItem{
			Key: row.GroupKey,
			Metrics: UsageMetrics{
				TotalRequests:      row.TotalRequests,
				SuccessfulRequests: row.SuccessfulRequests,
				FailedRequests:     row.FailedRequests,
				TotalBandwidthBytes: row.TotalBandwidthBytes,
				AvgLatencyMs:       nullFloatToPtr(row.AvgLatencyMs),
				P95LatencyMs:       nullFloatToPtr(row.P95LatencyMs),
				P99LatencyMs:       nullFloatToPtr(row.P99LatencyMs),
			},
		})
	}

	return totalMetrics, items, nil
}

func (r *usageRepository) GetCurrentSubscription(ctx context.Context, customerID uuid.UUID) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	err := r.DB().WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("created_at desc").
		First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *usageRepository) GetTierByID(ctx context.Context, tierID uuid.UUID) (*subscription.SubscriptionTier, error) {
	var tier subscription.SubscriptionTier
	err := r.DB().WithContext(ctx).
		Where("id = ?", tierID).
		First(&tier).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTierNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

func nullFloatToPtr(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	v := value.Float64
	return &v
}
