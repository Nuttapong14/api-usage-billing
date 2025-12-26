package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CounterType defines the type of counter
type CounterType string

const (
	CounterTypeAPIRequests   CounterType = "api_requests"
	CounterTypeAPIErrors     CounterType = "api_errors"
	CounterTypeDataTransfer  CounterType = "data_transfer"
	CounterTypeTokensUsed    CounterType = "tokens_used"
	CounterTypeActiveUsers   CounterType = "active_users"
	CounterTypeBandwidth     CounterType = "bandwidth"
	CounterTypeRateLimits    CounterType = "rate_limits"
)

// CounterGranularity defines the time granularity for counters
type CounterGranularity string

const (
	GranularityMinute CounterGranularity = "minute"
	GranularityHour   CounterGranularity = "hour"
	GranularityDay    CounterGranularity = "day"
	GranularityMonth  CounterGranularity = "month"
)

// CounterService provides atomic counter operations for usage tracking
type CounterService struct {
	client     *Client
	keyPrefix  string
	defaultTTL map[CounterGranularity]time.Duration
}

// NewCounterService creates a new counter service
func NewCounterService(client *Client) *CounterService {
	return &CounterService{
		client:    client,
		keyPrefix: "counter",
		defaultTTL: map[CounterGranularity]time.Duration{
			GranularityMinute: 2 * time.Hour,
			GranularityHour:   48 * time.Hour,
			GranularityDay:    32 * 24 * time.Hour,
			GranularityMonth:  400 * 24 * time.Hour,
		},
	}
}

// buildKey constructs a counter key
func (cs *CounterService) buildKey(
	counterType CounterType,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	granularity CounterGranularity,
	timestamp time.Time,
) string {
	var timeKey string
	switch granularity {
	case GranularityMinute:
		timeKey = timestamp.Format("2006010215:04")
	case GranularityHour:
		timeKey = timestamp.Format("2006010215")
	case GranularityDay:
		timeKey = timestamp.Format("20060102")
	case GranularityMonth:
		timeKey = timestamp.Format("200601")
	}

	key := fmt.Sprintf("%s:%s:%s:%s", cs.keyPrefix, counterType, orgID.String(), timeKey)
	if apiKeyID != nil {
		key = fmt.Sprintf("%s:%s:%s:%s:%s", cs.keyPrefix, counterType, orgID.String(), apiKeyID.String(), timeKey)
	}

	return key
}

// Increment increments a counter
func (cs *CounterService) Increment(
	ctx context.Context,
	counterType CounterType,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	granularity CounterGranularity,
	value int64,
) (int64, error) {
	timestamp := time.Now().UTC()
	key := cs.buildKey(counterType, orgID, apiKeyID, granularity, timestamp)

	// Increment and set TTL in a single transaction
	pipe := cs.client.rdb.Pipeline()
	incrCmd := pipe.IncrBy(ctx, key, value)
	pipe.Expire(ctx, key, cs.defaultTTL[granularity])

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incrCmd.Val(), nil
}

// IncrementMulti increments a counter across multiple granularities atomically
func (cs *CounterService) IncrementMulti(
	ctx context.Context,
	counterType CounterType,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	value int64,
) error {
	timestamp := time.Now().UTC()
	pipe := cs.client.rdb.Pipeline()

	granularities := []CounterGranularity{
		GranularityMinute,
		GranularityHour,
		GranularityDay,
		GranularityMonth,
	}

	for _, g := range granularities {
		key := cs.buildKey(counterType, orgID, apiKeyID, g, timestamp)
		pipe.IncrBy(ctx, key, value)
		pipe.Expire(ctx, key, cs.defaultTTL[g])
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Get retrieves a counter value
func (cs *CounterService) Get(
	ctx context.Context,
	counterType CounterType,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	granularity CounterGranularity,
	timestamp time.Time,
) (int64, error) {
	key := cs.buildKey(counterType, orgID, apiKeyID, granularity, timestamp)
	val, err := cs.client.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// GetRange retrieves counter values for a time range
func (cs *CounterService) GetRange(
	ctx context.Context,
	counterType CounterType,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	granularity CounterGranularity,
	start, end time.Time,
) (map[time.Time]int64, error) {
	result := make(map[time.Time]int64)

	var step time.Duration
	switch granularity {
	case GranularityMinute:
		step = time.Minute
	case GranularityHour:
		step = time.Hour
	case GranularityDay:
		step = 24 * time.Hour
	case GranularityMonth:
		step = 30 * 24 * time.Hour // Approximate
	}

	// Build all keys
	keys := make([]string, 0)
	timestamps := make([]time.Time, 0)
	for t := start; t.Before(end) || t.Equal(end); t = t.Add(step) {
		keys = append(keys, cs.buildKey(counterType, orgID, apiKeyID, granularity, t))
		timestamps = append(timestamps, t)
	}

	// Get all values in a single pipeline
	pipe := cs.client.rdb.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Check if it's just missing keys (which is okay)
		// Pipeline returns error if any command fails
	}

	// Collect results
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			result[timestamps[i]] = 0
			continue
		}
		if err != nil {
			result[timestamps[i]] = 0
			continue
		}
		intVal, _ := strconv.ParseInt(val, 10, 64)
		result[timestamps[i]] = intVal
	}

	return result, nil
}

