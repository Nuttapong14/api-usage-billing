package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Nuttapong14/api-usage-billing/internal/domain/organization"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrSlugAlreadyExists    = errors.New("organization slug already exists")
)

// OrganizationRepository handles organization data access
type OrganizationRepository interface {
	Repository[organization.Organization]

	// GetBySlug retrieves an organization by slug
	GetBySlug(ctx context.Context, slug string) (*organization.Organization, error)

	// ExistsBySlug checks if an organization with the given slug exists
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// ListActive returns all active organizations
	ListActive(ctx context.Context, params PaginationParams) (*PaginatedResult[organization.Organization], error)
}

type organizationRepository struct {
	*BaseRepository
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *organizationRepository) Create(ctx context.Context, org *organization.Organization) error {
	// Check for existing slug
	exists, err := r.ExistsBySlug(ctx, org.Slug)
	if err != nil {
		return err
	}
	if exists {
		return ErrSlugAlreadyExists
	}

	return r.DB().WithContext(ctx).Create(org).Error
}

func (r *organizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	var org organization.Organization
	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted()).
		Where("id = ?", id).
		First(&org).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrganizationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	var org organization.Organization
	err := r.DB().WithContext(ctx).
		Scopes(r.NotDeleted()).
		Where("slug = ?", slug).
		First(&org).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrganizationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).
		Model(&organization.Organization{}).
		Scopes(r.NotDeleted()).
		Where("slug = ?", slug).
		Count(&count).Error

	return count > 0, err
}

func (r *organizationRepository) Update(ctx context.Context, org *organization.Organization) error {
	result := r.DB().WithContext(ctx).
		Model(org).
		Updates(org)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}

func (r *organizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.DB().WithContext(ctx).
		Where("id = ?", id).
		Delete(&organization.Organization{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}

func (r *organizationRepository) List(ctx context.Context, params PaginationParams) (*PaginatedResult[organization.Organization], error) {
	var orgs []organization.Organization

	// Get total count
	total, err := r.CountTotal(ctx, &organization.Organization{}, r.NotDeleted())
	if err != nil {
		return nil, err
	}

	// Get paginated results
	err = r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.Paginate(params)).
		Find(&orgs).Error

	if err != nil {
		return nil, err
	}

	return &PaginatedResult[organization.Organization]{
		Items:      orgs,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

func (r *organizationRepository) ListActive(ctx context.Context, params PaginationParams) (*PaginatedResult[organization.Organization], error) {
	var orgs []organization.Organization

	// Get total count
	total, err := r.CountTotal(ctx, &organization.Organization{}, r.NotDeleted(), r.Active())
	if err != nil {
		return nil, err
	}

	// Get paginated results
	err = r.DB().WithContext(ctx).
		Scopes(r.NotDeleted(), r.Active(), r.Paginate(params)).
		Find(&orgs).Error

	if err != nil {
		return nil, err
	}

	return &PaginatedResult[organization.Organization]{
		Items:      orgs,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}
