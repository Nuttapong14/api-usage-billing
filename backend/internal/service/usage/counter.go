package usage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	counterPrefix     = "usage:counter"
	metricRequests    = "requests"
	metricSuccess     = "requests_success"
	metricFailed      = "requests_failed"
	metricBytes       = "bytes"
	metricLatencySum  = "latency_sum"
	metricLatencyCount = "latency_count"
)

// CounterStore manages real-time usage counters in Redis.
type CounterStore struct {
	rdb redis.UniversalClient
	ttl time.Duration
}

// NewCounterStore creates a new CounterStore.
func NewCounterStore(rdb redis.UniversalClient, ttl ...time.Duration) *CounterStore {
	counterTTL := 35 * 24 * time.Hour
	if len(ttl) > 0 {
		counterTTL = ttl[0]
	}
	return &CounterStore{rdb: rdb, ttl: counterTTL}
}

// Increment updates counters for a usage event.
func (c *CounterStore) Increment(
	ctx context.Context,
	orgID, customerID uuid.UUID,
	apiKeyID *uuid.UUID,
	recordedAt time.Time,
	statusCode int,
	requestSizeBytes, responseSizeBytes int,
	latencyMs int,
) error {
	monthKey := recordedAt.UTC().Format("2006-01")
	keys := []string{
		c.buildKey(metricRequests, orgID, customerID, apiKeyID, monthKey),
		c.buildKey(metricSuccess, orgID, customerID, apiKeyID, monthKey),
		c.buildKey(metricFailed, orgID, customerID, apiKeyID, monthKey),
		c.buildKey(metricBytes, orgID, customerID, apiKeyID, monthKey),
		c.buildKey(metricLatencySum, orgID, customerID, apiKeyID, monthKey),
		c.buildKey(metricLatencyCount, orgID, customerID, apiKeyID, monthKey),
	}

	pipe := c.rdb.Pipeline()
	pipe.IncrBy(ctx, keys[0], 1)
	if statusCode > 0 && statusCode < 400 {
		pipe.IncrBy(ctx, keys[1], 1)
	} else {
		pipe.IncrBy(ctx, keys[2], 1)
	}
	pipe.IncrBy(ctx, keys[3], int64(requestSizeBytes+responseSizeBytes))
	pipe.IncrBy(ctx, keys[4], int64(latencyMs))
	pipe.IncrBy(ctx, keys[5], 1)
	for _, key := range keys {
		pipe.Expire(ctx, key, c.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetMetrics returns counters summed over the given time range.
func (c *CounterStore) GetMetrics(
	ctx context.Context,
	orgID, customerID uuid.UUID,
	apiKeyID *uuid.UUID,
	start, end time.Time,
) (UsageMetrics, error) {
	months := monthsBetween(start, end)
	if len(months) == 0 {
		return UsageMetrics{}, nil
	}

	buildKeys := func(metric string) []string {
		keys := make([]string, 0, len(months))
		for _, month := range months {
			keys = append(keys, c.buildKey(metric, orgID, customerID, apiKeyID, month.Format("2006-01")))
		}
		return keys
	}

	totalRequests, err := c.sumKeys(ctx, buildKeys(metricRequests))
	if err != nil {
		return UsageMetrics{}, err
	}
	successRequests, err := c.sumKeys(ctx, buildKeys(metricSuccess))
	if err != nil {
		return UsageMetrics{}, err
	}
	failedRequests, err := c.sumKeys(ctx, buildKeys(metricFailed))
	if err != nil {
		return UsageMetrics{}, err
	}
	totalBytes, err := c.sumKeys(ctx, buildKeys(metricBytes))
	if err != nil {
		return UsageMetrics{}, err
	}
	latencySum, err := c.sumKeys(ctx, buildKeys(metricLatencySum))
	if err != nil {
		return UsageMetrics{}, err
	}
	latencyCount, err := c.sumKeys(ctx, buildKeys(metricLatencyCount))
	if err != nil {
		return UsageMetrics{}, err
	}

	var avgLatency *float64
	if latencyCount > 0 {
		avg := float64(latencySum) / float64(latencyCount)
		avgLatency = &avg
	}

	return UsageMetrics{
		TotalRequests:       totalRequests,
		SuccessfulRequests:  successRequests,
		FailedRequests:      failedRequests,
		TotalBandwidthBytes: totalBytes,
		AvgLatencyMs:        avgLatency,
	}, nil
}

func (c *CounterStore) buildKey(metric string, orgID, customerID uuid.UUID, apiKeyID *uuid.UUID, month string) string {
	if apiKeyID != nil {
		return fmt.Sprintf("%s:%s:%s:%s:%s:%s", counterPrefix, orgID.String(), customerID.String(), apiKeyID.String(), month, metric)
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s", counterPrefix, orgID.String(), customerID.String(), month, metric)
}

func (c *CounterStore) sumKeys(ctx context.Context, keys []string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	values, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return 0, err
	}

	var sum int64
	for _, value := range values {
		if value == nil {
			continue
		}
		switch v := value.(type) {
		case int64:
			sum += v
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				sum += parsed
			}
		case []byte:
			if parsed, err := strconv.ParseInt(string(v), 10, 64); err == nil {
				sum += parsed
			}
		}
	}

	return sum, nil
}

func monthsBetween(start, end time.Time) []time.Time {
	if end.Before(start) {
		return nil
	}

	startMonth := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)

	months := make([]time.Time, 0)
	for current := startMonth; !current.After(endMonth); current = current.AddDate(0, 1, 0) {
		months = append(months, current)
	}

	return months
}
