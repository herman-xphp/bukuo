package entity

import (
	"time"

	"github.com/google/uuid"
)

// Warehouse represents a warehouse/storage location
type Warehouse struct {
	ID        uuid.UUID `json:"id"`
	CompanyID uuid.UUID `json:"company_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Address   string    `json:"address,omitempty"`
	IsDefault bool      `json:"is_default"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewWarehouse creates a new warehouse
func NewWarehouse(companyID uuid.UUID, code, name string) *Warehouse {
	return &Warehouse{
		ID:        uuid.New(),
		CompanyID: companyID,
		Code:      code,
		Name:      name,
		IsDefault: false,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
