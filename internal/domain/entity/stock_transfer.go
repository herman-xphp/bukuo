package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// StockTransferStatus represents the status of a stock transfer
type StockTransferStatus string

const (
	StockTransferStatusDraft     StockTransferStatus = "DRAFT"
	StockTransferStatusCompleted StockTransferStatus = "COMPLETED"
	StockTransferStatusCancelled StockTransferStatus = "CANCELLED"
)

// StockTransfer represents inter-warehouse stock movement
type StockTransfer struct {
	ID              uuid.UUID           `json:"id"`
	CompanyID       uuid.UUID           `json:"company_id"`
	TransferNo      string              `json:"transfer_no"`
	FromWarehouseID uuid.UUID           `json:"from_warehouse_id"`
	ToWarehouseID   uuid.UUID           `json:"to_warehouse_id"`
	TransferDate    time.Time           `json:"transfer_date"`
	Status          StockTransferStatus `json:"status"`
	Notes           string              `json:"notes"`
	CreatedBy       uuid.UUID           `json:"created_by"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	Lines           []StockTransferLine `json:"lines,omitempty"`
	FromWarehouse   *Warehouse          `json:"from_warehouse,omitempty"`
	ToWarehouse     *Warehouse          `json:"to_warehouse,omitempty"`
}

// StockTransferLine represents a line item in stock transfer
type StockTransferLine struct {
	ID         uuid.UUID       `json:"id"`
	TransferID uuid.UUID       `json:"transfer_id"`
	ProductID  uuid.UUID       `json:"product_id"`
	Quantity   decimal.Decimal `json:"quantity"`
	Product    *Product        `json:"product,omitempty"`
}
