package warehouse

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

var (
	ErrWarehouseNotFound   = errors.New("warehouse not found")
	ErrWarehouseCodeExists = errors.New("warehouse code already exists")
	ErrCannotDeleteDefault = errors.New("cannot delete default warehouse")
)

type WarehouseUsecase struct {
	warehouseRepo repository.WarehouseRepository
}

func NewWarehouseUsecase(wr repository.WarehouseRepository) *WarehouseUsecase {
	return &WarehouseUsecase{warehouseRepo: wr}
}

type CreateWarehouseInput struct {
	CompanyID uuid.UUID
	Code      string
	Name      string
	Address   string
}

func (uc *WarehouseUsecase) CreateWarehouse(ctx context.Context, input CreateWarehouseInput) (*entity.Warehouse, error) {
	exists, _ := uc.warehouseRepo.ExistsByCode(ctx, input.CompanyID, input.Code)
	if exists {
		return nil, ErrWarehouseCodeExists
	}
	w := entity.NewWarehouse(input.CompanyID, input.Code, input.Name)
	w.Address = input.Address
	if err := uc.warehouseRepo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("failed to create warehouse: %w", err)
	}
	return w, nil
}

func (uc *WarehouseUsecase) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Warehouse, error) {
	w, err := uc.warehouseRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrWarehouseNotFound
	}
	return w, nil
}

func (uc *WarehouseUsecase) List(ctx context.Context, companyID uuid.UUID) ([]entity.Warehouse, error) {
	return uc.warehouseRepo.List(ctx, companyID)
}

func (uc *WarehouseUsecase) SetDefault(ctx context.Context, companyID, id uuid.UUID) error {
	_, err := uc.warehouseRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return ErrWarehouseNotFound
	}
	return uc.warehouseRepo.SetDefault(ctx, companyID, id)
}

func (uc *WarehouseUsecase) Update(ctx context.Context, companyID, id uuid.UUID, name, address string) (*entity.Warehouse, error) {
	w, err := uc.warehouseRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrWarehouseNotFound
	}
	w.Name = name
	w.Address = address
	if err := uc.warehouseRepo.Update(ctx, w); err != nil {
		return nil, fmt.Errorf("failed to update warehouse: %w", err)
	}
	return w, nil
}

func (uc *WarehouseUsecase) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	w, err := uc.warehouseRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return ErrWarehouseNotFound
	}
	if w.IsDefault {
		return ErrCannotDeleteDefault
	}
	return uc.warehouseRepo.Delete(ctx, companyID, id)
}
