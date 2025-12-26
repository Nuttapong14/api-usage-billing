package apikey

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	apiKeyCachePrefix = "apikey:"
	defaultCacheTTL   = 5 * time.Minute
)

// CachedAPIKey represents cached key metadata for validation.
type CachedAPIKey struct {
	ID                 uuid.UUID
	CustomerID         uuid.UUID
	OrganizationID     uuid.UUID
	KeyPrefix          string
	Permissions        []string
	Scopes             []string
	IPWhitelist        []string
	AllowedOrigins     []string
	ExpiresAt          *time.Time
	RotationGraceUntil *time.Time
	IsActive           bool
}

// Cache stores API key validation data in Redis.
type Cache struct {
	rdb redis.UniversalClient
	ttl time.Duration
}

// NewCache creates a new API key cache.
func NewCache(rdb redis.UniversalClient, ttl ...time.Duration) *Cache {
	cacheTTL := defaultCacheTTL
	if len(ttl) > 0 {
		cacheTTL = ttl[0]
	}
	return &Cache{rdb: rdb, ttl: cacheTTL}
}

func (c *Cache) cacheKey(hash string) string {
	return apiKeyCachePrefix + hash
}

// Get retrieves cached key metadata.
func (c *Cache) Get(ctx context.Context, hash string) (*CachedAPIKey, bool, error) {
	data, err := c.rdb.HGetAll(ctx, c.cacheKey(hash)).Result()
	if err != nil {
		return nil, false, err
	}
	if len(data) == 0 {
		return nil, false, nil
	}

	keyID, err := uuid.Parse(data["id"])
	if err != nil {
		return nil, false, err
	}
	customerID, err := uuid.Parse(data["customer_id"])
	if err != nil {
		return nil, false, err
	}
	orgID, err := uuid.Parse(data["organization_id"])
	if err != nil {
		return nil, false, err
	}

	permissions, err := decodeCacheSlice(data["permissions"])
	if err != nil {
		return nil, false, err
	}
	scopes, err := decodeCacheSlice(data["scopes"])
	if err != nil {
		return nil, false, err
	}
	ipWhitelist, err := decodeCacheSlice(data["ip_whitelist"])
	if err != nil {
		return nil, false, err
	}
	allowedOrigins, err := decodeCacheSlice(data["allowed_origins"])
	if err != nil {
		return nil, false, err
	}

	expiresAt, err := parseCacheTime(data["expires_at"])
	if err != nil {
		return nil, false, err
	}
	rotationGraceUntil, err := parseCacheTime(data["rotation_grace_until"])
	if err != nil {
		return nil, false, err
	}

	isActive := parseCacheBool(data["is_active"])

	return &CachedAPIKey{
		ID:                 keyID,
		CustomerID:         customerID,
		OrganizationID:     orgID,
		KeyPrefix:          data["key_prefix"],
		Permissions:        permissions,
		Scopes:             scopes,
		IPWhitelist:        ipWhitelist,
		AllowedOrigins:     allowedOrigins,
		ExpiresAt:          expiresAt,
		RotationGraceUntil: rotationGraceUntil,
		IsActive:           isActive,
	}, true, nil
}

// Set stores cached key metadata.
func (c *Cache) Set(ctx context.Context, hash string, key CachedAPIKey) error {
	permissions, err := encodeCacheSlice(key.Permissions)
	if err != nil {
		return err
	}
	scopes, err := encodeCacheSlice(key.Scopes)
	if err != nil {
		return err
	}
	ipWhitelist, err := encodeCacheSlice(key.IPWhitelist)
	if err != nil {
		return err
	}
	allowedOrigins, err := encodeCacheSlice(key.AllowedOrigins)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"id":                  key.ID.String(),
		"customer_id":         key.CustomerID.String(),
		"organization_id":     key.OrganizationID.String(),
		"key_prefix":          key.KeyPrefix,
		"permissions":         permissions,
		"scopes":              scopes,
		"ip_whitelist":        ipWhitelist,
		"allowed_origins":     allowedOrigins,
		"expires_at":          formatCacheTime(key.ExpiresAt),
		"rotation_grace_until": formatCacheTime(key.RotationGraceUntil),
		"is_active":           strconv.FormatBool(key.IsActive),
	}

	if err := c.rdb.HSet(ctx, c.cacheKey(hash), fields).Err(); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, c.cacheKey(hash), c.ttl).Err()
}

// Delete removes cached key metadata.
func (c *Cache) Delete(ctx context.Context, hash string) error {
	return c.rdb.Del(ctx, c.cacheKey(hash)).Err()
}

func encodeCacheSlice(values []string) (string, error) {
	if len(values) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeCacheSlice(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func formatCacheTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func parseCacheTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseCacheBool(raw string) bool {
	if raw == "" {
		return false
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return parsed
}
