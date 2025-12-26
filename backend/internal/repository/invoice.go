package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Nuttapong14/api-usage-billing/internal/domain/billing"
)

var (
	ErrInvoiceNotFound = errors.New("invoice not found")
)

// InvoiceFilters contains invoice list filters.
type InvoiceFilters struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	Status         *billing.InvoiceStatus
	StartDate      *time.Time
	EndDate        *time.Time
	Limit          int
	Offset         int
	SortBy         string
	SortDir        string
}

// InvoiceRepository handles invoice data access.
type InvoiceRepository interface {
	Create(ctx context.Context, invoice *billing.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*billing.Invoice, error)
	GetByIDForCustomer(ctx context.Context, customerID, id uuid.UUID) (*billing.Invoice, error)
	ListForCustomer(ctx context.Context, filters InvoiceFilters) ([]billing.Invoice, int64, error)
	Update(ctx context.Context, invoice *billing.Invoice) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status billing.InvoiceStatus, paidAt *time.Time) error
}

type invoiceRepository struct {
	*BaseRepository
}

// NewInvoiceRepository creates a new invoice repository.
func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *billing.Invoice) error {
	return r.DB().WithContext(ctx).Create(invoice).Error
}

func (r *invoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*billing.Invoice, error) {
	var invoice billing.Invoice
	err := r.DB().WithContext(ctx).
		Where("id = ?", id).
		First(&invoice).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) GetByIDForCustomer(ctx context.Context, customerID, id uuid.UUID) (*billing.Invoice, error) {
	var invoice billing.Invoice
	err := r.DB().WithContext(ctx).
		Where("id = ?", id).
		Where("customer_id = ?", customerID).
		First(&invoice).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) ListForCustomer(ctx context.Context, filters InvoiceFilters) ([]billing.Invoice, int64, error) {
	var invoices []billing.Invoice

	query := r.DB().WithContext(ctx).Model(&billing.Invoice{})
	if filters.CustomerID != uuid.Nil {
		query = query.Where("customer_id = ?", filters.CustomerID)
	}
	if filters.OrganizationID != uuid.Nil {
		query = query.Where("organization_id = ?", filters.OrganizationID)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.StartDate != nil {
		query = query.Where("issue_date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("issue_date <= ?", *filters.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := sanitizeInvoiceSort(filters.SortBy)
	sortDir := sanitizeSortDirection(filters.SortDir)

	err := query.Order(sortBy + " " + sortDir).
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error

	if err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (r *invoiceRepository) Update(ctx context.Context, invoice *billing.Invoice) error {
	result := r.DB().WithContext(ctx).
		Model(invoice).
		Updates(invoice)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInvoiceNotFound
	}
	return nil
}

func (r *invoiceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status billing.InvoiceStatus, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if paidAt != nil {
		updates["paid_at"] = *paidAt
	}

	result := r.DB().WithContext(ctx).
		Model(&billing.Invoice{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInvoiceNotFound
	}
	return nil
}

func sanitizeInvoiceSort(sortBy string) string {
	switch strings.ToLower(sortBy) {
	case "issue_date":
		return "issue_date"
	case "due_date":
		return "due_date"
	case "total":
		return "total"
	case "status":
		return "status"
	case "created_at":
		return "created_at"
	default:
		return "issue_date"
	}
}

func sanitizeSortDirection(dir string) string {
	if strings.ToLower(dir) == "asc" {
		return "asc"
	}
	return "desc"
}
