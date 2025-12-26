package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrConnectionFailed = errors.New("redis connection failed")
	ErrInvalidValue     = errors.New("invalid value")
)

// Config holds Redis connection configuration
type Config struct {
	// Single node configuration
	Addr     string
	Password string
	DB       int

	// Cluster configuration
	ClusterAddrs []string
	ClusterMode  bool

	// Connection pool settings
	PoolSize     int
	MinIdleConns int
	MaxRetries   int

	// Timeouts
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration

	// TLS
	TLSEnabled bool
}

// DefaultConfig returns default Redis configuration
func DefaultConfig() Config {
	return Config{
		Addr:         "localhost:6379",
		DB:           0,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	}
}

// Client wraps Redis client with additional functionality
type Client struct {
	rdb         redis.UniversalClient
	config      Config
	isCluster   bool
	keyPrefix   string
	defaultTTL  time.Duration
}

// NewClient creates a new Redis client
func NewClient(config Config) (*Client, error) {
	var rdb redis.UniversalClient

	if config.ClusterMode && len(config.ClusterAddrs) > 0 {
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.ClusterAddrs,
			Password:     config.Password,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			MaxRetries:   config.MaxRetries,
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			PoolTimeout:  config.PoolTimeout,
		})
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr:         config.Addr,
			Password:     config.Password,
			DB:           config.DB,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			MaxRetries:   config.MaxRetries,
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			PoolTimeout:  config.PoolTimeout,
		})
	}

	client := &Client{
		rdb:        rdb,
		config:     config,
		isCluster:  config.ClusterMode,
		defaultTTL: 24 * time.Hour,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return client, nil
}

// SetKeyPrefix sets a prefix for all keys
func (c *Client) SetKeyPrefix(prefix string) {
	c.keyPrefix = prefix
}

// SetDefaultTTL sets the default TTL for keys
func (c *Client) SetDefaultTTL(ttl time.Duration) {
	c.defaultTTL = ttl
}

// prefixKey adds the key prefix
func (c *Client) prefixKey(key string) string {
	if c.keyPrefix == "" {
		return key
	}
	return c.keyPrefix + ":" + key
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping checks Redis connectivity
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Raw returns the underlying Redis client for advanced operations
func (c *Client) Raw() redis.UniversalClient {
	return c.rdb
}

// Get retrieves a string value
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, c.prefixKey(key)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotFound
	}
	return val, err
}

// GetJSON retrieves and unmarshals a JSON value
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set stores a string value with optional TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	expiration := c.defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}

	var strVal string
	switch v := value.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		strVal = string(data)
	}

	return c.rdb.Set(ctx, c.prefixKey(key), strVal, expiration).Err()
}

// SetNX sets a value only if the key doesn't exist
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	var strVal string
	switch v := value.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return false, fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		strVal = string(data)
	}

	return c.rdb.SetNX(ctx, c.prefixKey(key), strVal, ttl).Result()
}

// Delete removes keys
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}
	return c.rdb.Del(ctx, prefixedKeys...).Err()
}

// Exists checks if keys exist
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}
	return c.rdb.Exists(ctx, prefixedKeys...).Result()
}

// Expire sets TTL on a key
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, c.prefixKey(key), ttl).Err()
}

// TTL gets remaining TTL of a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, c.prefixKey(key)).Result()
}

// Incr increments a counter
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, c.prefixKey(key)).Result()
}

// IncrBy increments a counter by a value
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.IncrBy(ctx, c.prefixKey(key), value).Result()
}

// IncrByFloat increments a counter by a float value
func (c *Client) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	return c.rdb.IncrByFloat(ctx, c.prefixKey(key), value).Result()
}

// Decr decrements a counter
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Decr(ctx, c.prefixKey(key)).Result()
}

// DecrBy decrements a counter by a value
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.DecrBy(ctx, c.prefixKey(key), value).Result()
}

// HGet gets a hash field value
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := c.rdb.HGet(ctx, c.prefixKey(key), field).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotFound
	}
	return val, err
}

// HSet sets hash field values
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HSet(ctx, c.prefixKey(key), values...).Err()
}

// HGetAll gets all hash fields
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, c.prefixKey(key)).Result()
}

// HIncrBy increments a hash field
func (c *Client) HIncrBy(ctx context.Context, key, field string, value int64) (int64, error) {
	return c.rdb.HIncrBy(ctx, c.prefixKey(key), field, value).Result()
}

// HDel deletes hash fields
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.rdb.HDel(ctx, c.prefixKey(key), fields...).Err()
}

// LPush prepends values to a list
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.LPush(ctx, c.prefixKey(key), values...).Err()
}

// RPush appends values to a list
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.RPush(ctx, c.prefixKey(key), values...).Err()
}

// LRange gets a range of list elements
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.LRange(ctx, c.prefixKey(key), start, stop).Result()
}

// LLen gets list length
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.LLen(ctx, c.prefixKey(key)).Result()
}

// SAdd adds members to a set
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SAdd(ctx, c.prefixKey(key), members...).Err()
}

// SMembers gets all set members
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, c.prefixKey(key)).Result()
}

// SIsMember checks if member exists in set
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.rdb.SIsMember(ctx, c.prefixKey(key), member).Result()
}

// SRem removes members from a set
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SRem(ctx, c.prefixKey(key), members...).Err()
}

// ZAdd adds members to a sorted set
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return c.rdb.ZAdd(ctx, c.prefixKey(key), members...).Err()
}

// ZRange gets sorted set members by rank
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRange(ctx, c.prefixKey(key), start, stop).Result()
}

// ZRangeByScore gets sorted set members by score
func (c *Client) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return c.rdb.ZRangeByScore(ctx, c.prefixKey(key), opt).Result()
}

// ZScore gets score of a member
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	return c.rdb.ZScore(ctx, c.prefixKey(key), member).Result()
}

// ZRem removes members from a sorted set
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.ZRem(ctx, c.prefixKey(key), members...).Err()
}

// Pipeline creates a pipeline for batch operations
func (c *Client) Pipeline() redis.Pipeliner {
	return c.rdb.Pipeline()
}

// TxPipeline creates a transactional pipeline
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.rdb.TxPipeline()
}

// Watch watches keys for a transaction
func (c *Client) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}
	return c.rdb.Watch(ctx, fn, prefixedKeys...)
}

// Scan iterates over keys matching a pattern
func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return c.rdb.Scan(ctx, cursor, c.prefixKey(match), count).Result()
}

// Keys gets all keys matching a pattern (use with caution in production)
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.rdb.Keys(ctx, c.prefixKey(pattern)).Result()
}
