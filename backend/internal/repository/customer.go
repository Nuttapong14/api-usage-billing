package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/api-usage-billing/backend/internal/domain/customer"
)

var (
	ErrCustomerNotFound       = errors.New("customer not found")
	ErrCustomerEmailExists    = errors.New("customer with this email already exists")
	ErrCustomerKeycloakExists = errors.New("customer with this keycloak user already exists")
)

// CustomerRepository handles customer data access
type CustomerRepository interface {
	TenantRepository[customer.Customer]

	// GetByEmail retrieves a customer by email within an organization
	GetByEmail(ctx context.Context, orgID uuid.UUID, email string) (*customer.Customer, error)

	// GetByKeycloakUserID retrieves a customer by Keycloak user ID
	GetByKeycloakUserID(ctx context.Context, keycloakUserID uuid.UUID) (*customer.Customer, error)

	// ListByStatus returns customers filtered by status
	ListByStatus(ctx context.Context, orgID uuid.UUID, status customer.Status, params PaginationParams) (*PaginatedResult[customer.Customer], error)

	// UpdateStatus updates a customer's status
	UpdateStatus(ctx context.Context, id uuid.UUID, status customer.Status, reason *string) error

	// CountByOrg returns the total number of customers for an organization
	CountByOrg(ctx context.Context, orgID uuid.UUID) (int64, error)
}

type customerRepository struct {
	*BaseRepository
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *customerRepository) Create(ctx context.Context, c *customer.Customer) error {
	return r.DB().WithContext(ctx).Create(c).Error
}

func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*customer.Customer, error) {
	var c customer.Customer
	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted()).
		Where("id = ?", id).
		First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCustomerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *customerRepository) GetByIDForOrg(ctx context.Context, orgID, id uuid.UUID) (*customer.Customer, error) {
	var c customer.Customer
	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.ForOrganization(orgID)).
		Where("id = ?", id).
		First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCustomerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *customerRepository) GetByEmail(ctx context.Context, orgID uuid.UUID, email string) (*customer.Customer, error) {
	var c customer.Customer
	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.ForOrganization(orgID)).
		Where("email = ?", email).
		First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCustomerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *customerRepository) GetByKeycloakUserID(ctx context.Context, keycloakUserID uuid.UUID) (*customer.Customer, error) {
	var c customer.Customer
	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted()).
		Where("keycloak_user_id = ?", keycloakUserID).
		First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCustomerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *customerRepository) Update(ctx context.Context, c *customer.Customer) error {
	result := r.DB().WithContext(ctx).
		Model(c).
		Updates(c)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

func (r *customerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status customer.Status, reason *string) error {
	updates := map[string]interface{}{
		"status":    status,
		"is_active": status == customer.StatusActive,
	}
	if reason != nil {
		updates["suspended_reason"] = *reason
	} else {
		updates["suspended_reason"] = nil
	}

	result := r.DB().WithContext(ctx).
		Model(&customer.Customer{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

func (r *customerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.DB().WithContext(ctx).
		Where("id = ?", id).
		Delete(&customer.Customer{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

func (r *customerRepository) List(ctx context.Context, params PaginationParams) (*PaginatedResult[customer.Customer], error) {
	var customers []customer.Customer

	total, err := r.CountTotal(ctx, &customer.Customer{}, r.NotDeleted())
	if err != nil {
		return nil, err
	}

	err = r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.Paginate(params)).
		Find(&customers).Error

	if err != nil {
		return nil, err
	}

	return &PaginatedResult[customer.Customer]{
		Items:      customers,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

func (r *customerRepository) ListForOrg(ctx context.Context, orgID uuid.UUID, params PaginationParams) (*PaginatedResult[customer.Customer], error) {
	var customers []customer.Customer

	total, err := r.CountTotal(ctx, &customer.Customer{}, r.NotDeleted(), r.ForOrganization(orgID))
	if err != nil {
		return nil, err
	}

	err = r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.ForOrganization(orgID), r.Paginate(params)).
		Find(&customers).Error

	if err != nil {
		return nil, err
	}

	return &PaginatedResult[customer.Customer]{
		Items:      customers,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

func (r *customerRepository) ListByStatus(ctx context.Context, orgID uuid.UUID, status customer.Status, params PaginationParams) (*PaginatedResult[customer.Customer], error) {
	var customers []customer.Customer

	statusScope := func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}

	total, err := r.CountTotal(ctx, &customer.Customer{}, r.NotDeleted(), r.ForOrganization(orgID), statusScope)
	if err != nil {
		return nil, err
	}

	err = r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.ForOrganization(orgID), statusScope, r.Paginate(params)).
		Find(&customers).Error

	if err != nil {
		return nil, err
	}

	return &PaginatedResult[customer.Customer]{
		Items:      customers,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

func (r *customerRepository) CountByOrg(ctx context.Context, orgID uuid.UUID) (int64, error) {
	return r.CountTotal(ctx, &customer.Customer{}, r.NotDeleted(), r.ForOrganization(orgID))
}
