package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/api-usage-billing/backend/internal/domain/billing"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

// PaymentRepository handles payment data access.
type PaymentRepository interface {
	Create(ctx context.Context, payment *billing.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*billing.Payment, error)
	ListByInvoice(ctx context.Context, invoiceID uuid.UUID, limit, offset int) ([]billing.Payment, int64, error)
	Update(ctx context.Context, payment *billing.Payment) error
}

type paymentRepository struct {
	*BaseRepository
}

// NewPaymentRepository creates a new payment repository.
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *paymentRepository) Create(ctx context.Context, payment *billing.Payment) error {
	return r.DB().WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*billing.Payment, error) {
	var payment billing.Payment
	err := r.DB().WithContext(ctx).
		Where("id = ?", id).
		First(&payment).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) ListByInvoice(ctx context.Context, invoiceID uuid.UUID, limit, offset int) ([]billing.Payment, int64, error) {
	var payments []billing.Payment

	query := r.DB().WithContext(ctx).Model(&billing.Payment{}).Where("invoice_id = ?", invoiceID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&payments).Error
	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

func (r *paymentRepository) Update(ctx context.Context, payment *billing.Payment) error {
	result := r.DB().WithContext(ctx).
		Model(payment).
		Updates(payment)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPaymentNotFound
	}
	return nil
}
