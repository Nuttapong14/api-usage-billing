package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaginationParams contains pagination parameters
type PaginationParams struct {
	Page     int    `query:"page" default:"1"`
	PageSize int    `query:"page_size" default:"20"`
	SortBy   string `query:"sort_by" default:"created_at"`
	SortDir  string `query:"sort_dir" default:"desc"`
}

// PaginatedResult wraps paginated query results
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// Repository is the base interface for all repositories
type Repository[T any] interface {
	// Create creates a new entity
	Create(ctx context.Context, entity *T) error

	// GetByID retrieves an entity by ID
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)

	// Update updates an existing entity
	Update(ctx context.Context, entity *T) error

	// Delete soft-deletes an entity
	Delete(ctx context.Context, id uuid.UUID) error

	// List returns a paginated list of entities
	List(ctx context.Context, params PaginationParams) (*PaginatedResult[T], error)
}

// TenantRepository extends Repository with multi-tenant support
type TenantRepository[T any] interface {
	Repository[T]

	// GetByIDForOrg retrieves an entity by ID within an organization
	GetByIDForOrg(ctx context.Context, orgID, id uuid.UUID) (*T, error)

	// ListForOrg returns a paginated list of entities for an organization
	ListForOrg(ctx context.Context, orgID uuid.UUID, params PaginationParams) (*PaginatedResult[T], error)
}

// BaseRepository provides common repository functionality
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// DB returns the underlying database connection
func (r *BaseRepository) DB() *gorm.DB {
	return r.db
}

// WithTx returns a new repository instance with the given transaction
func (r *BaseRepository) WithTx(tx *gorm.DB) *BaseRepository {
	return &BaseRepository{db: tx}
}

// Paginate applies pagination to a query
func (r *BaseRepository) Paginate(params PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page := params.Page
		if page < 1 {
			page = 1
		}

		pageSize := params.PageSize
		if pageSize < 1 {
			pageSize = 20
		}
		if pageSize > 100 {
			pageSize = 100
		}

		offset := (page - 1) * pageSize

		// Apply sorting
		sortBy := params.SortBy
		if sortBy == "" {
			sortBy = "created_at"
		}
		sortDir := params.SortDir
		if sortDir != "asc" && sortDir != "desc" {
			sortDir = "desc"
		}

		return db.Order(sortBy + " " + sortDir).Offset(offset).Limit(pageSize)
	}
}

// ForOrganization scopes query to an organization
func (r *BaseRepository) ForOrganization(orgID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("organization_id = ?", orgID)
	}
}

// NotDeleted excludes soft-deleted records
func (r *BaseRepository) NotDeleted() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("deleted_at IS NULL")
	}
}

// Active filters for active records only
func (r *BaseRepository) Active() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ?", true)
	}
}

// CountTotal counts total records matching the query
func (r *BaseRepository) CountTotal(ctx context.Context, model interface{}, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(model)
	for _, scope := range scopes {
		query = scope(query)
	}
	err := query.Count(&count).Error
	return count, err
}

// CalculateTotalPages calculates total pages from total count and page size
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return pages
}
