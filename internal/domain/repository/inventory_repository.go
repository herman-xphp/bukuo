package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// WarehouseRepository defines the interface for warehouse data access
type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *entity.Warehouse) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Warehouse, error)
	GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Warehouse, error)
	GetDefault(ctx context.Context, companyID uuid.UUID) (*entity.Warehouse, error)
	List(ctx context.Context, companyID uuid.UUID) ([]entity.Warehouse, error)
	Update(ctx context.Context, warehouse *entity.Warehouse) error
	SetDefault(ctx context.Context, companyID, warehouseID uuid.UUID) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
	ExistsByCode(ctx context.Context, companyID uuid.UUID, code string) (bool, error)
}

// InventoryFilter defines filter options for inventory queries
type InventoryFilter struct {
	ProductID   *uuid.UUID
	WarehouseID *uuid.UUID
	Type        *entity.InventoryTransactionType
	StartDate   *string
	EndDate     *string
	Page        int
	PageSize    int
}

// InventoryRepository defines the interface for inventory data access
type InventoryRepository interface {
	// Transaction operations
	CreateTransaction(ctx context.Context, tx *entity.InventoryTransaction) error
	GetTransactionByID(ctx context.Context, companyID, id uuid.UUID) (*entity.InventoryTransaction, error)
	ListTransactions(ctx context.Context, companyID uuid.UUID, filter InventoryFilter) ([]entity.InventoryTransaction, int64, error)

	// Stock operations
	GetStock(ctx context.Context, companyID, productID, warehouseID uuid.UUID) (*entity.ProductStock, error)
	GetStockByProduct(ctx context.Context, companyID, productID uuid.UUID) ([]entity.ProductStock, error)
	GetStockByWarehouse(ctx context.Context, companyID, warehouseID uuid.UUID) ([]entity.ProductStock, error)
	UpdateStock(ctx context.Context, stock *entity.ProductStock) error
	GetTotalStock(ctx context.Context, companyID, productID uuid.UUID) (*entity.ProductStock, error)
	ListStocks(ctx context.Context, companyID uuid.UUID) ([]entity.ProductStock, error)
}
