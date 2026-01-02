package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BankTransactionType represents the type of bank transaction
type BankTransactionType string

const (
	BankTransactionDeposit    BankTransactionType = "DEPOSIT"
	BankTransactionWithdrawal BankTransactionType = "WITHDRAWAL"
	BankTransactionTransfer   BankTransactionType = "TRANSFER"
	BankTransactionCharge     BankTransactionType = "CHARGE"
	BankTransactionInterest   BankTransactionType = "INTEREST"
)

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "CASH"
	PaymentMethodTransfer PaymentMethod = "TRANSFER"
	PaymentMethodCheck    PaymentMethod = "CHECK"
	PaymentMethodGiro     PaymentMethod = "GIRO"
)

// BankAccount represents a bank account
type BankAccount struct {
	ID             uuid.UUID       `json:"id"`
	CompanyID      uuid.UUID       `json:"company_id"`
	AccountID      uuid.UUID       `json:"account_id"` // GL Account
	BankName       string          `json:"bank_name"`
	AccountNumber  string          `json:"account_number"`
	AccountName    string          `json:"account_name"`
	CurrencyID     uuid.UUID       `json:"currency_id"`
	CurrentBalance decimal.Decimal `json:"current_balance"`
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// BankTransaction represents a bank transaction
type BankTransaction struct {
	ID              uuid.UUID           `json:"id"`
	CompanyID       uuid.UUID           `json:"company_id"`
	BankAccountID   uuid.UUID           `json:"bank_account_id"`
	TransactionNo   string              `json:"transaction_no"`
	TransactionDate time.Time           `json:"transaction_date"`
	Type            BankTransactionType `json:"type"`
	Amount          decimal.Decimal     `json:"amount"`
	Description     string              `json:"description"`
	Reference       string              `json:"reference,omitempty"`
	JournalID       *uuid.UUID          `json:"journal_id,omitempty"`
	Reconciled      bool                `json:"reconciled"`
	ReconciledAt    *time.Time          `json:"reconciled_at,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
}

// Payment represents a payment (AR Collection or AP Payment)
type Payment struct {
	ID            uuid.UUID       `json:"id"`
	CompanyID     uuid.UUID       `json:"company_id"`
	PaymentNo     string          `json:"payment_no"`
	PaymentType   string          `json:"payment_type"` // RECEIVE or PAY
	ContactID     uuid.UUID       `json:"contact_id"`
	BankAccountID uuid.UUID       `json:"bank_account_id"`
	PaymentDate   time.Time       `json:"payment_date"`
	Method        PaymentMethod   `json:"payment_method"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	Notes         string          `json:"notes,omitempty"`
	JournalID     *uuid.UUID      `json:"journal_id,omitempty"`
	Lines         []PaymentLine   `json:"lines,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// PaymentLine represents a payment allocation to an invoice
type PaymentLine struct {
	ID         uuid.UUID       `json:"id"`
	PaymentID  uuid.UUID       `json:"payment_id"`
	InvoiceID  uuid.UUID       `json:"invoice_id"`
	InvoiceNo  string          `json:"invoice_no"`
	InvoiceAmt decimal.Decimal `json:"invoice_amount"`
	PaidAmount decimal.Decimal `json:"paid_amount"`
}

// BankReconciliation represents a bank reconciliation
type BankReconciliation struct {
	ID               uuid.UUID       `json:"id"`
	CompanyID        uuid.UUID       `json:"company_id"`
	BankAccountID    uuid.UUID       `json:"bank_account_id"`
	PeriodID         uuid.UUID       `json:"period_id"`
	StatementDate    time.Time       `json:"statement_date"`
	StatementBalance decimal.Decimal `json:"statement_balance"`
	BookBalance      decimal.Decimal `json:"book_balance"`
	Difference       decimal.Decimal `json:"difference"`
	Status           string          `json:"status"` // DRAFT, COMPLETED
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

// CashTransaction represents a cash/petty cash transaction
type CashTransaction struct {
	ID              uuid.UUID           `json:"id"`
	CompanyID       uuid.UUID           `json:"company_id"`
	TransactionNo   string              `json:"transaction_no"`
	TransactionDate time.Time           `json:"transaction_date"`
	Type            BankTransactionType `json:"type"`
	Amount          decimal.Decimal     `json:"amount"`
	Description     string              `json:"description"`
	Category        string              `json:"category,omitempty"`
	JournalID       *uuid.UUID          `json:"journal_id,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
}
