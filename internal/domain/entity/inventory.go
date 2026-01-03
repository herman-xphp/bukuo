package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// InventoryTransactionType represents the type of inventory transaction
type InventoryTransactionType string

const (
	InventoryTransactionStockIn  InventoryTransactionType = "STOCK_IN"
	InventoryTransactionStockOut InventoryTransactionType = "STOCK_OUT"
	InventoryTransactionTransfer InventoryTransactionType = "TRANSFER"
	InventoryTransactionAdjust   InventoryTransactionType = "ADJUSTMENT"
)

// InventoryTransaction represents a stock movement
type InventoryTransaction struct {
	ID              uuid.UUID                `json:"id"`
	CompanyID       uuid.UUID                `json:"company_id"`
	TransactionNo   string                   `json:"transaction_no"`
	Type            InventoryTransactionType `json:"type"`
	ProductID       uuid.UUID                `json:"product_id"`
	WarehouseID     uuid.UUID                `json:"warehouse_id"`
	ToWarehouseID   *uuid.UUID               `json:"to_warehouse_id,omitempty"` // For transfers
	Quantity        decimal.Decimal          `json:"quantity"`
	UnitCost        decimal.Decimal          `json:"unit_cost"`
	TotalCost       decimal.Decimal          `json:"total_cost"`
	Reference       string                   `json:"reference,omitempty"` // PO, SO, etc.
	Notes           string                   `json:"notes,omitempty"`
	TransactionDate time.Time                `json:"transaction_date"`
	CreatedAt       time.Time                `json:"created_at"`
}

// NewInventoryTransaction creates a new inventory transaction
func NewInventoryTransaction(companyID uuid.UUID, txNo string, txType InventoryTransactionType, productID, warehouseID uuid.UUID, qty, unitCost decimal.Decimal) *InventoryTransaction {
	return &InventoryTransaction{
		ID:              uuid.New(),
		CompanyID:       companyID,
		TransactionNo:   txNo,
		Type:            txType,
		ProductID:       productID,
		WarehouseID:     warehouseID,
		Quantity:        qty,
		UnitCost:        unitCost,
		TotalCost:       qty.Mul(unitCost),
		TransactionDate: time.Now(),
		CreatedAt:       time.Now(),
	}
}

// ProductStock represents current stock of a product in a warehouse
type ProductStock struct {
	ID          uuid.UUID       `json:"id"`
	CompanyID   uuid.UUID       `json:"company_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	WarehouseID uuid.UUID       `json:"warehouse_id"`
	Quantity    decimal.Decimal `json:"quantity"`
	AverageCost decimal.Decimal `json:"average_cost"`
	UpdatedAt   time.Time       `json:"updated_at"`

	// View Fields
	ProductName   string `json:"product_name,omitempty"`
	ProductCode   string `json:"code,omitempty"` // Frontend uses 'code' now
	WarehouseName string `json:"warehouse_name,omitempty"`
}
