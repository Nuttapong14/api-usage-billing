package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Nuttapong14/api-usage-billing/internal/domain/subscription"
)

// TierRepository handles subscription tier data access.
type TierRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*subscription.SubscriptionTier, error)
	ListForOrg(ctx context.Context, orgID uuid.UUID) ([]subscription.SubscriptionTier, error)
	ListPublicForOrg(ctx context.Context, orgID uuid.UUID) ([]subscription.SubscriptionTier, error)
}

type tierRepository struct {
	*BaseRepository
}

// NewTierRepository creates a new tier repository.
func NewTierRepository(db *gorm.DB) TierRepository {
	return &tierRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *tierRepository) GetByID(ctx context.Context, id uuid.UUID) (*subscription.SubscriptionTier, error) {
	var tier subscription.SubscriptionTier
	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted()).
		Where("id = ?", id).
		First(&tier).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTierNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *tierRepository) ListForOrg(ctx context.Context, orgID uuid.UUID) ([]subscription.SubscriptionTier, error) {
	var tiers []subscription.SubscriptionTier

	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.ForOrganization(orgID)).
		Order("display_order asc, created_at asc").
		Find(&tiers).Error

	if err != nil {
		return nil, err
	}
	return tiers, nil
}

func (r *tierRepository) ListPublicForOrg(ctx context.Context, orgID uuid.UUID) ([]subscription.SubscriptionTier, error) {
	var tiers []subscription.SubscriptionTier

	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.ForOrganization(orgID), r.Active()).
		Where("is_public = ?", true).
		Order("display_order asc, created_at asc").
		Find(&tiers).Error

	if err != nil {
		return nil, err
	}
	return tiers, nil
}
