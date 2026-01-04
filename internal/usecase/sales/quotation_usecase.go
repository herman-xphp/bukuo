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

// QuotationUsecase handles quotation business logic
type QuotationUsecase struct {
	quotationRepo repository.SalesQuotationRepository
	orderRepo     repository.SalesOrderRepository
	productRepo   repository.ProductRepository
}

// NewQuotationUsecase creates a new QuotationUsecase
func NewQuotationUsecase(
	qr repository.SalesQuotationRepository,
	or repository.SalesOrderRepository,
	pr repository.ProductRepository,
) *QuotationUsecase {
	return &QuotationUsecase{
		quotationRepo: qr,
		orderRepo:     or,
		productRepo:   pr,
	}
}

// CreateQuotationInput holds input for quotation creation
type CreateQuotationInput struct {
	CustomerID    uuid.UUID
	QuotationDate time.Time
	ValidUntil    time.Time
	Notes         string
	Lines         []QuotationLineInput
}

// QuotationLineInput holds input for a quotation line
type QuotationLineInput struct {
	ProductID   uuid.UUID
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal
	DiscountPct decimal.Decimal
	TaxPct      decimal.Decimal
}

// CreateQuotation creates a new quotation
func (uc *QuotationUsecase) CreateQuotation(ctx context.Context, companyID uuid.UUID, input CreateQuotationInput) (*entity.SalesQuotation, error) {
	if len(input.Lines) == 0 {
		return nil, common.NewValidationError("quotation must have at least one line")
	}

	// Calculate totals
	subtotal := decimal.Zero
	taxAmount := decimal.Zero
	discountAmount := decimal.Zero

	lines := make([]entity.SalesQuotationLine, len(input.Lines))
	for i, l := range input.Lines {
		lineTotal := l.Quantity.Mul(l.UnitPrice)
		discount := lineTotal.Mul(l.DiscountPct.Div(decimal.NewFromInt(100)))
		lineAfterDiscount := lineTotal.Sub(discount)
		tax := lineAfterDiscount.Mul(l.TaxPct.Div(decimal.NewFromInt(100)))

		lines[i] = entity.SalesQuotationLine{
			ID:          uuid.New(),
			ProductID:   l.ProductID,
			Description: l.Description,
			Quantity:    l.Quantity,
			UnitPrice:   l.UnitPrice,
			DiscountPct: l.DiscountPct,
			TaxPct:      l.TaxPct,
			LineTotal:   lineAfterDiscount.Add(tax),
		}

		subtotal = subtotal.Add(lineTotal)
		discountAmount = discountAmount.Add(discount)
		taxAmount = taxAmount.Add(tax)
	}

	quotation := &entity.SalesQuotation{
		ID:             uuid.New(),
		CompanyID:      companyID,
		QuotationNo:    fmt.Sprintf("QUO-%s", time.Now().Format("20060102150405")),
		CustomerID:     input.CustomerID,
		QuotationDate:  input.QuotationDate,
		ValidUntil:     input.ValidUntil,
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
		lines[i].QuotationID = quotation.ID
	}

	if err := uc.quotationRepo.Create(ctx, quotation, lines); err != nil {
		return nil, common.WrapErr("create quotation", err)
	}

	quotation.Lines = lines
	return quotation, nil
}

// GetQuotation retrieves a quotation by ID
func (uc *QuotationUsecase) GetQuotation(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesQuotation, error) {
	return uc.quotationRepo.GetByID(ctx, companyID, id)
}

// ListQuotations lists quotations with filters
func (uc *QuotationUsecase) ListQuotations(ctx context.Context, companyID uuid.UUID, filter repository.SalesFilter) ([]entity.SalesQuotation, int64, error) {
	return uc.quotationRepo.List(ctx, companyID, filter)
}

// SendQuotation marks quotation as sent
func (uc *QuotationUsecase) SendQuotation(ctx context.Context, companyID, id uuid.UUID) error {
	return uc.quotationRepo.UpdateStatus(ctx, companyID, id, entity.SalesStatusSent)
}

// AcceptQuotation converts quotation to sales order
func (uc *QuotationUsecase) AcceptQuotation(ctx context.Context, companyID, quotationID uuid.UUID) (*entity.SalesOrder, error) {
	quotation, err := uc.quotationRepo.GetByID(ctx, companyID, quotationID)
	if err != nil {
		return nil, common.WrapErr("get quotation", err)
	}

	if quotation.Status == entity.SalesStatusAccepted {
		return nil, common.NewValidationError("quotation already accepted")
	}

	// Create Sales Order from Quotation
	order := &entity.SalesOrder{
		ID:             uuid.New(),
		CompanyID:      companyID,
		OrderNo:        fmt.Sprintf("SO-%s", time.Now().Format("20060102150405")),
		CustomerID:     quotation.CustomerID,
		QuotationID:    &quotationID,
		OrderDate:      time.Now(),
		Status:         entity.SalesStatusDraft,
		Subtotal:       quotation.Subtotal,
		TaxAmount:      quotation.TaxAmount,
		DiscountAmount: quotation.DiscountAmount,
		Total:          quotation.Total,
		Notes:          quotation.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	orderLines := make([]entity.SalesOrderLine, len(quotation.Lines))
	for i, ql := range quotation.Lines {
		orderLines[i] = entity.SalesOrderLine{
			ID:           uuid.New(),
			OrderID:      order.ID,
			ProductID:    ql.ProductID,
			Description:  ql.Description,
			Quantity:     ql.Quantity,
			DeliveredQty: decimal.Zero,
			UnitPrice:    ql.UnitPrice,
			DiscountPct:  ql.DiscountPct,
			TaxPct:       ql.TaxPct,
			LineTotal:    ql.LineTotal,
		}
	}

	if err := uc.orderRepo.Create(ctx, order, orderLines); err != nil {
		return nil, common.WrapErr("create order from quotation", err)
	}

	// Update quotation status
	if err := uc.quotationRepo.UpdateStatus(ctx, companyID, quotationID, entity.SalesStatusAccepted); err != nil {
		return nil, common.WrapErr("update quotation status", err)
	}

	order.Lines = orderLines
	return order, nil
}

// DeleteQuotation deletes a draft quotation
func (uc *QuotationUsecase) DeleteQuotation(ctx context.Context, companyID, id uuid.UUID) error {
	quotation, err := uc.quotationRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return common.WrapErr("get quotation", err)
	}

	if quotation.Status != entity.SalesStatusDraft {
		return common.NewValidationError("can only delete draft quotations")
	}

	return uc.quotationRepo.Delete(ctx, companyID, id)
}
