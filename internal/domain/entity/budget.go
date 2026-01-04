package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BudgetStatus represents the status of a budget
type BudgetStatus string

const (
	BudgetStatusDraft    BudgetStatus = "DRAFT"
	BudgetStatusApproved BudgetStatus = "APPROVED"
	BudgetStatusActive   BudgetStatus = "ACTIVE"
	BudgetStatusClosed   BudgetStatus = "CLOSED"
)

// Budget represents an annual or period budget
type Budget struct {
	ID          uuid.UUID       `json:"id"`
	CompanyID   uuid.UUID       `json:"company_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	PeriodID    uuid.UUID       `json:"period_id"`
	StartDate   time.Time       `json:"start_date"`
	EndDate     time.Time       `json:"end_date"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	Status      BudgetStatus    `json:"status"`
	Lines       []BudgetLine    `json:"lines,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// BudgetLine represents a budget allocation for an account
type BudgetLine struct {
	ID           uuid.UUID       `json:"id"`
	BudgetID     uuid.UUID       `json:"budget_id"`
	AccountID    uuid.UUID       `json:"account_id"`
	AccountCode  string          `json:"account_code,omitempty"`
	AccountName  string          `json:"account_name,omitempty"`
	BudgetAmount decimal.Decimal `json:"budget_amount"`
	ActualAmount decimal.Decimal `json:"actual_amount"`
	Variance     decimal.Decimal `json:"variance"`
	VariancePct  decimal.Decimal `json:"variance_pct"`
	Notes        string          `json:"notes,omitempty"`
}

// BudgetVsActual represents a comparison report
type BudgetVsActual struct {
	BudgetID      uuid.UUID       `json:"budget_id"`
	BudgetName    string          `json:"budget_name"`
	PeriodName    string          `json:"period_name"`
	TotalBudget   decimal.Decimal `json:"total_budget"`
	TotalActual   decimal.Decimal `json:"total_actual"`
	TotalVariance decimal.Decimal `json:"total_variance"`
	Lines         []BudgetLine    `json:"lines"`
}
