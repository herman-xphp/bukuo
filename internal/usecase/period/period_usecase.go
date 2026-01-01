package period

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

// PeriodUsecase handles accounting period business logic
type PeriodUsecase struct {
	periodRepo repository.PeriodRepository
}

// NewPeriodUsecase creates a new PeriodUsecase
func NewPeriodUsecase(pr repository.PeriodRepository) *PeriodUsecase {
	return &PeriodUsecase{periodRepo: pr}
}

// CreatePeriodInput represents input for creating a period
type CreatePeriodInput struct {
	CompanyID uuid.UUID
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

// CreatePeriod creates a new accounting period
func (uc *PeriodUsecase) CreatePeriod(ctx context.Context, input CreatePeriodInput) (*entity.AccountingPeriod, error) {
	if input.EndDate.Before(input.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	period := entity.NewAccountingPeriod(input.CompanyID, input.Name, input.StartDate, input.EndDate)

	if err := uc.periodRepo.Create(ctx, period); err != nil {
		return nil, fmt.Errorf("failed to create period: %w", err)
	}

	return period, nil
}

// GetByID retrieves a period by ID
func (uc *PeriodUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccountingPeriod, error) {
	return uc.periodRepo.GetByID(ctx, id)
}

// GetByCompany retrieves all periods for a company
func (uc *PeriodUsecase) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error) {
	return uc.periodRepo.GetByCompany(ctx, companyID)
}

// GetOpenPeriods retrieves open periods for a company
func (uc *PeriodUsecase) GetOpenPeriods(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error) {
	return uc.periodRepo.GetOpenPeriods(ctx, companyID)
}

// ClosePeriod closes an accounting period
func (uc *PeriodUsecase) ClosePeriod(ctx context.Context, id, userID uuid.UUID) (*entity.AccountingPeriod, error) {
	period, err := uc.periodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := period.Close(userID); err != nil {
		return nil, err
	}

	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, err
	}

	return period, nil
}
