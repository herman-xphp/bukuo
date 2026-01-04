package payroll

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/shopspring/decimal"
)

type PayrollRepository interface {
	CreateEmployee(ctx context.Context, e *entity.Employee) error
	GetEmployee(ctx context.Context, companyID, id uuid.UUID) (*entity.Employee, error)
	ListEmployees(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.Employee, int64, error)
	CreateComponent(ctx context.Context, c *entity.SalaryComponent) error
	ListComponents(ctx context.Context, companyID uuid.UUID) ([]entity.SalaryComponent, error)
	CreatePayRun(ctx context.Context, pr *entity.PayRun) error
	GetPayRun(ctx context.Context, companyID, id uuid.UUID) (*entity.PayRun, error)
	UpdatePayRunStatus(ctx context.Context, companyID, id uuid.UUID, status entity.PayRunStatus, journalID *uuid.UUID) error
	GetPaySlips(ctx context.Context, payRunID uuid.UUID) ([]entity.PaySlip, error)
}

type PayrollUsecase struct {
	repo        PayrollRepository
	journalRepo repository.JournalRepository
	txManager   repository.TransactionManager
}

func NewPayrollUsecase(repo PayrollRepository, jr repository.JournalRepository, tx repository.TransactionManager) *PayrollUsecase {
	return &PayrollUsecase{repo: repo, journalRepo: jr, txManager: tx}
}

// -- Employee --

func (uc *PayrollUsecase) CreateEmployee(ctx context.Context, companyID uuid.UUID, req entity.Employee) error {
	req.ID = uuid.New()
	req.CompanyID = companyID
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	return uc.repo.CreateEmployee(ctx, &req)
}

func (uc *PayrollUsecase) ListEmployees(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.Employee, int64, error) {
	offset := (page - 1) * pageSize
	return uc.repo.ListEmployees(ctx, companyID, pageSize, offset)
}

func (uc *PayrollUsecase) GetEmployee(ctx context.Context, companyID, id uuid.UUID) (*entity.Employee, error) {
	return uc.repo.GetEmployee(ctx, companyID, id)
}

// -- Pay Run --

type CreatePayRunInput struct {
	PeriodID    uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
	PaymentDate time.Time
}

func (uc *PayrollUsecase) GeneratePayRun(ctx context.Context, companyID uuid.UUID, input CreatePayRunInput) (*entity.PayRun, error) {
	// 1. Get all active employees
	employees, _, err := uc.repo.ListEmployees(ctx, companyID, 1000, 0) // Assume <1000 for now
	if err != nil {
		return nil, err
	}

	payRun := &entity.PayRun{
		ID:          uuid.New(),
		CompanyID:   companyID,
		PeriodID:    input.PeriodID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		PaymentDate: input.PaymentDate,
		Status:      entity.PayRunStatusDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 2. Calculate slips
	for _, empSummary := range employees {
		// Need full details for components
		emp, err := uc.repo.GetEmployee(ctx, companyID, empSummary.ID)
		if err != nil {
			return nil, err
		}

		if emp.Status == entity.EmploymentStatusResigned || emp.Status == entity.EmploymentStatusTerminated {
			continue // Skip non-active
		}

		slip := entity.PaySlip{
			ID:          uuid.New(),
			PayRunID:    payRun.ID,
			EmployeeID:  emp.ID,
			BasicSalary: emp.BasicSalary,
			GrossPay:    emp.BasicSalary,
			Deductions:  decimal.Zero,
		}

		// Basic Salary Item
		slip.Items = append(slip.Items, entity.PaySlipItem{
			ID:     uuid.New(),
			Name:   "Basic Salary",
			Type:   entity.ComponentTypeEarning,
			Amount: emp.BasicSalary,
		})

		// Add components
		for _, c := range emp.Components {
			item := entity.PaySlipItem{
				ID:          uuid.New(),
				ComponentID: c.ComponentID,
				Name:        c.Component.Name,
				Type:        c.Component.Type,
				Amount:      c.Amount,
			}
			slip.Items = append(slip.Items, item)

			if c.Component.Type == entity.ComponentTypeEarning {
				slip.GrossPay = slip.GrossPay.Add(c.Amount)
			} else {
				slip.Deductions = slip.Deductions.Add(c.Amount)
			}
		}

		slip.NetPay = slip.GrossPay.Sub(slip.Deductions)
		payRun.Slips = append(payRun.Slips, slip)
		payRun.TotalGross = payRun.TotalGross.Add(slip.GrossPay)
		payRun.TotalNet = payRun.TotalNet.Add(slip.NetPay)
	}

	if err := uc.repo.CreatePayRun(ctx, payRun); err != nil {
		return nil, err
	}

	return payRun, nil
}

func (uc *PayrollUsecase) GetPayRun(ctx context.Context, companyID, id uuid.UUID) (*entity.PayRun, error) {
	pr, err := uc.repo.GetPayRun(ctx, companyID, id)
	if err != nil {
		return nil, err
	}
	slips, err := uc.repo.GetPaySlips(ctx, pr.ID)
	if err == nil {
		pr.Slips = slips
	}
	return pr, nil
}

func (uc *PayrollUsecase) ApprovePayRun(ctx context.Context, companyID, id uuid.UUID, bankAccountID uuid.UUID) error {
	payRun, err := uc.repo.GetPayRun(ctx, companyID, id)
	if err != nil {
		return err
	}
	if payRun.Status != entity.PayRunStatusDraft {
		return common.NewValidationError("only draft pay runs can be approved")
	}

	return uc.txManager.RunAtomic(ctx, func(ctx context.Context) error {
		// Create Journal Entry
		// Dr. Salary Expense (Total Gross) - Simplified, ideally split by component accounts
		// Cr. Bank (Total Net)
		// Cr. Tax Payable / Deductions (Total Deductions)

		// For simplicity in this iteration:
		// Dr. Salary Expense (Gross)
		// Cr. Bank (Net)
		// Cr. Other Payable (Difference/Deductions) -- We need an account for deductions

		// TODO: Real implementation needs mapping from Components to Accounts.
		// For now, we assume a generic "Salary Expense" account is passed or configured.
		// Since we don't have account config here, we'll placeholder the Journal Creation
		// or use the JournalRepository if we knew the accounts.

		// Let's just update status for now as Journal mapping requires more setup
		return uc.repo.UpdatePayRunStatus(ctx, companyID, id, entity.PayRunStatusApproved, nil)
	})
}
