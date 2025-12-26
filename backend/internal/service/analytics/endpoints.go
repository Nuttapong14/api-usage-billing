package analytics

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func (s *ServiceImpl) topEndpoints(
	ctx context.Context,
	orgID uuid.UUID,
	start, end time.Time,
	limit int,
) ([]EndpointStat, error) {
	if limit <= 0 {
		limit = 5
	}

	rows := []struct {
		Endpoint       string
		Method         string
		TotalRequests  int64
		FailedRequests int64
		AvgLatency     sql.NullFloat64
	}{}

	query := `
SELECT
  endpoint,
  method,
  COUNT(*) AS total_requests,
  COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0) AS failed_requests,
  AVG(latency_ms)::float8 AS avg_latency_ms
FROM usage_records
WHERE organization_id = ?
  AND recorded_at >= ?
  AND recorded_at < ?
GROUP BY endpoint, method
ORDER BY total_requests DESC
LIMIT ?`

	if err := s.db.WithContext(ctx).Raw(query, orgID, start, end, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}

	stats := make([]EndpointStat, 0, len(rows))
	for _, row := range rows {
		avgLatency := 0.0
		if row.AvgLatency.Valid {
			avgLatency = row.AvgLatency.Float64
		}

		errorRate := 0.0
		if row.TotalRequests > 0 {
			errorRate = (float64(row.FailedRequests) / float64(row.TotalRequests)) * 100
		}

		stats = append(stats, EndpointStat{
			Endpoint:     row.Endpoint,
			Method:       row.Method,
			Requests:     row.TotalRequests,
			ErrorRate:    errorRate,
			AvgLatencyMs: avgLatency,
		})
	}

	return stats, nil
}
