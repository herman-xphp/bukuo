package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// ProductFilter defines filter options for listing products
type ProductFilter struct {
	ProductType *entity.ProductType
	CategoryID  *uuid.UUID
	Search      string // Search by name or code
	IsActive    *bool
	Page        int
	PageSize    int
}

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	// Create creates a new product
	Create(ctx context.Context, product *entity.Product) error

	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Product, error)

	// GetByCode retrieves a product by code
	GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Product, error)

	// List retrieves products with filtering and pagination
	List(ctx context.Context, companyID uuid.UUID, filter ProductFilter) ([]entity.Product, int64, error)

	// Update updates an existing product
	Update(ctx context.Context, product *entity.Product) error

	// Delete soft-deletes a product
	Delete(ctx context.Context, companyID, id uuid.UUID) error

	// ExistsByCode checks if a product with the given code exists
	ExistsByCode(ctx context.Context, companyID uuid.UUID, code string, excludeID *uuid.UUID) (bool, error)
}
