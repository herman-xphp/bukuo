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

// StockTransferRepository interface
type StockTransferRepository interface {
	Create(ctx context.Context, transfer *entity.StockTransfer, lines []entity.StockTransferLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.StockTransfer, error)
	List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.StockTransfer, int64, error)
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.StockTransferStatus) error
}

// TransferUsecase handles stock transfer business logic
type TransferUsecase struct {
	transferRepo  StockTransferRepository
	inventoryUC   *InventoryUsecase
	inventoryRepo repository.InventoryRepository
	txManager     repository.TransactionManager
}

// NewTransferUsecase creates a new TransferUsecase
func NewTransferUsecase(
	tr StockTransferRepository,
	iuc *InventoryUsecase,
	ir repository.InventoryRepository,
	tm repository.TransactionManager,
) *TransferUsecase {
	return &TransferUsecase{
		transferRepo:  tr,
		inventoryUC:   iuc,
		inventoryRepo: ir,
		txManager:     tm,
	}
}

// CreateTransferInput holds input for transfer creation
type CreateTransferInput struct {
	FromWarehouseID uuid.UUID
	ToWarehouseID   uuid.UUID
	TransferDate    time.Time
	Notes           string
	Lines           []TransferLineInput
}

// TransferLineInput holds input for a transfer line
type TransferLineInput struct {
	ProductID uuid.UUID
	Quantity  decimal.Decimal
}

// CreateTransfer creates a stock transfer and moves stock atomically
func (uc *TransferUsecase) CreateTransfer(ctx context.Context, companyID, userID uuid.UUID, input CreateTransferInput) (*entity.StockTransfer, error) {
	if len(input.Lines) == 0 {
		return nil, common.NewValidationError("transfer must have at least one line")
	}

	if input.FromWarehouseID == input.ToWarehouseID {
		return nil, common.NewValidationError("source and destination warehouse must be different")
	}

	var transfer *entity.StockTransfer

	err := uc.txManager.RunAtomic(ctx, func(ctx context.Context) error {
		transfer = &entity.StockTransfer{
			ID:              uuid.New(),
			CompanyID:       companyID,
			TransferNo:      fmt.Sprintf("TRF-%s", time.Now().Format("20060102150405")),
			FromWarehouseID: input.FromWarehouseID,
			ToWarehouseID:   input.ToWarehouseID,
			TransferDate:    input.TransferDate,
			Status:          entity.StockTransferStatusCompleted, // Immediate transfer
			Notes:           input.Notes,
			CreatedBy:       userID,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		lines := make([]entity.StockTransferLine, len(input.Lines))
		for i, l := range input.Lines {
			// Validate stock availability in source warehouse
			stock, _ := uc.inventoryRepo.GetStock(ctx, companyID, l.ProductID, input.FromWarehouseID)
			if stock == nil || stock.Quantity.LessThan(l.Quantity) {
				return common.NewValidationError(fmt.Sprintf("insufficient stock for product %s", l.ProductID))
			}

			lines[i] = entity.StockTransferLine{
				ID:         uuid.New(),
				TransferID: transfer.ID,
				ProductID:  l.ProductID,
				Quantity:   l.Quantity,
			}

			// Stock Out from source warehouse
			_, err := uc.inventoryUC.StockOut(ctx, StockOutInput{
				CompanyID:   companyID,
				ProductID:   l.ProductID,
				WarehouseID: input.FromWarehouseID,
				Quantity:    l.Quantity,
				Reference:   transfer.TransferNo,
				Notes:       fmt.Sprintf("Transfer to %s", input.ToWarehouseID),
			})
			if err != nil {
				return common.WrapErr("stock out from source", err)
			}

			// Stock In to destination warehouse (use average cost from source)
			_, err = uc.inventoryUC.StockIn(ctx, StockInInput{
				CompanyID:   companyID,
				ProductID:   l.ProductID,
				WarehouseID: input.ToWarehouseID,
				Quantity:    l.Quantity,
				UnitCost:    stock.AverageCost, // Preserve cost
				Reference:   transfer.TransferNo,
				Notes:       fmt.Sprintf("Transfer from %s", input.FromWarehouseID),
			})
			if err != nil {
				return common.WrapErr("stock in to destination", err)
			}
		}

		if err := uc.transferRepo.Create(ctx, transfer, lines); err != nil {
			return common.WrapErr("create transfer", err)
		}

		transfer.Lines = lines
		return nil
	})

	if err != nil {
		return nil, err
	}

	return transfer, nil
}

// GetTransfer retrieves a transfer by ID
func (uc *TransferUsecase) GetTransfer(ctx context.Context, companyID, id uuid.UUID) (*entity.StockTransfer, error) {
	return uc.transferRepo.GetByID(ctx, companyID, id)
}

// ListTransfers lists transfers with pagination
func (uc *TransferUsecase) ListTransfers(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.StockTransfer, int64, error) {
	offset := (page - 1) * pageSize
	return uc.transferRepo.List(ctx, companyID, pageSize, offset)
}
