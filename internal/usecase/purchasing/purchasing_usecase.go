package purchasing

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

// PurchaseOrderRepository interface
type PurchaseOrderRepository interface {
	Create(ctx context.Context, order *entity.PurchaseOrder, lines []entity.PurchaseOrderLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.PurchaseOrder, error)
	List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.PurchaseOrder, int64, error)
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.PurchaseStatus) error
}

// PurchaseInvoiceRepository interface
type PurchaseInvoiceRepository interface {
	Create(ctx context.Context, inv *entity.PurchaseInvoice, lines []entity.PurchaseInvoiceLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.PurchaseInvoice, error)
	List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.PurchaseInvoice, int64, error)
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.PurchaseStatus) error
}

// PurchasingUsecase handles purchasing business logic
type PurchasingUsecase struct {
	orderRepo   PurchaseOrderRepository
	invoiceRepo PurchaseInvoiceRepository
	journalRepo repository.JournalRepository
	txManager   repository.TransactionManager
}

// NewPurchasingUsecase creates a new PurchasingUsecase
func NewPurchasingUsecase(
	or PurchaseOrderRepository,
	ir PurchaseInvoiceRepository,
	jr repository.JournalRepository,
	tm repository.TransactionManager,
) *PurchasingUsecase {
	return &PurchasingUsecase{
		orderRepo:   or,
		invoiceRepo: ir,
		journalRepo: jr,
		txManager:   tm,
	}
}

// CreateOrderInput holds input for purchase order creation
type CreateOrderInput struct {
	SupplierID   uuid.UUID
	OrderDate    time.Time
	DeliveryDate *time.Time
	Notes        string
	Lines        []OrderLineInput
}

type OrderLineInput struct {
	ProductID   uuid.UUID
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal
	DiscountPct decimal.Decimal
	TaxPct      decimal.Decimal
}

// CreateOrder creates a new purchase order
func (uc *PurchasingUsecase) CreateOrder(ctx context.Context, companyID uuid.UUID, input CreateOrderInput) (*entity.PurchaseOrder, error) {
	if len(input.Lines) == 0 {
		return nil, common.NewValidationError("order must have at least one line")
	}

	order := &entity.PurchaseOrder{
		ID:             uuid.New(),
		CompanyID:      companyID,
		OrderNo:        fmt.Sprintf("PO-%s", time.Now().Format("20060102150405")),
		SupplierID:     input.SupplierID,
		OrderDate:      input.OrderDate,
		DeliveryDate:   input.DeliveryDate,
		Status:         entity.PurchaseStatusDraft,
		Subtotal:       decimal.Zero,
		TaxAmount:      decimal.Zero,
		DiscountAmount: decimal.Zero,
		Total:          decimal.Zero,
		Notes:          input.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	lines := make([]entity.PurchaseOrderLine, len(input.Lines))
	for i, l := range input.Lines {
		lineTotal := l.Quantity.Mul(l.UnitPrice)
		discountAmt := lineTotal.Mul(l.DiscountPct).Div(decimal.NewFromInt(100))
		lineTotal = lineTotal.Sub(discountAmt)
		taxAmt := lineTotal.Mul(l.TaxPct).Div(decimal.NewFromInt(100))

		order.Subtotal = order.Subtotal.Add(l.Quantity.Mul(l.UnitPrice))
		order.DiscountAmount = order.DiscountAmount.Add(discountAmt)
		order.TaxAmount = order.TaxAmount.Add(taxAmt)

		lines[i] = entity.PurchaseOrderLine{
			ID:          uuid.New(),
			OrderID:     order.ID,
			ProductID:   l.ProductID,
			Description: l.Description,
			Quantity:    l.Quantity,
			ReceivedQty: decimal.Zero,
			UnitPrice:   l.UnitPrice,
			DiscountPct: l.DiscountPct,
			TaxPct:      l.TaxPct,
			LineTotal:   lineTotal.Add(taxAmt),
		}
	}
	order.Total = order.Subtotal.Sub(order.DiscountAmount).Add(order.TaxAmount)

	if err := uc.orderRepo.Create(ctx, order, lines); err != nil {
		return nil, common.WrapErr("create order", err)
	}

	order.Lines = lines
	return order, nil
}

// ListOrders lists purchase orders
func (uc *PurchasingUsecase) ListOrders(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.PurchaseOrder, int64, error) {
	offset := (page - 1) * pageSize
	return uc.orderRepo.List(ctx, companyID, pageSize, offset)
}

// GetOrder retrieves a purchase order by ID
func (uc *PurchasingUsecase) GetOrder(ctx context.Context, companyID, id uuid.UUID) (*entity.PurchaseOrder, error) {
	return uc.orderRepo.GetByID(ctx, companyID, id)
}

// ApproveOrder approves a purchase order
func (uc *PurchasingUsecase) ApproveOrder(ctx context.Context, companyID, id uuid.UUID) error {
	order, err := uc.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return common.WrapErr("get order", err)
	}
	if order.Status != entity.PurchaseStatusDraft {
		return common.NewValidationError("can only approve draft orders")
	}
	return uc.orderRepo.UpdateStatus(ctx, companyID, id, entity.PurchaseStatusApproved)
}

