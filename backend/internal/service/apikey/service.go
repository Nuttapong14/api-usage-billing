package apikey

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	apikeydomain "github.com/your-org/api-usage-billing/backend/internal/domain/apikey"
	"github.com/your-org/api-usage-billing/backend/internal/infrastructure/redis"
	"github.com/your-org/api-usage-billing/backend/internal/middleware"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/hash"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
)

const (
	defaultGracePeriod = 24 * time.Hour
)

// Config customizes API key service behavior.
type Config struct {
	KeyPrefix          string
	DefaultGracePeriod time.Duration
	MaxKeysPerCustomer int
}

// DefaultConfig returns default service config.
func DefaultConfig() Config {
	return Config{
		KeyPrefix:          hash.DefaultLivePrefix,
		DefaultGracePeriod: defaultGracePeriod,
		MaxKeysPerCustomer: 0,
	}
}

// ServiceImpl implements API key operations.
type ServiceImpl struct {
	repo         repository.APIKeyRepository
	customers    repository.CustomerRepository
	cache        *Cache
	pubsub       *redis.PubSub
	keyPrefix    string
	maxKeys      int
	gracePeriod  time.Duration
	clock        func() time.Time
}

// NewService creates a new API key service.
func NewService(
	repo repository.APIKeyRepository,
	customers repository.CustomerRepository,
	cache *Cache,
	pubsub *redis.PubSub,
	config ...Config,
) *ServiceImpl {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = hash.DefaultLivePrefix
	}
	if cfg.DefaultGracePeriod <= 0 {
		cfg.DefaultGracePeriod = defaultGracePeriod
	}

	return &ServiceImpl{
		repo:        repo,
		customers:   customers,
		cache:       cache,
		pubsub:      pubsub,
		keyPrefix:   cfg.KeyPrefix,
		maxKeys:     cfg.MaxKeysPerCustomer,
		gracePeriod: cfg.DefaultGracePeriod,
		clock:       time.Now,
	}
}

// ListAPIKeys returns API keys for a customer.
func (s *ServiceImpl) ListAPIKeys(ctx context.Context, params ListAPIKeysParams) (*APIKeyList, error) {
	_, err := s.resolveOrganization(ctx, params.OrganizationID, params.CustomerID)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	keys, total, err := s.repo.ListByCustomer(ctx, params.CustomerID, params.IsActive, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]APIKey, 0, len(keys))
	now := s.clock().UTC()
	for i := range keys {
		key := keys[i]
		permissions, err := decodeStringSlice(key.Permissions)
		if err != nil {
			return nil, err
		}
		scopes, err := decodeStringSlice(key.Scopes)
		if err != nil {
			return nil, err
		}
		items = append(items, buildAPIKeyResponse(&key, permissions, scopes, now))
	}

	return &APIKeyList{
		Data: items,
		Pagination: Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: int64(offset+limit) < total,
		},
	}, nil
}

// CreateAPIKey creates a new API key.
func (s *ServiceImpl) CreateAPIKey(ctx context.Context, params CreateAPIKeyParams) (*CreateAPIKeyResponse, error) {
	orgID, err := s.resolveOrganization(ctx, params.OrganizationID, params.CustomerID)
	if err != nil {
		return nil, err
	}

	if s.maxKeys > 0 {
		count, err := s.repo.CountByCustomer(ctx, params.CustomerID)
		if err != nil {
			return nil, err
		}
		if int(count) >= s.maxKeys {
			return nil, ErrAPIKeyLimitReached
		}
	}

	permissions, err := normalizePermissions(params.Permissions)
	if err != nil {
		return nil, err
	}
	scopes := normalizeScopes(params.Scopes)

	plainKey, hashedKey, err := hash.GenerateAPIKey(s.keyPrefix)
	if err != nil {
		return nil, err
	}
	prefix := hash.APIKeyPrefix(plainKey)

	permissionsJSON, err := encodeStringSlice(permissions)
	if err != nil {
		return nil, err
	}
	scopesJSON, err := encodeStringSlice(scopes)
	if err != nil {
		return nil, err
	}

	now := s.clock().UTC()
	key := &apikeydomain.APIKey{
		CustomerID:     params.CustomerID,
		KeyHash:        hashedKey,
		KeyPrefix:      prefix,
		Name:           &params.Name,
		Description:    params.Description,
		Permissions:    permissionsJSON,
		Scopes:         scopesJSON,
		IPWhitelist:    params.IPWhitelist,
		AllowedOrigins: params.AllowedOrigins,
		ExpiresAt:      params.ExpiresAt,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, hashedKey, buildCachedAPIKey(key, orgID, permissions, scopes))
	}

	response := buildAPIKeyResponse(key, permissions, scopes, now)
	return &CreateAPIKeyResponse{
		Key:     plainKey,
		APIKey:  response,
		Warning: "This is the only time the full key will be displayed. Store it securely.",
	}, nil
}

