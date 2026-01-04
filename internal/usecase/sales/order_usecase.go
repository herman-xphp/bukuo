package sales

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

// OrderUsecase handles sales order business logic
type OrderUsecase struct {
	orderRepo    repository.SalesOrderRepository
	deliveryRepo repository.DeliveryOrderRepository
	txManager    repository.TransactionManager
}

// NewOrderUsecase creates a new OrderUsecase
func NewOrderUsecase(
	or repository.SalesOrderRepository,
	dr repository.DeliveryOrderRepository,
	tm repository.TransactionManager,
) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:    or,
		deliveryRepo: dr,
		txManager:    tm,
	}
}

// CreateOrderInput holds input for order creation
type CreateOrderInput struct {
	CustomerID   uuid.UUID
	QuotationID  *uuid.UUID
	OrderDate    time.Time
	DeliveryDate *time.Time
	Notes        string
	Lines        []OrderLineInput
}

// OrderLineInput holds input for an order line
type OrderLineInput struct {
	ProductID   uuid.UUID
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal
	DiscountPct decimal.Decimal
	TaxPct      decimal.Decimal
}

// CreateOrder creates a new sales order
func (uc *OrderUsecase) CreateOrder(ctx context.Context, companyID uuid.UUID, input CreateOrderInput) (*entity.SalesOrder, error) {
	if len(input.Lines) == 0 {
		return nil, common.NewValidationError("order must have at least one line")
	}

	// Calculate totals
	subtotal := decimal.Zero
	taxAmount := decimal.Zero
	discountAmount := decimal.Zero

	lines := make([]entity.SalesOrderLine, len(input.Lines))
	for i, l := range input.Lines {
		lineTotal := l.Quantity.Mul(l.UnitPrice)
		discount := lineTotal.Mul(l.DiscountPct.Div(decimal.NewFromInt(100)))
		lineAfterDiscount := lineTotal.Sub(discount)
		tax := lineAfterDiscount.Mul(l.TaxPct.Div(decimal.NewFromInt(100)))

		lines[i] = entity.SalesOrderLine{
			ID:           uuid.New(),
			ProductID:    l.ProductID,
			Description:  l.Description,
			Quantity:     l.Quantity,
			DeliveredQty: decimal.Zero,
			UnitPrice:    l.UnitPrice,
			DiscountPct:  l.DiscountPct,
			TaxPct:       l.TaxPct,
			LineTotal:    lineAfterDiscount.Add(tax),
		}

		subtotal = subtotal.Add(lineTotal)
		discountAmount = discountAmount.Add(discount)
		taxAmount = taxAmount.Add(tax)
	}

	order := &entity.SalesOrder{
		ID:             uuid.New(),
		CompanyID:      companyID,
		OrderNo:        fmt.Sprintf("SO-%s", time.Now().Format("20060102150405")),
		CustomerID:     input.CustomerID,
		QuotationID:    input.QuotationID,
		OrderDate:      input.OrderDate,
		DeliveryDate:   input.DeliveryDate,
		Status:         entity.SalesStatusDraft,
		Subtotal:       subtotal,
		TaxAmount:      taxAmount,
		DiscountAmount: discountAmount,
		Total:          subtotal.Sub(discountAmount).Add(taxAmount),
		Notes:          input.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	for i := range lines {
		lines[i].OrderID = order.ID
	}

	if err := uc.orderRepo.Create(ctx, order, lines); err != nil {
		return nil, common.WrapErr("create order", err)
	}

	order.Lines = lines
	return order, nil
}

// GetOrder retrieves an order by ID
func (uc *OrderUsecase) GetOrder(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesOrder, error) {
	return uc.orderRepo.GetByID(ctx, companyID, id)
}

// ListOrders lists orders with filters
func (uc *OrderUsecase) ListOrders(ctx context.Context, companyID uuid.UUID, filter repository.SalesFilter) ([]entity.SalesOrder, int64, error) {
	return uc.orderRepo.List(ctx, companyID, filter)
}

// ConfirmOrder confirms a draft order (makes it ready for delivery)
func (uc *OrderUsecase) ConfirmOrder(ctx context.Context, companyID, id uuid.UUID) error {
	order, err := uc.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return common.WrapErr("get order", err)
	}

	if order.Status != entity.SalesStatusDraft {
		return common.NewValidationError("can only confirm draft orders")
	}

	return uc.orderRepo.UpdateStatus(ctx, companyID, id, entity.SalesStatusSent)
}

// CancelOrder cancels an order
func (uc *OrderUsecase) CancelOrder(ctx context.Context, companyID, id uuid.UUID) error {
	order, err := uc.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return common.WrapErr("get order", err)
	}

	if order.Status == entity.SalesStatusCompleted {
		return common.NewValidationError("cannot cancel completed orders")
	}

	return uc.orderRepo.UpdateStatus(ctx, companyID, id, entity.SalesStatusCancelled)
}
