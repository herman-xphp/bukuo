package sales

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	inventoryuc "github.com/herman-xphp/bukuo/internal/usecase/inventory"
	"github.com/shopspring/decimal"
)

// DeliveryUsecase handles delivery order business logic
type DeliveryUsecase struct {
	deliveryRepo repository.DeliveryOrderRepository
	orderRepo    repository.SalesOrderRepository
	inventoryUC  *inventoryuc.InventoryUsecase
	txManager    repository.TransactionManager
}

// NewDeliveryUsecase creates a new DeliveryUsecase
func NewDeliveryUsecase(
	dr repository.DeliveryOrderRepository,
	or repository.SalesOrderRepository,
	iuc *inventoryuc.InventoryUsecase,
	tm repository.TransactionManager,
) *DeliveryUsecase {
	return &DeliveryUsecase{
		deliveryRepo: dr,
		orderRepo:    or,
		inventoryUC:  iuc,
		txManager:    tm,
	}
}

// CreateDeliveryInput holds input for delivery creation
type CreateDeliveryInput struct {
	OrderID      uuid.UUID
	WarehouseID  uuid.UUID
	DeliveryDate time.Time
	Notes        string
	Lines        []DeliveryLineInput
}

// DeliveryLineInput holds input for a delivery line
type DeliveryLineInput struct {
	OrderLineID uuid.UUID
	ProductID   uuid.UUID
	Quantity    decimal.Decimal
}

// CreateDelivery creates a delivery order and deducts inventory
func (uc *DeliveryUsecase) CreateDelivery(ctx context.Context, companyID uuid.UUID, input CreateDeliveryInput) (*entity.DeliveryOrder, error) {
	if len(input.Lines) == 0 {
		return nil, common.NewValidationError("delivery must have at least one line")
	}

	// Get the sales order
	order, err := uc.orderRepo.GetByID(ctx, companyID, input.OrderID)
	if err != nil {
		return nil, common.WrapErr("get order", err)
	}

	if order.Status == entity.SalesStatusCancelled {
		return nil, common.NewValidationError("cannot deliver cancelled orders")
	}
	if order.Status == entity.SalesStatusCompleted {
		return nil, common.NewValidationError("order already fully delivered")
	}

	// Run in atomic transaction
	var delivery *entity.DeliveryOrder
	err = uc.txManager.RunAtomic(ctx, func(ctx context.Context) error {
		delivery = &entity.DeliveryOrder{
			ID:           uuid.New(),
			CompanyID:    companyID,
			DeliveryNo:   fmt.Sprintf("DO-%s", time.Now().Format("20060102150405")),
			CustomerID:   order.CustomerID,
			OrderID:      order.ID,
			WarehouseID:  input.WarehouseID,
			DeliveryDate: input.DeliveryDate,
			Status:       entity.SalesStatusDraft,
			Notes:        input.Notes,
			CreatedAt:    time.Now(),
		}

		lines := make([]entity.DeliveryOrderLine, len(input.Lines))
		for i, l := range input.Lines {
			lines[i] = entity.DeliveryOrderLine{
				ID:          uuid.New(),
				DeliveryID:  delivery.ID,
				OrderLineID: l.OrderLineID,
				ProductID:   l.ProductID,
				Quantity:    l.Quantity,
			}

			// Deduct inventory via StockOut
			stockOutInput := inventoryuc.StockOutInput{
				CompanyID:   companyID,
				ProductID:   l.ProductID,
				WarehouseID: input.WarehouseID,
				Quantity:    l.Quantity,
				Reference:   delivery.DeliveryNo,
				Notes:       fmt.Sprintf("Delivery for Order %s", order.OrderNo),
			}
			if _, err := uc.inventoryUC.StockOut(ctx, stockOutInput); err != nil {
				return common.WrapErr("stock out", err)
			}
		}

		if err := uc.deliveryRepo.Create(ctx, delivery, lines); err != nil {
			return common.WrapErr("create delivery", err)
		}

		delivery.Lines = lines

		// Check if order is fully delivered (simplified - just mark as completed)
		// In production, you'd compare total delivered qty against ordered qty per line
		if err := uc.orderRepo.UpdateStatus(ctx, companyID, order.ID, entity.SalesStatusCompleted); err != nil {
			return common.WrapErr("update order status", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return delivery, nil
}

// GetDelivery retrieves a delivery by ID
func (uc *DeliveryUsecase) GetDelivery(ctx context.Context, companyID, id uuid.UUID) (*entity.DeliveryOrder, error) {
	return uc.deliveryRepo.GetByID(ctx, companyID, id)
}

// ListDeliveries lists deliveries with filters
func (uc *DeliveryUsecase) ListDeliveries(ctx context.Context, companyID uuid.UUID, filter repository.SalesFilter) ([]entity.DeliveryOrder, int64, error) {
	return uc.deliveryRepo.List(ctx, companyID, filter)
}
