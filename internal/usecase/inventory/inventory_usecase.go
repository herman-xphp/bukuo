package inventory

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrTransactionNotFound = errors.New("transaction not found")
)

type InventoryUsecase struct {
	invRepo       repository.InventoryRepository
	warehouseRepo repository.WarehouseRepository
	productRepo   repository.ProductRepository
	txManager     repository.TransactionManager
}

func NewInventoryUsecase(
	ir repository.InventoryRepository,
	wr repository.WarehouseRepository,
	pr repository.ProductRepository,
	tm repository.TransactionManager,
) *InventoryUsecase {
	return &InventoryUsecase{invRepo: ir, warehouseRepo: wr, productRepo: pr, txManager: tm}
}

type StockInInput struct {
	CompanyID       uuid.UUID
	TransactionNo   string
	ProductID       uuid.UUID
	WarehouseID     uuid.UUID
	Quantity        decimal.Decimal
	UnitCost        decimal.Decimal
	Reference       string
	Notes           string
	TransactionDate time.Time
}

func (uc *InventoryUsecase) StockIn(ctx context.Context, input StockInInput) (*entity.InventoryTransaction, error) {
	var tx *entity.InventoryTransaction

	err := uc.txManager.RunAtomic(ctx, func(ctx context.Context) error {
		tx = entity.NewInventoryTransaction(input.CompanyID, input.TransactionNo, entity.InventoryTransactionStockIn, input.ProductID, input.WarehouseID, input.Quantity, input.UnitCost)
		tx.Reference = input.Reference
		tx.Notes = input.Notes
		tx.TransactionDate = input.TransactionDate

		if err := uc.invRepo.CreateTransaction(ctx, tx); err != nil {
			return common.WrapErr("create transaction", err)
		}

		// Update stock with average cost
		stock, _ := uc.invRepo.GetStock(ctx, input.CompanyID, input.ProductID, input.WarehouseID)
		if stock == nil {
			stock = &entity.ProductStock{CompanyID: input.CompanyID, ProductID: input.ProductID, WarehouseID: input.WarehouseID, Quantity: decimal.Zero, AverageCost: decimal.Zero}
		}

		totalValue := stock.Quantity.Mul(stock.AverageCost).Add(input.Quantity.Mul(input.UnitCost))
		newQty := stock.Quantity.Add(input.Quantity)
		if !newQty.IsZero() {
			stock.AverageCost = totalValue.Div(newQty)
		}
		stock.Quantity = newQty

		if err := uc.invRepo.UpdateStock(ctx, stock); err != nil {
			return common.WrapErr("update stock", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tx, nil
}

type StockOutInput struct {
	CompanyID       uuid.UUID
	TransactionNo   string
	ProductID       uuid.UUID
	WarehouseID     uuid.UUID
	Quantity        decimal.Decimal
	Reference       string
	Notes           string
	TransactionDate time.Time
}

func (uc *InventoryUsecase) StockOut(ctx context.Context, input StockOutInput) (*entity.InventoryTransaction, error) {
	var tx *entity.InventoryTransaction

	err := uc.txManager.RunAtomic(ctx, func(ctx context.Context) error {
		stock, _ := uc.invRepo.GetStock(ctx, input.CompanyID, input.ProductID, input.WarehouseID)
		if stock == nil || stock.Quantity.LessThan(input.Quantity) {
			return ErrInsufficientStock
		}

		tx = entity.NewInventoryTransaction(input.CompanyID, input.TransactionNo, entity.InventoryTransactionStockOut, input.ProductID, input.WarehouseID, input.Quantity, stock.AverageCost)
		tx.Reference = input.Reference
		tx.Notes = input.Notes
		tx.TransactionDate = input.TransactionDate

		if err := uc.invRepo.CreateTransaction(ctx, tx); err != nil {
			return common.WrapErr("create transaction", err)
		}

		stock.Quantity = stock.Quantity.Sub(input.Quantity)
		if err := uc.invRepo.UpdateStock(ctx, stock); err != nil {
			return common.WrapErr("update stock", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (uc *InventoryUsecase) GetStock(ctx context.Context, companyID, productID, warehouseID uuid.UUID) (*entity.ProductStock, error) {
	return uc.invRepo.GetStock(ctx, companyID, productID, warehouseID)
}

func (uc *InventoryUsecase) GetStockByProduct(ctx context.Context, companyID, productID uuid.UUID) ([]entity.ProductStock, error) {
	stocks, err := uc.invRepo.GetStockByProduct(ctx, companyID, productID)
	if err != nil {
		return nil, common.WrapErr("get stock by product", err)
	}
	return stocks, nil
}

func (uc *InventoryUsecase) ListTransactions(ctx context.Context, companyID uuid.UUID, filter repository.InventoryFilter) ([]entity.InventoryTransaction, int64, error) {
	txs, total, err := uc.invRepo.ListTransactions(ctx, companyID, filter)
	if err != nil {
		return nil, 0, common.WrapErr("list transactions", err)
	}
	return txs, total, nil
}

func (uc *InventoryUsecase) ListStocks(ctx context.Context, companyID uuid.UUID) ([]entity.ProductStock, error) {
	stocks, err := uc.invRepo.ListStocks(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("list stocks", err)
	}
	return stocks, nil
}
