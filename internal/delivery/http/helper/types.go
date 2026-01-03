package helper

import "github.com/shopspring/decimal"

// Pagination holds parsed pagination parameters from request
type Pagination struct {
	Page     int
	PageSize int
	Limit    int
	Offset   int
}

// DefaultPagination returns default pagination values
func DefaultPagination() Pagination {
	return Pagination{
		Page:     1,
		PageSize: 20,
		Limit:    20,
		Offset:   0,
	}
}

// Calculate computes limit and offset from page and pageSize
func (p *Pagination) Calculate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	p.Limit = p.PageSize
	p.Offset = (p.Page - 1) * p.PageSize
}

// TotalPages calculates total pages from total items
func (p Pagination) TotalPages(total int64) int64 {
	if p.PageSize <= 0 {
		return 0
	}
	pages := total / int64(p.PageSize)
	if total%int64(p.PageSize) > 0 {
		pages++
	}
	return pages
}

// PaginatedResult holds pagination metadata for response
type PaginatedResult struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int64       `json:"total_pages"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult(items interface{}, total int64, p Pagination) PaginatedResult {
	return PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: p.TotalPages(total),
	}
}

// MoneyAmount represents a monetary value as string for JSON
// This avoids floating point precision issues
type MoneyAmount = decimal.Decimal
