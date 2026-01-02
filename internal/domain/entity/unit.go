package entity

import (
	"time"

	"github.com/google/uuid"
)

// UnitOfMeasure represents a unit of measurement (PCS, KG, BOX, etc.)
type UnitOfMeasure struct {
	ID        uuid.UUID `json:"id"`
	CompanyID uuid.UUID `json:"company_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUnitOfMeasure creates a new unit of measure
func NewUnitOfMeasure(companyID uuid.UUID, code, name string) *UnitOfMeasure {
	return &UnitOfMeasure{
		ID:        uuid.New(),
		CompanyID: companyID,
		Code:      code,
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// ProductCategory represents a product category (hierarchical)
type ProductCategory struct {
	ID        uuid.UUID  `json:"id"`
	CompanyID uuid.UUID  `json:"company_id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// NewProductCategory creates a new product category
func NewProductCategory(companyID uuid.UUID, name string, parentID *uuid.UUID) *ProductCategory {
	return &ProductCategory{
		ID:        uuid.New(),
		CompanyID: companyID,
		Name:      name,
		ParentID:  parentID,
		CreatedAt: time.Now(),
	}
}
