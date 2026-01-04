package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// StockOpnameStatus represents the status of a stock opname
type StockOpnameStatus string

const (
	StockOpnameStatusDraft     StockOpnameStatus = "DRAFT"
	StockOpnameStatusApproved  StockOpnameStatus = "APPROVED"
	StockOpnameStatusCancelled StockOpnameStatus = "CANCELLED"
)

// StockOpname represents a physical inventory count
type StockOpname struct {
	ID          uuid.UUID         `json:"id"`
	CompanyID   uuid.UUID         `json:"company_id"`
	OpnameNo    string            `json:"opname_no"`
	WarehouseID uuid.UUID         `json:"warehouse_id"`
	OpnameDate  time.Time         `json:"opname_date"`
	Status      StockOpnameStatus `json:"status"`
	Notes       string            `json:"notes"`
	CreatedBy   uuid.UUID         `json:"created_by"`
	ApprovedBy  *uuid.UUID        `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time        `json:"approved_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Lines       []StockOpnameLine `json:"lines,omitempty"`
}

// StockOpnameLine represents a line item in stock opname
type StockOpnameLine struct {
	ID            uuid.UUID       `json:"id"`
	OpnameID      uuid.UUID       `json:"opname_id"`
	ProductID     uuid.UUID       `json:"product_id"`
	SystemQty     decimal.Decimal `json:"system_qty"`
	ActualQty     decimal.Decimal `json:"actual_qty"`
	DifferenceQty decimal.Decimal `json:"difference_qty"`
	Notes         string          `json:"notes"`
	Product       *Product        `json:"product,omitempty"`
}

// CalculateDifference calculates the difference between actual and system qty
func (l *StockOpnameLine) CalculateDifference() {
	l.DifferenceQty = l.ActualQty.Sub(l.SystemQty)
}
