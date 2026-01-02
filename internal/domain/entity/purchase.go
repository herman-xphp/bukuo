package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Purchase document statuses
type PurchaseStatus string

const (
	PurchaseStatusDraft     PurchaseStatus = "DRAFT"
	PurchaseStatusSent      PurchaseStatus = "SENT"
	PurchaseStatusApproved  PurchaseStatus = "APPROVED"
	PurchaseStatusRejected  PurchaseStatus = "REJECTED"
	PurchaseStatusReceived  PurchaseStatus = "RECEIVED"
	PurchaseStatusCancelled PurchaseStatus = "CANCELLED"
	PurchaseStatusCompleted PurchaseStatus = "COMPLETED"
)

// PurchaseRequest represents a purchase request
type PurchaseRequest struct {
	ID          uuid.UUID             `json:"id"`
	CompanyID   uuid.UUID             `json:"company_id"`
	RequestNo   string                `json:"request_no"`
	RequestDate time.Time             `json:"request_date"`
	RequestedBy uuid.UUID             `json:"requested_by"`
	Status      PurchaseStatus        `json:"status"`
	Notes       string                `json:"notes,omitempty"`
	Lines       []PurchaseRequestLine `json:"lines,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
}

type PurchaseRequestLine struct {
	ID          uuid.UUID       `json:"id"`
	RequestID   uuid.UUID       `json:"request_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	Description string          `json:"description"`
	Quantity    decimal.Decimal `json:"quantity"`
}

// PurchaseOrder represents a purchase order
type PurchaseOrder struct {
	ID             uuid.UUID           `json:"id"`
	CompanyID      uuid.UUID           `json:"company_id"`
	OrderNo        string              `json:"order_no"`
	SupplierID     uuid.UUID           `json:"supplier_id"`
	RequestID      *uuid.UUID          `json:"request_id,omitempty"`
	OrderDate      time.Time           `json:"order_date"`
	DeliveryDate   *time.Time          `json:"delivery_date,omitempty"`
	Status         PurchaseStatus      `json:"status"`
	Subtotal       decimal.Decimal     `json:"subtotal"`
	TaxAmount      decimal.Decimal     `json:"tax_amount"`
	DiscountAmount decimal.Decimal     `json:"discount_amount"`
	Total          decimal.Decimal     `json:"total"`
	Notes          string              `json:"notes,omitempty"`
	Lines          []PurchaseOrderLine `json:"lines,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type PurchaseOrderLine struct {
	ID          uuid.UUID       `json:"id"`
	OrderID     uuid.UUID       `json:"order_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	Description string          `json:"description"`
	Quantity    decimal.Decimal `json:"quantity"`
	ReceivedQty decimal.Decimal `json:"received_qty"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	DiscountPct decimal.Decimal `json:"discount_pct"`
	TaxPct      decimal.Decimal `json:"tax_pct"`
	LineTotal   decimal.Decimal `json:"line_total"`
}

// GoodsReceivedNote represents a goods received note
type GoodsReceivedNote struct {
	ID          uuid.UUID      `json:"id"`
	CompanyID   uuid.UUID      `json:"company_id"`
	GRNNo       string         `json:"grn_no"`
	SupplierID  uuid.UUID      `json:"supplier_id"`
	OrderID     uuid.UUID      `json:"order_id"`
	WarehouseID uuid.UUID      `json:"warehouse_id"`
	ReceiveDate time.Time      `json:"receive_date"`
	Status      PurchaseStatus `json:"status"`
	Notes       string         `json:"notes,omitempty"`
	Lines       []GRNLine      `json:"lines,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type GRNLine struct {
	ID          uuid.UUID       `json:"id"`
	GRNID       uuid.UUID       `json:"grn_id"`
	OrderLineID uuid.UUID       `json:"order_line_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	Quantity    decimal.Decimal `json:"quantity"`
	UnitCost    decimal.Decimal `json:"unit_cost"`
}

// PurchaseInvoice represents a purchase invoice
type PurchaseInvoice struct {
	ID             uuid.UUID             `json:"id"`
	CompanyID      uuid.UUID             `json:"company_id"`
	InvoiceNo      string                `json:"invoice_no"`
	SupplierID     uuid.UUID             `json:"supplier_id"`
	OrderID        *uuid.UUID            `json:"order_id,omitempty"`
	InvoiceDate    time.Time             `json:"invoice_date"`
	DueDate        time.Time             `json:"due_date"`
	Status         PurchaseStatus        `json:"status"`
	Subtotal       decimal.Decimal       `json:"subtotal"`
	TaxAmount      decimal.Decimal       `json:"tax_amount"`
	DiscountAmount decimal.Decimal       `json:"discount_amount"`
	Total          decimal.Decimal       `json:"total"`
	PaidAmount     decimal.Decimal       `json:"paid_amount"`
	Notes          string                `json:"notes,omitempty"`
	JournalID      *uuid.UUID            `json:"journal_id,omitempty"`
	Lines          []PurchaseInvoiceLine `json:"lines,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

type PurchaseInvoiceLine struct {
	ID          uuid.UUID       `json:"id"`
	InvoiceID   uuid.UUID       `json:"invoice_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	Description string          `json:"description"`
	Quantity    decimal.Decimal `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	DiscountPct decimal.Decimal `json:"discount_pct"`
	TaxPct      decimal.Decimal `json:"tax_pct"`
	LineTotal   decimal.Decimal `json:"line_total"`
}

// PurchaseReturn represents a purchase return
type PurchaseReturn struct {
	ID          uuid.UUID            `json:"id"`
	CompanyID   uuid.UUID            `json:"company_id"`
	ReturnNo    string               `json:"return_no"`
	SupplierID  uuid.UUID            `json:"supplier_id"`
	InvoiceID   uuid.UUID            `json:"invoice_id"`
	WarehouseID uuid.UUID            `json:"warehouse_id"`
	ReturnDate  time.Time            `json:"return_date"`
	Status      PurchaseStatus       `json:"status"`
	Total       decimal.Decimal      `json:"total"`
	Notes       string               `json:"notes,omitempty"`
	Lines       []PurchaseReturnLine `json:"lines,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}

type PurchaseReturnLine struct {
	ID            uuid.UUID       `json:"id"`
	ReturnID      uuid.UUID       `json:"return_id"`
	InvoiceLineID uuid.UUID       `json:"invoice_line_id"`
	ProductID     uuid.UUID       `json:"product_id"`
	Quantity      decimal.Decimal `json:"quantity"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	LineTotal     decimal.Decimal `json:"line_total"`
}
