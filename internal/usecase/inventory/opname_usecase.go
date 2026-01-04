package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/shopspring/decimal"
)

// StockOpnameRepository interface
type StockOpnameRepository interface {
	Create(ctx context.Context, opname *entity.StockOpname, lines []entity.StockOpnameLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.StockOpname, error)
	List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.StockOpname, int64, error)
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.StockOpnameStatus, approvedBy *uuid.UUID) error
}

// OpnameUsecase handles stock opname business logic
type OpnameUsecase struct {
	opnameRepo    StockOpnameRepository
	inventoryUC   *InventoryUsecase
	inventoryRepo repository.InventoryRepository
	txManager     repository.TransactionManager
}

// NewOpnameUsecase creates a new OpnameUsecase
func NewOpnameUsecase(
	or StockOpnameRepository,
	iuc *InventoryUsecase,
	ir repository.InventoryRepository,
	tm repository.TransactionManager,
) *OpnameUsecase {
	return &OpnameUsecase{
		opnameRepo:    or,
		inventoryUC:   iuc,
		inventoryRepo: ir,
		txManager:     tm,
	}
}

// CreateOpnameInput holds input for opname creation
type CreateOpnameInput struct {
	WarehouseID uuid.UUID
	OpnameDate  time.Time
	Notes       string
	Lines       []OpnameLineInput
}

// OpnameLineInput holds input for an opname line
type OpnameLineInput struct {
	ProductID uuid.UUID
	ActualQty decimal.Decimal
	Notes     string
}

// CreateOpname creates a new stock opname with system quantities
func (uc *OpnameUsecase) CreateOpname(ctx context.Context, companyID, userID uuid.UUID, input CreateOpnameInput) (*entity.StockOpname, error) {
	if len(input.Lines) == 0 {
		return nil, common.NewValidationError("opname must have at least one line")
	}

	opname := &entity.StockOpname{
		ID:          uuid.New(),
		CompanyID:   companyID,
		OpnameNo:    fmt.Sprintf("OPN-%s", time.Now().Format("20060102150405")),
		WarehouseID: input.WarehouseID,
		OpnameDate:  input.OpnameDate,
		Status:      entity.StockOpnameStatusDraft,
		Notes:       input.Notes,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	lines := make([]entity.StockOpnameLine, len(input.Lines))
	for i, l := range input.Lines {
		// Get current system quantity
		stock, _ := uc.inventoryRepo.GetStock(ctx, companyID, l.ProductID, input.WarehouseID)
		systemQty := decimal.Zero
		if stock != nil {
			systemQty = stock.Quantity
		}

		line := entity.StockOpnameLine{
			ID:        uuid.New(),
			OpnameID:  opname.ID,
			ProductID: l.ProductID,
			SystemQty: systemQty,
			ActualQty: l.ActualQty,
			Notes:     l.Notes,
		}
		line.CalculateDifference()
		lines[i] = line
	}

	if err := uc.opnameRepo.Create(ctx, opname, lines); err != nil {
		return nil, common.WrapErr("create opname", err)
	}

	opname.Lines = lines
	return opname, nil
}

// GetOpname retrieves an opname by ID
func (uc *OpnameUsecase) GetOpname(ctx context.Context, companyID, id uuid.UUID) (*entity.StockOpname, error) {
	return uc.opnameRepo.GetByID(ctx, companyID, id)
}

// ListOpnames lists opnames with pagination
func (uc *OpnameUsecase) ListOpnames(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.StockOpname, int64, error) {
	offset := (page - 1) * pageSize
	return uc.opnameRepo.List(ctx, companyID, pageSize, offset)
}

// ApproveOpname approves an opname and creates adjustment transactions
func (uc *OpnameUsecase) ApproveOpname(ctx context.Context, companyID, opnameID, userID uuid.UUID) error {
	opname, err := uc.opnameRepo.GetByID(ctx, companyID, opnameID)
	if err != nil {
		return common.WrapErr("get opname", err)
	}

	if opname.Status != entity.StockOpnameStatusDraft {
		return common.NewValidationError("can only approve draft opnames")
	}

	// Apply adjustments atomically
	err = uc.txManager.RunAtomic(ctx, func(ctx context.Context) error {
		for _, line := range opname.Lines {
			if line.DifferenceQty.IsZero() {
				continue // No adjustment needed
			}

			if line.DifferenceQty.IsPositive() {
				// Stock increase (actual > system) -> StockIn
				_, err := uc.inventoryUC.StockIn(ctx, StockInInput{
					CompanyID:   companyID,
					ProductID:   line.ProductID,
					WarehouseID: opname.WarehouseID,
					Quantity:    line.DifferenceQty,
					UnitCost:    decimal.Zero, // Adjustment has no cost impact (or use average cost)
					Reference:   opname.OpnameNo,
					Notes:       fmt.Sprintf("Stock Opname Adjustment: +%s", line.DifferenceQty.String()),
				})
				if err != nil {
					return common.WrapErr("stock in adjustment", err)
				}
			} else {
				// Stock decrease (actual < system) -> StockOut
				_, err := uc.inventoryUC.StockOut(ctx, StockOutInput{
					CompanyID:   companyID,
					ProductID:   line.ProductID,
					WarehouseID: opname.WarehouseID,
					Quantity:    line.DifferenceQty.Abs(),
					Reference:   opname.OpnameNo,
					Notes:       fmt.Sprintf("Stock Opname Adjustment: %s", line.DifferenceQty.String()),
				})
				if err != nil {
					return common.WrapErr("stock out adjustment", err)
				}
			}
		}

		return uc.opnameRepo.UpdateStatus(ctx, companyID, opnameID, entity.StockOpnameStatusApproved, &userID)
	})

	return err
}
