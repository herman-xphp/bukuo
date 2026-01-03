package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Sales document statuses
type SalesStatus string

const (
	SalesStatusDraft     SalesStatus = "DRAFT"
	SalesStatusSent      SalesStatus = "SENT"
	SalesStatusAccepted  SalesStatus = "ACCEPTED"
	SalesStatusRejected  SalesStatus = "REJECTED"
	SalesStatusCancelled SalesStatus = "CANCELLED"
	SalesStatusCompleted SalesStatus = "COMPLETED"
)

// SalesQuotation represents a sales quotation
type SalesQuotation struct {
	ID             uuid.UUID            `json:"id"`
	CompanyID      uuid.UUID            `json:"company_id"`
	QuotationNo    string               `json:"quotation_no"`
	CustomerID     uuid.UUID            `json:"customer_id"`
	QuotationDate  time.Time            `json:"quotation_date"`
	ValidUntil     time.Time            `json:"valid_until"`
	Status         SalesStatus          `json:"status"`
	Subtotal       decimal.Decimal      `json:"subtotal"`
	TaxAmount      decimal.Decimal      `json:"tax_amount"`
	DiscountAmount decimal.Decimal      `json:"discount_amount"`
	Total          decimal.Decimal      `json:"total"`
	Notes          string               `json:"notes,omitempty"`
	Lines          []SalesQuotationLine `json:"lines,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type SalesQuotationLine struct {
	ID          uuid.UUID       `json:"id"`
	QuotationID uuid.UUID       `json:"quotation_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	Description string          `json:"description"`
	Quantity    decimal.Decimal `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	DiscountPct decimal.Decimal `json:"discount_pct"`
	TaxPct      decimal.Decimal `json:"tax_pct"`
	LineTotal   decimal.Decimal `json:"line_total"`
}

// SalesOrder represents a sales order
type SalesOrder struct {
	ID             uuid.UUID        `json:"id"`
	CompanyID      uuid.UUID        `json:"company_id"`
	OrderNo        string           `json:"order_no"`
	CustomerID     uuid.UUID        `json:"customer_id"`
	QuotationID    *uuid.UUID       `json:"quotation_id,omitempty"`
	OrderDate      time.Time        `json:"order_date"`
	DeliveryDate   *time.Time       `json:"delivery_date,omitempty"`
	Status         SalesStatus      `json:"status"`
	Subtotal       decimal.Decimal  `json:"subtotal"`
	TaxAmount      decimal.Decimal  `json:"tax_amount"`
	DiscountAmount decimal.Decimal  `json:"discount_amount"`
	Total          decimal.Decimal  `json:"total"`
	Notes          string           `json:"notes,omitempty"`
	Lines          []SalesOrderLine `json:"lines,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type SalesOrderLine struct {
	ID           uuid.UUID       `json:"id"`
	OrderID      uuid.UUID       `json:"order_id"`
	ProductID    uuid.UUID       `json:"product_id"`
	Description  string          `json:"description"`
	Quantity     decimal.Decimal `json:"quantity"`
	DeliveredQty decimal.Decimal `json:"delivered_qty"`
	UnitPrice    decimal.Decimal `json:"unit_price"`
	DiscountPct  decimal.Decimal `json:"discount_pct"`
	TaxPct       decimal.Decimal `json:"tax_pct"`
	LineTotal    decimal.Decimal `json:"line_total"`
}

// SalesInvoice represents a sales invoice
type SalesInvoice struct {
	ID             uuid.UUID          `json:"id"`
	CompanyID      uuid.UUID          `json:"company_id"`
	InvoiceNo      string             `json:"invoice_no"`
	CustomerID     uuid.UUID          `json:"customer_id"`
	OrderID        *uuid.UUID         `json:"order_id,omitempty"`
	InvoiceDate    time.Time          `json:"invoice_date"`
	DueDate        time.Time          `json:"due_date"`
	Status         SalesStatus        `json:"status"`
	Subtotal       decimal.Decimal    `json:"subtotal"`
	TaxAmount      decimal.Decimal    `json:"tax_amount"`
	DiscountAmount decimal.Decimal    `json:"discount_amount"`
	Total          decimal.Decimal    `json:"total"`
	PaidAmount     decimal.Decimal    `json:"paid_amount"`
	Notes          string             `json:"notes,omitempty"`
	JournalID      *uuid.UUID         `json:"journal_id,omitempty"` // Auto-generated journal
	Lines          []SalesInvoiceLine `json:"lines,omitempty"`
	Customer       *Customer          `json:"customer,omitempty"` // Populated via Join
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

type SalesInvoiceLine struct {
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

// DeliveryOrder represents a delivery order
type DeliveryOrder struct {
	ID           uuid.UUID           `json:"id"`
	CompanyID    uuid.UUID           `json:"company_id"`
	DeliveryNo   string              `json:"delivery_no"`
	CustomerID   uuid.UUID           `json:"customer_id"`
	OrderID      uuid.UUID           `json:"order_id"`
	WarehouseID  uuid.UUID           `json:"warehouse_id"`
	DeliveryDate time.Time           `json:"delivery_date"`
	Status       SalesStatus         `json:"status"`
	Notes        string              `json:"notes,omitempty"`
	Lines        []DeliveryOrderLine `json:"lines,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}

type DeliveryOrderLine struct {
	ID          uuid.UUID       `json:"id"`
	DeliveryID  uuid.UUID       `json:"delivery_id"`
	OrderLineID uuid.UUID       `json:"order_line_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	Quantity    decimal.Decimal `json:"quantity"`
}

// SalesReturn represents a sales return
type SalesReturn struct {
	ID           uuid.UUID         `json:"id"`
	CompanyID    uuid.UUID         `json:"company_id"`
	ReturnNo     string            `json:"return_no"`
	CustomerID   uuid.UUID         `json:"customer_id"`
	InvoiceID    uuid.UUID         `json:"invoice_id"`
	WarehouseID  uuid.UUID         `json:"warehouse_id"`
	ReturnDate   time.Time         `json:"return_date"`
	Status       SalesStatus       `json:"status"`
	Total        decimal.Decimal   `json:"total"`
	Notes        string            `json:"notes,omitempty"`
	CreditNoteID *uuid.UUID        `json:"credit_note_id,omitempty"`
	Lines        []SalesReturnLine `json:"lines,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

type SalesReturnLine struct {
	ID            uuid.UUID       `json:"id"`
	ReturnID      uuid.UUID       `json:"return_id"`
	InvoiceLineID uuid.UUID       `json:"invoice_line_id"`
	ProductID     uuid.UUID       `json:"product_id"`
	Quantity      decimal.Decimal `json:"quantity"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	LineTotal     decimal.Decimal `json:"line_total"`
}
