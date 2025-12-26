package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/api-usage-billing/backend/internal/domain/subscription"
)

var (
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
)

// SubscriptionRepository handles subscription data access.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *subscription.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) (*subscription.Subscription, error)
	Update(ctx context.Context, sub *subscription.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type subscriptionRepository struct {
	*BaseRepository
}

// NewSubscriptionRepository creates a new subscription repository.
func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *subscriptionRepository) Create(ctx context.Context, sub *subscription.Subscription) error {
	return r.DB().WithContext(ctx).Create(sub).Error
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	err := r.DB().WithContext(ctx).
		Where("id = ?", id).
		First(&sub).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) (*subscription.Subscription, error) {
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

func (r *subscriptionRepository) Update(ctx context.Context, sub *subscription.Subscription) error {
	result := r.DB().WithContext(ctx).
		Model(sub).
		Updates(sub)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSubscriptionNotFound
	}
	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.DB().WithContext(ctx).
		Where("id = ?", id).
		Delete(&subscription.Subscription{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSubscriptionNotFound
	}
	return nil
}