// ValidateKey validates an API key hash for middleware.
func (s *ServiceImpl) ValidateKey(ctx context.Context, keyHash string) (*middleware.APIKeyInfo, error) {
	now := s.clock().UTC()
	if s.cache != nil {
		cached, found, err := s.cache.Get(ctx, keyHash)
		if err == nil && found {
			return buildMiddlewareInfoFromCache(cached, now), nil
		}
	}

	record, err := s.repo.GetByHashWithOrg(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	permissions, err := decodeStringSlice(record.Permissions)
	if err != nil {
		return nil, err
	}
	scopes, err := decodeStringSlice(record.Scopes)
	if err != nil {
		return nil, err
	}

	info := buildMiddlewareInfo(&record.APIKey, record.OrganizationID, permissions, scopes, now)

	if s.cache != nil {
		_ = s.cache.Set(ctx, keyHash, buildCachedAPIKey(&record.APIKey, record.OrganizationID, permissions, scopes))
	}

	return info, nil
}

// RecordUsage updates last used metadata.
func (s *ServiceImpl) RecordUsage(ctx context.Context, keyID uuid.UUID, ip string) error {
	if keyID == uuid.Nil {
		return nil
	}
	return s.repo.RecordUsage(ctx, keyID, ip, s.clock().UTC())
}

func (s *ServiceImpl) resolveOrganization(ctx context.Context, orgID, customerID uuid.UUID) (uuid.UUID, error) {
	if orgID != uuid.Nil {
		customer, err := s.customers.GetByIDForOrg(ctx, orgID, customerID)
		if err != nil {
			return uuid.Nil, err
		}
		return customer.OrganizationID, nil
	}
	customer, err := s.customers.GetByID(ctx, customerID)
	if err != nil {
		return uuid.Nil, err
	}
	return customer.OrganizationID, nil
}

func normalizePermissions(values []string) ([]string, error) {
	if len(values) == 0 {
		return []string{"read"}, nil
	}
	allowed := map[string]bool{"read": true, "write": true, "admin": true}
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		item := strings.ToLower(strings.TrimSpace(value))
		if item == "" {
			continue
		}
		if !allowed[item] {
			return nil, ErrInvalidPermissions
		}
		if !seen[item] {
			seen[item] = true
			normalized = append(normalized, item)
		}
	}
	if len(normalized) == 0 {
		return []string{"read"}, nil
	}
	return normalized, nil
}

func normalizeScopes(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if !seen[item] {
			seen[item] = true
			normalized = append(normalized, item)
		}
	}
	return normalized
}

func encodeStringSlice(values []string) (datatypes.JSON, error) {
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(data), nil
}

func decodeStringSlice(data datatypes.JSON) ([]string, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func buildAPIKeyResponse(
	key *apikeydomain.APIKey,
	permissions, scopes []string,
	now time.Time,
) APIKey {
	permissions = ensurePermissions(permissions)
	name := ""
	if key.Name != nil {
		name = *key.Name
	}
	isRotating := key.RotationGraceUntil != nil && now.Before(*key.RotationGraceUntil) && key.IsActive && key.RevokedAt == nil
	isActive := isKeyActive(key, now)

	return APIKey{
		ID:                 key.ID,
		KeyPrefix:          key.KeyPrefix,
		Name:               name,
		Description:        key.Description,
		Permissions:        permissions,
		Scopes:             scopes,
		IPWhitelist:        key.IPWhitelist,
		AllowedOrigins:     key.AllowedOrigins,
		ExpiresAt:          key.ExpiresAt,
		LastUsedAt:         key.LastUsedAt,
		LastUsedIP:         key.LastUsedIP,
		IsActive:           isActive,
		IsRotating:         isRotating,
		RotationGraceUntil: key.RotationGraceUntil,
		CreatedAt:          key.CreatedAt,
		RevokedAt:          key.RevokedAt,
	}
}

func buildCachedAPIKey(
	key *apikeydomain.APIKey,
	orgID uuid.UUID,
	permissions, scopes []string,
) CachedAPIKey {
	permissions = ensurePermissions(permissions)
	return CachedAPIKey{
		ID:                 key.ID,
		CustomerID:         key.CustomerID,
		OrganizationID:     orgID,
		KeyPrefix:          key.KeyPrefix,
		Permissions:        permissions,
		Scopes:             scopes,
		IPWhitelist:        key.IPWhitelist,
		AllowedOrigins:     key.AllowedOrigins,
		ExpiresAt:          key.ExpiresAt,
		RotationGraceUntil: key.RotationGraceUntil,
		IsActive:           isKeyActive(key, time.Now().UTC()),
	}
}

func buildMiddlewareInfoFromCache(cache *CachedAPIKey, now time.Time) *middleware.APIKeyInfo {
	isActive := cache.IsActive
	if cache.RotationGraceUntil != nil && now.After(*cache.RotationGraceUntil) {
		isActive = false
	}

	return &middleware.APIKeyInfo{
		ID:             cache.ID,
		CustomerID:     cache.CustomerID,
		OrganizationID: cache.OrganizationID,
		KeyPrefix:      cache.KeyPrefix,
		Permissions:    ensurePermissions(cache.Permissions),
		Scopes:         cache.Scopes,
		IPWhitelist:    cache.IPWhitelist,
		AllowedOrigins: cache.AllowedOrigins,
		ExpiresAt:      cache.ExpiresAt,
		IsActive:       isActive,
	}
}

func buildMiddlewareInfo(
	key *apikeydomain.APIKey,
	orgID uuid.UUID,
	permissions, scopes []string,
	now time.Time,
) *middleware.APIKeyInfo {
	return &middleware.APIKeyInfo{
		ID:             key.ID,
		CustomerID:     key.CustomerID,
		OrganizationID: orgID,
		KeyPrefix:      key.KeyPrefix,
		Permissions:    ensurePermissions(permissions),
		Scopes:         scopes,
		IPWhitelist:    key.IPWhitelist,
		AllowedOrigins: key.AllowedOrigins,
		ExpiresAt:      key.ExpiresAt,
		IsActive:       isKeyActive(key, now),
	}
}

func isKeyActive(key *apikeydomain.APIKey, now time.Time) bool {
	if key == nil {
		return false
	}
	if !key.IsActive || key.RevokedAt != nil {
		return false
	}
	if key.RotationGraceUntil != nil && now.After(*key.RotationGraceUntil) {
		return false
	}
	return true
}

func ensurePermissions(values []string) []string {
	if len(values) == 0 {
		return []string{"read"}
	}
	return values
}
