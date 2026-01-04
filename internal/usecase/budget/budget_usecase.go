package budget

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/shopspring/decimal"
)

// BudgetRepository interface
type BudgetRepository interface {
	Create(ctx context.Context, budget *entity.Budget, lines []entity.BudgetLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Budget, error)
	List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.Budget, int64, error)
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.BudgetStatus) error
	GetBudgetVsActual(ctx context.Context, companyID, budgetID uuid.UUID) (*entity.BudgetVsActual, error)
}

// BudgetUsecase handles budget business logic
type BudgetUsecase struct {
	repo BudgetRepository
}

// NewBudgetUsecase creates a new BudgetUsecase
func NewBudgetUsecase(r BudgetRepository) *BudgetUsecase {
	return &BudgetUsecase{repo: r}
}

// CreateBudgetInput holds input for budget creation
type CreateBudgetInput struct {
	Name        string
	Description string
	PeriodID    uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
	Lines       []BudgetLineInput
}

// BudgetLineInput holds input for budget line
type BudgetLineInput struct {
	AccountID    uuid.UUID
	BudgetAmount decimal.Decimal
	Notes        string
}

// CreateBudget creates a new budget
func (uc *BudgetUsecase) CreateBudget(ctx context.Context, companyID uuid.UUID, input CreateBudgetInput) (*entity.Budget, error) {
	if len(input.Lines) == 0 {
		return nil, common.NewValidationError("budget must have at least one line")
	}

	budget := &entity.Budget{
		ID:          uuid.New(),
		CompanyID:   companyID,
		Name:        input.Name,
		Description: input.Description,
		PeriodID:    input.PeriodID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		TotalAmount: decimal.Zero,
		Status:      entity.BudgetStatusDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	lines := make([]entity.BudgetLine, len(input.Lines))
	for i, l := range input.Lines {
		lines[i] = entity.BudgetLine{
			ID:           uuid.New(),
			BudgetID:     budget.ID,
			AccountID:    l.AccountID,
			BudgetAmount: l.BudgetAmount,
			Notes:        l.Notes,
		}
		budget.TotalAmount = budget.TotalAmount.Add(l.BudgetAmount)
	}

	if err := uc.repo.Create(ctx, budget, lines); err != nil {
		return nil, common.WrapErr("create budget", err)
	}

	budget.Lines = lines
	return budget, nil
}

// ListBudgets lists budgets with pagination
func (uc *BudgetUsecase) ListBudgets(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.Budget, int64, error) {
	offset := (page - 1) * pageSize
	return uc.repo.List(ctx, companyID, pageSize, offset)
}

// GetBudget retrieves a budget by ID
func (uc *BudgetUsecase) GetBudget(ctx context.Context, companyID, id uuid.UUID) (*entity.Budget, error) {
	return uc.repo.GetByID(ctx, companyID, id)
}

// ApproveBudget approves a budget
func (uc *BudgetUsecase) ApproveBudget(ctx context.Context, companyID, id uuid.UUID) error {
	budget, err := uc.repo.GetByID(ctx, companyID, id)
	if err != nil {
		return common.WrapErr("get budget", err)
	}
	if budget.Status != entity.BudgetStatusDraft {
		return common.NewValidationError("can only approve draft budgets")
	}
	return uc.repo.UpdateStatus(ctx, companyID, id, entity.BudgetStatusApproved)
}

// ActivateBudget activates an approved budget
func (uc *BudgetUsecase) ActivateBudget(ctx context.Context, companyID, id uuid.UUID) error {
	budget, err := uc.repo.GetByID(ctx, companyID, id)
	if err != nil {
		return common.WrapErr("get budget", err)
	}
	if budget.Status != entity.BudgetStatusApproved {
		return common.NewValidationError("can only activate approved budgets")
	}
	return uc.repo.UpdateStatus(ctx, companyID, id, entity.BudgetStatusActive)
}

// GetBudgetVsActual compares budget to actual spending
func (uc *BudgetUsecase) GetBudgetVsActual(ctx context.Context, companyID, budgetID uuid.UUID) (*entity.BudgetVsActual, error) {
	return uc.repo.GetBudgetVsActual(ctx, companyID, budgetID)
}
