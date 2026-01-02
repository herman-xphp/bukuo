package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// UnitRepository defines the interface for unit of measure data access
type UnitRepository interface {
	Create(ctx context.Context, unit *entity.UnitOfMeasure) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.UnitOfMeasure, error)
	GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.UnitOfMeasure, error)
	List(ctx context.Context, companyID uuid.UUID) ([]entity.UnitOfMeasure, error)
	Delete(ctx context.Context, companyID, id uuid.UUID) error
	ExistsByCode(ctx context.Context, companyID uuid.UUID, code string) (bool, error)
}

// CategoryRepository defines the interface for product category data access
type CategoryRepository interface {
	Create(ctx context.Context, category *entity.ProductCategory) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.ProductCategory, error)
	List(ctx context.Context, companyID uuid.UUID) ([]entity.ProductCategory, error)
	Update(ctx context.Context, category *entity.ProductCategory) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}
