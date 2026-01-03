package warehouse

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
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
		return nil, common.WrapErr("create warehouse", err)
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
	warehouses, err := uc.warehouseRepo.List(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("list warehouses", err)
	}
	return warehouses, nil
}

func (uc *WarehouseUsecase) SetDefault(ctx context.Context, companyID, id uuid.UUID) error {
	if _, err := uc.warehouseRepo.GetByID(ctx, companyID, id); err != nil {
		return ErrWarehouseNotFound
	}
	if err := uc.warehouseRepo.SetDefault(ctx, companyID, id); err != nil {
		return common.WrapErr("set default warehouse", err)
	}
	return nil
}

func (uc *WarehouseUsecase) Update(ctx context.Context, companyID, id uuid.UUID, name, address string) (*entity.Warehouse, error) {
	w, err := uc.warehouseRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrWarehouseNotFound
	}

	w.Name = name
	w.Address = address

	if err := uc.warehouseRepo.Update(ctx, w); err != nil {
		return nil, common.WrapErr("update warehouse", err)
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
	if err := uc.warehouseRepo.Delete(ctx, companyID, id); err != nil {
		return common.WrapErr("delete warehouse", err)
	}
	return nil
}
