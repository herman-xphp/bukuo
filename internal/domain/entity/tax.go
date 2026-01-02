package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TaxType represents the type of tax
type TaxType string

const (
	TaxTypeVAT      TaxType = "VAT"      // PPN
	TaxTypeIncome   TaxType = "INCOME"   // PPh
	TaxTypeLuxury   TaxType = "LUXURY"   // PPnBM
	TaxTypeWithhold TaxType = "WITHHOLD" // PPh Potput
)

// TaxRate represents a tax rate configuration
type TaxRate struct {
	ID                uuid.UUID       `json:"id"`
	CompanyID         uuid.UUID       `json:"company_id"`
	Name              string          `json:"name"` // e.g. "PPN 11%"
	Code              string          `json:"code"` // e.g. "PPN11"
	Type              TaxType         `json:"type"`
	Rate              decimal.Decimal `json:"rate"`                // Percentage, e.g. 11.0
	SalesAccountID    uuid.UUID       `json:"sales_account_id"`    // Account for Tax Payable (Output VAT)
	PurchaseAccountID uuid.UUID       `json:"purchase_account_id"` // Account for Tax Receivable (Input VAT)
	Description       string          `json:"description,omitempty"`
	IsActive          bool            `json:"is_active"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// TaxReturnStatus represents the status of a tax return
type TaxReturnStatus string

const (
	TaxReturnDraft TaxReturnStatus = "DRAFT"
	TaxReturnFiled TaxReturnStatus = "FILED"
	TaxReturnPaid  TaxReturnStatus = "PAID"
)

// TaxReturn represents a tax return filing (e.g. SPT Masa)
type TaxReturn struct {
	ID            uuid.UUID       `json:"id"`
	CompanyID     uuid.UUID       `json:"company_id"`
	PeriodID      uuid.UUID       `json:"period_id"`
	TaxRateID     uuid.UUID       `json:"tax_rate_id"` // Which tax is this return for
	ReturnNo      string          `json:"return_no"`
	ReturnDate    time.Time       `json:"return_date"`
	TaxableAmount decimal.Decimal `json:"taxable_amount"` // DPP
	TaxAmount     decimal.Decimal `json:"tax_amount"`     // PPN/PPh
	Credits       decimal.Decimal `json:"credits"`        // Input VAT / Prepaid Tax
	PayableAmount decimal.Decimal `json:"payable_amount"` // Net Payable
	Status        TaxReturnStatus `json:"status"`
	Notes         string          `json:"notes,omitempty"`
	JournalID     *uuid.UUID      `json:"journal_id,omitempty"` // Journal for tax settlement
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
