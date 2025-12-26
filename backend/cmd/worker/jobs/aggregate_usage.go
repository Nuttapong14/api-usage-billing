package jobs

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// AggregateUsageDaily aggregates usage_records into usage_daily for the given date.
func AggregateUsageDaily(ctx context.Context, db *gorm.DB, date time.Time) (int64, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	sql := `
INSERT INTO usage_daily (
	customer_id,
	api_key_id,
	organization_id,
	date,
	total_requests,
	successful_requests,
	failed_requests,
	total_request_bytes,
	total_response_bytes,
	avg_latency_ms,
	min_latency_ms,
	max_latency_ms,
	p50_latency_ms,
	p95_latency_ms,
	p99_latency_ms,
	endpoints_breakdown,
	status_codes_breakdown,
	created_at,
	updated_at
)
SELECT
	customer_id,
	api_key_id,
	organization_id,
	DATE(recorded_at) AS date,
	COUNT(*) AS total_requests,
	SUM(CASE WHEN status_code < 400 THEN 1 ELSE 0 END) AS successful_requests,
	SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) AS failed_requests,
	SUM(request_size_bytes) AS total_request_bytes,
	SUM(response_size_bytes) AS total_response_bytes,
	AVG(latency_ms)::decimal(10,2) AS avg_latency_ms,
	MIN(latency_ms) AS min_latency_ms,
	MAX(latency_ms) AS max_latency_ms,
	PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY latency_ms)::decimal(10,2) AS p50_latency_ms,
	PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms)::decimal(10,2) AS p95_latency_ms,
	PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms)::decimal(10,2) AS p99_latency_ms,
	'{}'::jsonb AS endpoints_breakdown,
	'{}'::jsonb AS status_codes_breakdown,
	NOW(),
	NOW()
FROM usage_records
WHERE recorded_at >= ? AND recorded_at < ?
GROUP BY customer_id, api_key_id, organization_id, DATE(recorded_at)
ON CONFLICT (customer_id, api_key_id, date)
DO UPDATE SET
	total_requests = EXCLUDED.total_requests,
	successful_requests = EXCLUDED.successful_requests,
	failed_requests = EXCLUDED.failed_requests,
	total_request_bytes = EXCLUDED.total_request_bytes,
	total_response_bytes = EXCLUDED.total_response_bytes,
	avg_latency_ms = EXCLUDED.avg_latency_ms,
	min_latency_ms = EXCLUDED.min_latency_ms,
	max_latency_ms = EXCLUDED.max_latency_ms,
	p50_latency_ms = EXCLUDED.p50_latency_ms,
	p95_latency_ms = EXCLUDED.p95_latency_ms,
	p99_latency_ms = EXCLUDED.p99_latency_ms,
	endpoints_breakdown = EXCLUDED.endpoints_breakdown,
	status_codes_breakdown = EXCLUDED.status_codes_breakdown,
	updated_at = NOW();
`

	result := db.WithContext(ctx).Exec(sql, start, end)
	return result.RowsAffected, result.Error
}