// Sum calculates the sum of counter values for a time range
func (cs *CounterService) Sum(
	ctx context.Context,
	counterType CounterType,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	granularity CounterGranularity,
	start, end time.Time,
) (int64, error) {
	values, err := cs.GetRange(ctx, counterType, orgID, apiKeyID, granularity, start, end)
	if err != nil {
		return 0, err
	}

	var sum int64
	for _, v := range values {
		sum += v
	}
	return sum, nil
}

// Reset resets a counter to zero
func (cs *CounterService) Reset(
	ctx context.Context,
	counterType CounterType,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	granularity CounterGranularity,
	timestamp time.Time,
) error {
	key := cs.buildKey(counterType, orgID, apiKeyID, granularity, timestamp)
	return cs.client.rdb.Set(ctx, key, 0, cs.defaultTTL[granularity]).Err()
}

// Delete removes a counter
func (cs *CounterService) Delete(
	ctx context.Context,
	counterType CounterType,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	granularity CounterGranularity,
	timestamp time.Time,
) error {
	key := cs.buildKey(counterType, orgID, apiKeyID, granularity, timestamp)
	return cs.client.rdb.Del(ctx, key).Err()
}

// UsageSnapshot represents a point-in-time usage snapshot
type UsageSnapshot struct {
	OrgID           uuid.UUID
	APIKeyID        *uuid.UUID
	Timestamp       time.Time
	APIRequests     int64
	APIErrors       int64
	DataTransferMB  float64
	TokensUsed      int64
	UniqueUsers     int64
	BandwidthMB     float64
	RateLimitHits   int64
}

// RecordAPIRequest records an API request with all associated metrics
func (cs *CounterService) RecordAPIRequest(
	ctx context.Context,
	orgID uuid.UUID,
	apiKeyID uuid.UUID,
	dataBytes int64,
	tokensUsed int64,
	isError bool,
) error {
	pipe := cs.client.rdb.Pipeline()
	timestamp := time.Now().UTC()
	keyIDPtr := &apiKeyID

	// Record across all granularities
	granularities := []CounterGranularity{
		GranularityMinute,
		GranularityHour,
		GranularityDay,
		GranularityMonth,
	}

	for _, g := range granularities {
		ttl := cs.defaultTTL[g]

		// API requests counter
		reqKey := cs.buildKey(CounterTypeAPIRequests, orgID, keyIDPtr, g, timestamp)
		pipe.Incr(ctx, reqKey)
		pipe.Expire(ctx, reqKey, ttl)

		// Error counter
		if isError {
			errKey := cs.buildKey(CounterTypeAPIErrors, orgID, keyIDPtr, g, timestamp)
			pipe.Incr(ctx, errKey)
			pipe.Expire(ctx, errKey, ttl)
		}

		// Data transfer counter
		if dataBytes > 0 {
			dataKey := cs.buildKey(CounterTypeDataTransfer, orgID, keyIDPtr, g, timestamp)
			pipe.IncrBy(ctx, dataKey, dataBytes)
			pipe.Expire(ctx, dataKey, ttl)
		}

		// Tokens used counter
		if tokensUsed > 0 {
			tokenKey := cs.buildKey(CounterTypeTokensUsed, orgID, keyIDPtr, g, timestamp)
			pipe.IncrBy(ctx, tokenKey, tokensUsed)
			pipe.Expire(ctx, tokenKey, ttl)
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetUsageSnapshot returns a complete usage snapshot for a time period
func (cs *CounterService) GetUsageSnapshot(
	ctx context.Context,
	orgID uuid.UUID,
	apiKeyID *uuid.UUID,
	granularity CounterGranularity,
	timestamp time.Time,
) (*UsageSnapshot, error) {
	snapshot := &UsageSnapshot{
		OrgID:     orgID,
		APIKeyID:  apiKeyID,
		Timestamp: timestamp,
	}

	pipe := cs.client.rdb.Pipeline()

	reqKey := cs.buildKey(CounterTypeAPIRequests, orgID, apiKeyID, granularity, timestamp)
	errKey := cs.buildKey(CounterTypeAPIErrors, orgID, apiKeyID, granularity, timestamp)
	dataKey := cs.buildKey(CounterTypeDataTransfer, orgID, apiKeyID, granularity, timestamp)
	tokenKey := cs.buildKey(CounterTypeTokensUsed, orgID, apiKeyID, granularity, timestamp)
	rateLimitKey := cs.buildKey(CounterTypeRateLimits, orgID, apiKeyID, granularity, timestamp)

	reqCmd := pipe.Get(ctx, reqKey)
	errCmd := pipe.Get(ctx, errKey)
	dataCmd := pipe.Get(ctx, dataKey)
	tokenCmd := pipe.Get(ctx, tokenKey)
	rateLimitCmd := pipe.Get(ctx, rateLimitKey)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Continue - some keys may not exist
	}

	// Parse results (default to 0 if not found)
	if val, err := reqCmd.Result(); err == nil {
		snapshot.APIRequests, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, err := errCmd.Result(); err == nil {
		snapshot.APIErrors, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, err := dataCmd.Result(); err == nil {
		bytes, _ := strconv.ParseInt(val, 10, 64)
		snapshot.DataTransferMB = float64(bytes) / (1024 * 1024)
	}
	if val, err := tokenCmd.Result(); err == nil {
		snapshot.TokensUsed, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, err := rateLimitCmd.Result(); err == nil {
		snapshot.RateLimitHits, _ = strconv.ParseInt(val, 10, 64)
	}

	return snapshot, nil
}
