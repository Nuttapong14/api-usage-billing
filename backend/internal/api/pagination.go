package api

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)

// PaginationParams represents pagination parameters from request
type PaginationParams struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	SortBy   string `query:"sort_by"`
	SortDir  string `query:"sort_dir"`
	Search   string `query:"search"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Pagination *Pagination `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	CurrentPage  int    `json:"current_page"`
	PageSize     int    `json:"page_size"`
	TotalItems   int64  `json:"total_items"`
	TotalPages   int    `json:"total_pages"`
	HasNext      bool   `json:"has_next"`
	HasPrev      bool   `json:"has_prev"`
	NextPage     *int   `json:"next_page,omitempty"`
	PrevPage     *int   `json:"prev_page,omitempty"`
	FirstPage    int    `json:"first_page"`
	LastPage     int    `json:"last_page"`
	Links        *Links `json:"links,omitempty"`
}

// Links contains HATEOAS-style pagination links
type Links struct {
	Self  string `json:"self"`
	First string `json:"first"`
	Last  string `json:"last"`
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
}

// GetPaginationParams extracts and validates pagination parameters from request
func GetPaginationParams(c *fiber.Ctx) PaginationParams {
	page, _ := strconv.Atoi(c.Query("page", strconv.Itoa(DefaultPage)))
	pageSize, _ := strconv.Atoi(c.Query("page_size", strconv.Itoa(DefaultPageSize)))
	sortBy := c.Query("sort_by", "created_at")
	sortDir := strings.ToLower(c.Query("sort_dir", "desc"))
	search := c.Query("search", "")

	// Validate and normalize
	if page < 1 {
		page = DefaultPage
	}

	if pageSize < MinPageSize {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
		SortDir:  sortDir,
		Search:   search,
	}
}

// Offset calculates the database offset for pagination
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the page size (limit for database query)
func (p PaginationParams) Limit() int {
	return p.PageSize
}

// OrderBy returns the SQL order by clause
func (p PaginationParams) OrderBy() string {
	return fmt.Sprintf("%s %s", p.SortBy, p.SortDir)
}

// NewPagination creates pagination metadata from results
func NewPagination(currentPage, pageSize int, totalItems int64) *Pagination {
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	hasNext := currentPage < totalPages
	hasPrev := currentPage > 1

	pagination := &Pagination{
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
		FirstPage:   1,
		LastPage:    totalPages,
	}

	if hasNext {
		nextPage := currentPage + 1
		pagination.NextPage = &nextPage
	}

	if hasPrev {
		prevPage := currentPage - 1
		pagination.PrevPage = &prevPage
	}

	return pagination
}

// NewPaginationWithLinks creates pagination with HATEOAS links
func NewPaginationWithLinks(c *fiber.Ctx, currentPage, pageSize int, totalItems int64) *Pagination {
	pagination := NewPagination(currentPage, pageSize, totalItems)
	pagination.Links = buildPaginationLinks(c, pagination)
	return pagination
}

// buildPaginationLinks builds HATEOAS-style pagination links
func buildPaginationLinks(c *fiber.Ctx, p *Pagination) *Links {
	baseURL := buildBaseURL(c)

	links := &Links{
		Self:  buildPageURL(baseURL, c, p.CurrentPage, p.PageSize),
		First: buildPageURL(baseURL, c, p.FirstPage, p.PageSize),
		Last:  buildPageURL(baseURL, c, p.LastPage, p.PageSize),
	}

	if p.HasNext {
		links.Next = buildPageURL(baseURL, c, *p.NextPage, p.PageSize)
	}

	if p.HasPrev {
		links.Prev = buildPageURL(baseURL, c, *p.PrevPage, p.PageSize)
	}

	return links
}

// buildBaseURL constructs the base URL from the request
func buildBaseURL(c *fiber.Ctx) string {
	protocol := "http"
	if c.Protocol() == "https" {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s%s", protocol, c.Hostname(), c.Path())
}

// buildPageURL builds a URL for a specific page
func buildPageURL(baseURL string, c *fiber.Ctx, page, pageSize int) string {
	// Get existing query parameters (excluding page and page_size)
	query := c.Request().URI().QueryArgs()
	params := make([]string, 0)

	query.VisitAll(func(key, value []byte) {
		k := string(key)
		if k != "page" && k != "page_size" {
			params = append(params, fmt.Sprintf("%s=%s", k, string(value)))
		}
	})

	// Add page and page_size
	params = append(params, fmt.Sprintf("page=%d", page))
	params = append(params, fmt.Sprintf("page_size=%d", pageSize))

	return baseURL + "?" + strings.Join(params, "&")
}

// Paginate sends a paginated response
func Paginate(c *fiber.Ctx, items interface{}, params PaginationParams, totalItems int64) error {
	pagination := NewPaginationWithLinks(c, params.Page, params.PageSize, totalItems)

	return c.JSON(PaginatedResponse{
		Items:      items,
		Pagination: pagination,
	})
}

// PaginateWithMeta sends a paginated response with additional metadata
func PaginateWithMeta(c *fiber.Ctx, items interface{}, params PaginationParams, totalItems int64, meta map[string]interface{}) error {
	pagination := NewPaginationWithLinks(c, params.Page, params.PageSize, totalItems)

	response := map[string]interface{}{
		"items":      items,
		"pagination": pagination,
		"meta":       meta,
	}

	return c.JSON(response)
}

// CursorPaginationParams represents cursor-based pagination parameters
type CursorPaginationParams struct {
	Cursor    string `query:"cursor"`
	Limit     int    `query:"limit"`
	Direction string `query:"direction"` // "next" or "prev"
}

// CursorPagination represents cursor-based pagination metadata
type CursorPagination struct {
	Limit      int     `json:"limit"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor,omitempty"`
	PrevCursor *string `json:"prev_cursor,omitempty"`
}

// CursorPaginatedResponse represents a cursor-paginated response
type CursorPaginatedResponse struct {
	Items      interface{}       `json:"items"`
	Pagination *CursorPagination `json:"pagination"`
}

// GetCursorPaginationParams extracts cursor pagination parameters
func GetCursorPaginationParams(c *fiber.Ctx) CursorPaginationParams {
	limit, _ := strconv.Atoi(c.Query("limit", strconv.Itoa(DefaultPageSize)))
	direction := c.Query("direction", "next")

	if limit < MinPageSize {
		limit = DefaultPageSize
	} else if limit > MaxPageSize {
		limit = MaxPageSize
	}

	if direction != "next" && direction != "prev" {
		direction = "next"
	}

	return CursorPaginationParams{
		Cursor:    c.Query("cursor", ""),
		Limit:     limit,
		Direction: direction,
	}
}

// NewCursorPagination creates cursor pagination metadata
func NewCursorPagination(limit int, hasMore bool, nextCursor, prevCursor *string) *CursorPagination {
	return &CursorPagination{
		Limit:      limit,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
	}
}

// CursorPaginate sends a cursor-paginated response
func CursorPaginate(c *fiber.Ctx, items interface{}, params CursorPaginationParams, hasMore bool, nextCursor, prevCursor *string) error {
	return c.JSON(CursorPaginatedResponse{
		Items:      items,
		Pagination: NewCursorPagination(params.Limit, hasMore, nextCursor, prevCursor),
	})
}
