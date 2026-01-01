package entity

import (
	"time"

	"github.com/google/uuid"
)

// Company represents a company/tenant
type Company struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	TaxID           string    `json:"tax_id"`
	Address         string    `json:"address"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	FiscalYearStart int       `json:"fiscal_year_start"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NewCompany creates a new company entity
func NewCompany(name, taxID string) *Company {
	return &Company{
		ID:              uuid.New(),
		Name:            name,
		TaxID:           taxID,
		FiscalYearStart: 1, // January
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
