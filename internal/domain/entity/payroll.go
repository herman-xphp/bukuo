package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// EmploymentStatus constants
type EmploymentStatus string

const (
	EmploymentStatusActive     EmploymentStatus = "ACTIVE"
	EmploymentStatusProbation  EmploymentStatus = "PROBATION"
	EmploymentStatusResigned   EmploymentStatus = "RESIGNED"
	EmploymentStatusTerminated EmploymentStatus = "TERMINATED"
)

// ComponentType constants
type ComponentType string

const (
	ComponentTypeEarning   ComponentType = "EARNING"
	ComponentTypeDeduction ComponentType = "DEDUCTION"
)

// PayRunStatus constants
type PayRunStatus string

const (
	PayRunStatusDraft    PayRunStatus = "DRAFT"
	PayRunStatusApproved PayRunStatus = "APPROVED"
	PayRunStatusPaid     PayRunStatus = "PAID"
	PayRunStatusVoid     PayRunStatus = "VOID"
)

// Employee represents a staff member
type Employee struct {
	ID          uuid.UUID                 `json:"id"`
	CompanyID   uuid.UUID                 `json:"company_id"`
	FirstName   string                    `json:"first_name"`
	LastName    string                    `json:"last_name"`
	Email       string                    `json:"email"`
	Phone       string                    `json:"phone,omitempty"`
	JoinDate    time.Time                 `json:"join_date"`
	ResignDate  *time.Time                `json:"resign_date,omitempty"`
	Status      EmploymentStatus          `json:"status"`
	JobTitle    string                    `json:"job_title"`
	Department  string                    `json:"department,omitempty"`
	BasicSalary decimal.Decimal           `json:"basic_salary"`
	BankName    string                    `json:"bank_name,omitempty"`
	BankAccount string                    `json:"bank_account,omitempty"`
	TaxID       string                    `json:"tax_id,omitempty"` // NPWP
	Components  []EmployeeSalaryComponent `json:"components,omitempty"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// SalaryComponent represents a type of pay item (e.g. Transport Allowance, Health Insurance)
type SalaryComponent struct {
	ID        uuid.UUID       `json:"id"`
	CompanyID uuid.UUID       `json:"company_id"`
	Name      string          `json:"name"`
	Type      ComponentType   `json:"type"`
	Amount    decimal.Decimal `json:"amount"` // Default/Fixed amount
	IsTaxable bool            `json:"is_taxable"`
	AccountID uuid.UUID       `json:"account_id"` // Configured expense/liability account
	CreatedAt time.Time       `json:"created_at"`
}

// EmployeeSalaryComponent links an employee to specific components (overrides default amount)
type EmployeeSalaryComponent struct {
	ID          uuid.UUID       `json:"id"`
	EmployeeID  uuid.UUID       `json:"employee_id"`
	ComponentID uuid.UUID       `json:"component_id"`
	Component   SalaryComponent `json:"component,omitempty"`
	Amount      decimal.Decimal `json:"amount"`
}

// PayRun represents a payroll processing period
type PayRun struct {
	ID          uuid.UUID       `json:"id"`
	CompanyID   uuid.UUID       `json:"company_id"`
	PeriodID    uuid.UUID       `json:"period_id"` // Accounting period
	StartDate   time.Time       `json:"start_date"`
	EndDate     time.Time       `json:"end_date"`
	PaymentDate time.Time       `json:"payment_date"`
	TotalGross  decimal.Decimal `json:"total_gross"`
	TotalNet    decimal.Decimal `json:"total_net"`
	Status      PayRunStatus    `json:"status"`
	Slips       []PaySlip       `json:"slips,omitempty"`
	JournalID   *uuid.UUID      `json:"journal_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// PaySlip represents an individual employee's pay for a run
type PaySlip struct {
	ID           uuid.UUID       `json:"id"`
	PayRunID     uuid.UUID       `json:"pay_run_id"`
	EmployeeID   uuid.UUID       `json:"employee_id"`
	EmployeeName string          `json:"employee_name"`
	BasicSalary  decimal.Decimal `json:"basic_salary"`
	GrossPay     decimal.Decimal `json:"gross_pay"`
	Deductions   decimal.Decimal `json:"deductions"`
	NetPay       decimal.Decimal `json:"net_pay"`
	Items        []PaySlipItem   `json:"items,omitempty"`
}

// PaySlipItem represents a line item on a payslip
type PaySlipItem struct {
	ID          uuid.UUID       `json:"id"`
	PaySlipID   uuid.UUID       `json:"pay_slip_id"`
	ComponentID uuid.UUID       `json:"component_id"` // Optional, nil for Basic Salary
	Name        string          `json:"name"`
	Type        ComponentType   `json:"type"`
	Amount      decimal.Decimal `json:"amount"`
}