// CreateInvoiceInput holds input for purchase invoice creation
type CreateInvoiceInput struct {
	SupplierID  uuid.UUID
	OrderID     *uuid.UUID
	InvoiceDate time.Time
	DueDate     time.Time
	Notes       string
	Lines       []InvoiceLineInput
}

type InvoiceLineInput struct {
	ProductID   uuid.UUID
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal
	DiscountPct decimal.Decimal
	TaxPct      decimal.Decimal
}

// CreateInvoice creates a new purchase invoice
func (uc *PurchasingUsecase) CreateInvoice(ctx context.Context, companyID uuid.UUID, input CreateInvoiceInput) (*entity.PurchaseInvoice, error) {
	if len(input.Lines) == 0 {
		return nil, common.NewValidationError("invoice must have at least one line")
	}

	inv := &entity.PurchaseInvoice{
		ID:             uuid.New(),
		CompanyID:      companyID,
		InvoiceNo:      fmt.Sprintf("PINV-%s", time.Now().Format("20060102150405")),
		SupplierID:     input.SupplierID,
		OrderID:        input.OrderID,
		InvoiceDate:    input.InvoiceDate,
		DueDate:        input.DueDate,
		Status:         entity.PurchaseStatusDraft,
		Subtotal:       decimal.Zero,
		TaxAmount:      decimal.Zero,
		DiscountAmount: decimal.Zero,
		Total:          decimal.Zero,
		PaidAmount:     decimal.Zero,
		Notes:          input.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	lines := make([]entity.PurchaseInvoiceLine, len(input.Lines))
	for i, l := range input.Lines {
		lineTotal := l.Quantity.Mul(l.UnitPrice)
		discountAmt := lineTotal.Mul(l.DiscountPct).Div(decimal.NewFromInt(100))
		lineTotal = lineTotal.Sub(discountAmt)
		taxAmt := lineTotal.Mul(l.TaxPct).Div(decimal.NewFromInt(100))

		inv.Subtotal = inv.Subtotal.Add(l.Quantity.Mul(l.UnitPrice))
		inv.DiscountAmount = inv.DiscountAmount.Add(discountAmt)
		inv.TaxAmount = inv.TaxAmount.Add(taxAmt)

		lines[i] = entity.PurchaseInvoiceLine{
			ID:          uuid.New(),
			InvoiceID:   inv.ID,
			ProductID:   l.ProductID,
			Description: l.Description,
			Quantity:    l.Quantity,
			UnitPrice:   l.UnitPrice,
			DiscountPct: l.DiscountPct,
			TaxPct:      l.TaxPct,
			LineTotal:   lineTotal.Add(taxAmt),
		}
	}
	inv.Total = inv.Subtotal.Sub(inv.DiscountAmount).Add(inv.TaxAmount)

	if err := uc.invoiceRepo.Create(ctx, inv, lines); err != nil {
		return nil, common.WrapErr("create invoice", err)
	}

	inv.Lines = lines
	return inv, nil
}

// ListInvoices lists purchase invoices
func (uc *PurchasingUsecase) ListInvoices(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.PurchaseInvoice, int64, error) {
	offset := (page - 1) * pageSize
	return uc.invoiceRepo.List(ctx, companyID, pageSize, offset)
}

// GetInvoice retrieves a purchase invoice by ID
func (uc *PurchasingUsecase) GetInvoice(ctx context.Context, companyID, id uuid.UUID) (*entity.PurchaseInvoice, error) {
	return uc.invoiceRepo.GetByID(ctx, companyID, id)
}
