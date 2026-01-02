package unit

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

// Errors
var (
	ErrUnitNotFound   = errors.New("unit not found")
	ErrUnitCodeExists = errors.New("unit code already exists")
)

// UnitUsecase handles unit of measure business logic
type UnitUsecase struct {
	unitRepo repository.UnitRepository
}

// NewUnitUsecase creates a new UnitUsecase
func NewUnitUsecase(ur repository.UnitRepository) *UnitUsecase {
	return &UnitUsecase{unitRepo: ur}
}

// CreateUnitInput represents input for creating a unit
type CreateUnitInput struct {
	CompanyID uuid.UUID
	Code      string
	Name      string
}

// CreateUnit creates a new unit of measure
func (uc *UnitUsecase) CreateUnit(ctx context.Context, input CreateUnitInput) (*entity.UnitOfMeasure, error) {
	// Check if code exists
	exists, err := uc.unitRepo.ExistsByCode(ctx, input.CompanyID, input.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check code: %w", err)
	}
	if exists {
		return nil, ErrUnitCodeExists
	}

	unit := entity.NewUnitOfMeasure(input.CompanyID, input.Code, input.Name)

	if err := uc.unitRepo.Create(ctx, unit); err != nil {
		return nil, fmt.Errorf("failed to create unit: %w", err)
	}

	return unit, nil
}

// List retrieves all units for a company
func (uc *UnitUsecase) List(ctx context.Context, companyID uuid.UUID) ([]entity.UnitOfMeasure, error) {
	return uc.unitRepo.List(ctx, companyID)
}

// GetByID retrieves a unit by ID
func (uc *UnitUsecase) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.UnitOfMeasure, error) {
	unit, err := uc.unitRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrUnitNotFound
	}
	return unit, nil
}

// Delete deletes a unit
func (uc *UnitUsecase) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	_, err := uc.unitRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return ErrUnitNotFound
	}
	return uc.unitRepo.Delete(ctx, companyID, id)
}
