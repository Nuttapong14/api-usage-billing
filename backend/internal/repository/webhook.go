package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/api-usage-billing/backend/internal/domain/webhook"
)

var (
	ErrWebhookNotFound        = errors.New("webhook endpoint not found")
	ErrWebhookDeliveryNotFound = errors.New("webhook delivery not found")
)

// WebhookRepository handles webhook persistence.
type WebhookRepository interface {
	CreateEndpoint(ctx context.Context, endpoint *webhook.WebhookEndpoint) error
	UpdateEndpoint(ctx context.Context, endpoint *webhook.WebhookEndpoint) error
	GetEndpoint(ctx context.Context, orgID, customerID, id uuid.UUID) (*webhook.WebhookEndpoint, error)
	GetEndpointByID(ctx context.Context, id uuid.UUID) (*webhook.WebhookEndpoint, error)
	ListEndpoints(ctx context.Context, orgID, customerID uuid.UUID, isActive *bool, limit, offset int) ([]webhook.WebhookEndpoint, int64, error)

	CreateDelivery(ctx context.Context, delivery *webhook.WebhookDelivery) error
	UpdateDelivery(ctx context.Context, delivery *webhook.WebhookDelivery) error
	GetDelivery(ctx context.Context, orgID, id uuid.UUID) (*webhook.WebhookDelivery, error)
	ListPendingDeliveries(ctx context.Context, asOf time.Time, limit int) ([]webhook.WebhookDelivery, error)
}

type webhookRepository struct {
	*BaseRepository
}

// NewWebhookRepository creates a new webhook repository.
func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *webhookRepository) CreateEndpoint(ctx context.Context, endpoint *webhook.WebhookEndpoint) error {
	return r.DB().WithContext(ctx).Create(endpoint).Error
}

func (r *webhookRepository) UpdateEndpoint(ctx context.Context, endpoint *webhook.WebhookEndpoint) error {
	result := r.DB().WithContext(ctx).Model(endpoint).Updates(endpoint)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrWebhookNotFound
	}
	return nil
}

func (r *webhookRepository) GetEndpoint(ctx context.Context, orgID, customerID, id uuid.UUID) (*webhook.WebhookEndpoint, error) {
	var endpoint webhook.WebhookEndpoint
	err := r.DB().WithContext(ctx).
		Where("id = ? AND organization_id = ? AND customer_id = ?", id, orgID, customerID).
		First(&endpoint).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWebhookNotFound
	}
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (r *webhookRepository) GetEndpointByID(ctx context.Context, id uuid.UUID) (*webhook.WebhookEndpoint, error) {
	var endpoint webhook.WebhookEndpoint
	err := r.DB().WithContext(ctx).
		Where("id = ?", id).
		First(&endpoint).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWebhookNotFound
	}
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (r *webhookRepository) ListEndpoints(
	ctx context.Context,
	orgID, customerID uuid.UUID,
	isActive *bool,
	limit, offset int,
) ([]webhook.WebhookEndpoint, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := r.DB().WithContext(ctx).
		Model(&webhook.WebhookEndpoint{}).
		Where("organization_id = ? AND customer_id = ?", orgID, customerID)

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var endpoints []webhook.WebhookEndpoint
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&endpoints).Error; err != nil {
		return nil, 0, err
	}

	return endpoints, total, nil
}

func (r *webhookRepository) CreateDelivery(ctx context.Context, delivery *webhook.WebhookDelivery) error {
	return r.DB().WithContext(ctx).Create(delivery).Error
}

func (r *webhookRepository) UpdateDelivery(ctx context.Context, delivery *webhook.WebhookDelivery) error {
	result := r.DB().WithContext(ctx).Model(delivery).Updates(delivery)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrWebhookDeliveryNotFound
	}
	return nil
}

func (r *webhookRepository) GetDelivery(ctx context.Context, orgID, id uuid.UUID) (*webhook.WebhookDelivery, error) {
	var delivery webhook.WebhookDelivery
	err := r.DB().WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&delivery).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWebhookDeliveryNotFound
	}
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *webhookRepository) ListPendingDeliveries(ctx context.Context, asOf time.Time, limit int) ([]webhook.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 50
	}

	query := r.DB().WithContext(ctx).
		Model(&webhook.WebhookDelivery{}).
		Where("status = ?", webhook.DeliveryStatusPending)

	if !asOf.IsZero() {
		query = query.Where("next_retry_at IS NULL OR next_retry_at <= ?", asOf)
	}

	var deliveries []webhook.WebhookDelivery
	if err := query.Order("created_at asc").Limit(limit).Find(&deliveries).Error; err != nil {
		return nil, err
	}
	return deliveries, nil
}
