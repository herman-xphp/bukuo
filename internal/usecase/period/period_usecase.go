package period

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
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
		return nil, common.WrapErr("create period", err)
	}

	return period, nil
}

// GetByID retrieves a period by ID
func (uc *PeriodUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccountingPeriod, error) {
	return uc.periodRepo.GetByID(ctx, id)
}

// GetByCompany retrieves all periods for a company
func (uc *PeriodUsecase) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error) {
	periods, err := uc.periodRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get periods", err)
	}
	return periods, nil
}

func (uc *PeriodUsecase) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.AccountingPeriod, int, error) {
	p := common.ValidatePagination(1, limit)
	if limit <= 0 {
		limit = p.PageSize
	}

	periods, err := uc.periodRepo.List(ctx, companyID, limit, offset, search)
	if err != nil {
		return nil, 0, common.WrapErr("list periods", err)
	}

	total, err := uc.periodRepo.Count(ctx, companyID, search)
	if err != nil {
		return nil, 0, common.WrapErr("count periods", err)
	}

	return periods, total, nil
}

// GetOpenPeriods retrieves open periods for a company
func (uc *PeriodUsecase) GetOpenPeriods(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error) {
	periods, err := uc.periodRepo.GetOpenPeriods(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get open periods", err)
	}
	return periods, nil
}

// UpdatePeriod updates an existing accounting period
func (uc *PeriodUsecase) UpdatePeriod(ctx context.Context, id uuid.UUID, input CreatePeriodInput) (*entity.AccountingPeriod, error) {
	period, err := uc.periodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.EndDate.Before(input.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	period.Name = input.Name
	period.StartDate = input.StartDate
	period.EndDate = input.EndDate

	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, common.WrapErr("update period", err)
	}

	return period, nil
}

// DeletePeriod deletes an accounting period
func (uc *PeriodUsecase) DeletePeriod(ctx context.Context, id uuid.UUID) error {
	if err := uc.periodRepo.Delete(ctx, id); err != nil {
		return common.WrapErr("delete period", err)
	}
	return nil
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
		return nil, common.WrapErr("close period", err)
	}

	return period, nil
}
