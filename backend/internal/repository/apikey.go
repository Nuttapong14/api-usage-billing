package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Nuttapong14/api-usage-billing/internal/domain/apikey"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
)

// APIKeyWithOrganization includes organization data for a key lookup.
type APIKeyWithOrganization struct {
	apikey.APIKey
	OrganizationID uuid.UUID `gorm:"column:organization_id"`
}

// APIKeyRepository handles API key data access.
type APIKeyRepository interface {
	Create(ctx context.Context, key *apikey.APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*apikey.APIKey, error)
	GetByIDForCustomer(ctx context.Context, customerID, id uuid.UUID) (*apikey.APIKey, error)
	GetByHash(ctx context.Context, hash string) (*apikey.APIKey, error)
	GetByHashWithOrg(ctx context.Context, hash string) (*APIKeyWithOrganization, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID, isActive *bool, limit, offset int) ([]apikey.APIKey, int64, error)
	Update(ctx context.Context, key *apikey.APIKey) error
	RecordUsage(ctx context.Context, keyID uuid.UUID, ip string, usedAt time.Time) error
	CountByCustomer(ctx context.Context, customerID uuid.UUID) (int64, error)
}

type apiKeyRepository struct {
	*BaseRepository
}

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *apiKeyRepository) Create(ctx context.Context, key *apikey.APIKey) error {
	return r.DB().WithContext(ctx).Create(key).Error
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*apikey.APIKey, error) {
	var key apikey.APIKey
	err := r.DB().WithContext(ctx).
		Where("id = ?", id).
		First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) GetByIDForCustomer(ctx context.Context, customerID, id uuid.UUID) (*apikey.APIKey, error) {
	var key apikey.APIKey
	err := r.DB().WithContext(ctx).
		Where("id = ? AND customer_id = ?", id, customerID).
		First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) GetByHash(ctx context.Context, hash string) (*apikey.APIKey, error) {
	var key apikey.APIKey
	err := r.DB().WithContext(ctx).
		Where("key_hash = ?", hash).
		First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) GetByHashWithOrg(ctx context.Context, hash string) (*APIKeyWithOrganization, error) {
	var key APIKeyWithOrganization
	err := r.DB().WithContext(ctx).
		Table("api_keys").
		Select("api_keys.*, customers.organization_id").
		Joins("JOIN customers ON customers.id = api_keys.customer_id").
		Where("api_keys.key_hash = ?", hash).
		First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) ListByCustomer(
	ctx context.Context,
	customerID uuid.UUID,
	isActive *bool,
	limit, offset int,
) ([]apikey.APIKey, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := r.DB().WithContext(ctx).Model(&apikey.APIKey{}).
		Where("customer_id = ?", customerID)
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var keys []apikey.APIKey
	err := query.
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&keys).Error
	if err != nil {
		return nil, 0, err
	}

	return keys, total, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, key *apikey.APIKey) error {
	result := r.DB().WithContext(ctx).
		Model(key).
		Updates(key)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (r *apiKeyRepository) RecordUsage(ctx context.Context, keyID uuid.UUID, ip string, usedAt time.Time) error {
	updates := map[string]interface{}{
		"last_used_at": usedAt,
	}
	if ip != "" {
		updates["last_used_ip"] = ip
	} else {
		updates["last_used_ip"] = nil
	}

	result := r.DB().WithContext(ctx).
		Model(&apikey.APIKey{}).
		Where("id = ?", keyID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (r *apiKeyRepository) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int64, error) {
	var total int64
	err := r.DB().WithContext(ctx).
		Model(&apikey.APIKey{}).
		Where("customer_id = ?", customerID).
		Count(&total).Error
	return total, err
}
