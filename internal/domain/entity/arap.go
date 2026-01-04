package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ARAgingBucket represents an aging bucket for receivables
type ARAgingBucket struct {
	Label        string          `json:"label"`
	Amount       decimal.Decimal `json:"amount"`
	InvoiceCount int             `json:"invoice_count"`
}

// ARAgingReport represents an aging report for receivables
type ARAgingReport struct {
	CompanyID     uuid.UUID       `json:"company_id"`
	ReportDate    time.Time       `json:"report_date"`
	TotalAR       decimal.Decimal `json:"total_ar"`
	Current       ARAgingBucket   `json:"current"`     // 0-30 days
	ThirtyDays    ARAgingBucket   `json:"thirty_days"` // 31-60 days
	SixtyDays     ARAgingBucket   `json:"sixty_days"`  // 61-90 days
	NinetyDays    ARAgingBucket   `json:"ninety_days"` // 91-120 days
	OverOneTwenty ARAgingBucket   `json:"over_120"`    // >120 days
	ByCustomer    []CustomerAging `json:"by_customer,omitempty"`
}

// CustomerAging represents aging by customer
type CustomerAging struct {
	ContactID     uuid.UUID       `json:"contact_id"`
	ContactName   string          `json:"contact_name"`
	TotalDue      decimal.Decimal `json:"total_due"`
	Current       decimal.Decimal `json:"current"`
	ThirtyDays    decimal.Decimal `json:"thirty_days"`
	SixtyDays     decimal.Decimal `json:"sixty_days"`
	NinetyDays    decimal.Decimal `json:"ninety_days"`
	OverOneTwenty decimal.Decimal `json:"over_120"`
}

// APAgingReport represents an aging report for payables
type APAgingReport struct {
	CompanyID     uuid.UUID       `json:"company_id"`
	ReportDate    time.Time       `json:"report_date"`
	TotalAP       decimal.Decimal `json:"total_ap"`
	Current       ARAgingBucket   `json:"current"`
	ThirtyDays    ARAgingBucket   `json:"thirty_days"`
	SixtyDays     ARAgingBucket   `json:"sixty_days"`
	NinetyDays    ARAgingBucket   `json:"ninety_days"`
	OverOneTwenty ARAgingBucket   `json:"over_120"`
	BySupplier    []SupplierAging `json:"by_supplier,omitempty"`
}

// SupplierAging represents aging by supplier
type SupplierAging struct {
	ContactID     uuid.UUID       `json:"contact_id"`
	ContactName   string          `json:"contact_name"`
	TotalDue      decimal.Decimal `json:"total_due"`
	Current       decimal.Decimal `json:"current"`
	ThirtyDays    decimal.Decimal `json:"thirty_days"`
	SixtyDays     decimal.Decimal `json:"sixty_days"`
	NinetyDays    decimal.Decimal `json:"ninety_days"`
	OverOneTwenty decimal.Decimal `json:"over_120"`
}

// OutstandingInvoice represents an unpaid invoice
type OutstandingInvoice struct {
	InvoiceID   uuid.UUID       `json:"invoice_id"`
	InvoiceNo   string          `json:"invoice_no"`
	InvoiceType string          `json:"invoice_type"` // SALES or PURCHASE
	ContactID   uuid.UUID       `json:"contact_id"`
	ContactName string          `json:"contact_name"`
	InvoiceDate time.Time       `json:"invoice_date"`
	DueDate     time.Time       `json:"due_date"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	PaidAmount  decimal.Decimal `json:"paid_amount"`
	BalanceDue  decimal.Decimal `json:"balance_due"`
	DaysOverdue int             `json:"days_overdue"`
}
