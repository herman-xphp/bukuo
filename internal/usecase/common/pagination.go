package common

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 20

// MaxPageSize is the maximum allowed page size
const MaxPageSize = 100

// Pagination holds validated pagination parameters
type Pagination struct {
	Page     int
	PageSize int
	Offset   int
}

// ValidatePagination validates and normalizes pagination parameters.
// Returns a Pagination struct with validated values.
func ValidatePagination(page, pageSize int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Pagination{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

// CalculateTotalPages calculates total pages from total items and page size.
func CalculateTotalPages(total int64, pageSize int) int64 {
	if pageSize <= 0 {
		return 0
	}
	pages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		pages++
	}
	return pages
}

// PaginatedResult holds pagination metadata for list operations.
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int64 `json:"total_pages"`
}

// NewPaginatedResult creates a new paginated result with calculated total pages.
func NewPaginatedResult[T any](items []T, total int64, p Pagination) PaginatedResult[T] {
	return PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: CalculateTotalPages(total, p.PageSize),
	}
}
