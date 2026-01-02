package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
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
}

func NewInventoryUsecase(ir repository.InventoryRepository, wr repository.WarehouseRepository, pr repository.ProductRepository) *InventoryUsecase {
	return &InventoryUsecase{invRepo: ir, warehouseRepo: wr, productRepo: pr}
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
	tx := entity.NewInventoryTransaction(input.CompanyID, input.TransactionNo, entity.InventoryTransactionStockIn, input.ProductID, input.WarehouseID, input.Quantity, input.UnitCost)
	tx.Reference = input.Reference
	tx.Notes = input.Notes
	tx.TransactionDate = input.TransactionDate
	if err := uc.invRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
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
		return nil, fmt.Errorf("failed to update stock: %w", err)
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
	stock, _ := uc.invRepo.GetStock(ctx, input.CompanyID, input.ProductID, input.WarehouseID)
	if stock == nil || stock.Quantity.LessThan(input.Quantity) {
		return nil, ErrInsufficientStock
	}
	tx := entity.NewInventoryTransaction(input.CompanyID, input.TransactionNo, entity.InventoryTransactionStockOut, input.ProductID, input.WarehouseID, input.Quantity, stock.AverageCost)
	tx.Reference = input.Reference
	tx.Notes = input.Notes
	tx.TransactionDate = input.TransactionDate
	if err := uc.invRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	stock.Quantity = stock.Quantity.Sub(input.Quantity)
	if err := uc.invRepo.UpdateStock(ctx, stock); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}
	return tx, nil
}

func (uc *InventoryUsecase) GetStock(ctx context.Context, companyID, productID, warehouseID uuid.UUID) (*entity.ProductStock, error) {
	return uc.invRepo.GetStock(ctx, companyID, productID, warehouseID)
}

func (uc *InventoryUsecase) GetStockByProduct(ctx context.Context, companyID, productID uuid.UUID) ([]entity.ProductStock, error) {
	return uc.invRepo.GetStockByProduct(ctx, companyID, productID)
}

func (uc *InventoryUsecase) ListTransactions(ctx context.Context, companyID uuid.UUID, filter repository.InventoryFilter) ([]entity.InventoryTransaction, int64, error) {
	return uc.invRepo.ListTransactions(ctx, companyID, filter)
}
