package sales

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/herman-xphp/bukuo/internal/usecase/inventory"
	"github.com/shopspring/decimal"
)

var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrOrderNotFound   = errors.New("order not found")
)

// SalesUsecase handles sales business logic
type SalesUsecase struct {
	invoiceRepo   repository.SalesInvoiceRepository
	inventoryUC   *inventory.InventoryUsecase
	warehouseRepo repository.WarehouseRepository
	txManager     repository.TransactionManager
}

// NewSalesUsecase creates a new SalesUsecase
func NewSalesUsecase(
	ir repository.SalesInvoiceRepository,
	iuc *inventory.InventoryUsecase,
	wr repository.WarehouseRepository,
	tm repository.TransactionManager,
) *SalesUsecase {
	return &SalesUsecase{invoiceRepo: ir, inventoryUC: iuc, warehouseRepo: wr, txManager: tm}
}

// CreateInvoiceInput represents input for creating an invoice
type CreateInvoiceInput struct {
	CompanyID     uuid.UUID
	InvoiceNo     string
	CustomerID    uuid.UUID
	InvoiceDate   time.Time
	DueDate       time.Time
	Lines         []InvoiceLineInput
	Notes         string
	Status        entity.SalesStatus
	PaymentAmount decimal.Decimal
}

type InvoiceLineInput struct {
	ProductID   uuid.UUID
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal
	DiscountPct decimal.Decimal
	TaxPct      decimal.Decimal
}

// CreateInvoice creates a new sales invoice
func (uc *SalesUsecase) CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*entity.SalesInvoice, error) {
	var inv *entity.SalesInvoice

	err := uc.txManager.RunAtomic(ctx, func(ctx context.Context) error {
		inv = &entity.SalesInvoice{
			ID:          uuid.New(),
			CompanyID:   input.CompanyID,
			InvoiceNo:   input.InvoiceNo,
			CustomerID:  input.CustomerID,
			InvoiceDate: input.InvoiceDate,
			DueDate:     input.DueDate,
			Status:      entity.SalesStatusDraft,
			Notes:       input.Notes,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		subtotal := decimal.Zero
		taxTotal := decimal.Zero
		discountTotal := decimal.Zero

		for _, line := range input.Lines {
			lineSubtotal := line.Quantity.Mul(line.UnitPrice)
			lineDiscount := lineSubtotal.Mul(line.DiscountPct).Div(decimal.NewFromInt(100))
			lineTax := lineSubtotal.Sub(lineDiscount).Mul(line.TaxPct).Div(decimal.NewFromInt(100))
			lineTotal := lineSubtotal.Sub(lineDiscount).Add(lineTax)

			inv.Lines = append(inv.Lines, entity.SalesInvoiceLine{
				ID:          uuid.New(),
				InvoiceID:   inv.ID,
				ProductID:   line.ProductID,
				Description: line.Description,
				Quantity:    line.Quantity,
				UnitPrice:   line.UnitPrice,
				DiscountPct: line.DiscountPct,
				TaxPct:      line.TaxPct,
				LineTotal:   lineTotal,
			})

			subtotal = subtotal.Add(lineSubtotal)
			discountTotal = discountTotal.Add(lineDiscount)
			taxTotal = taxTotal.Add(lineTax)
		}

		inv.Subtotal = subtotal
		inv.DiscountAmount = discountTotal
		inv.TaxAmount = taxTotal
		inv.Total = subtotal.Sub(discountTotal).Add(taxTotal)

		inv.PaidAmount = input.PaymentAmount
		if input.Status != "" {
			inv.Status = input.Status
		} else {
			inv.Status = entity.SalesStatusDraft
		}

		if inv.PaidAmount.GreaterThanOrEqual(inv.Total) && inv.Total.GreaterThan(decimal.Zero) {
			inv.Status = entity.SalesStatusCompleted
		}

		if err := uc.invoiceRepo.Create(ctx, inv, inv.Lines); err != nil {
			return common.WrapErr("create invoice", err)
		}

		if inv.Status == entity.SalesStatusCompleted {
			warehouse, err := uc.warehouseRepo.GetDefault(ctx, input.CompanyID)
			if err != nil {
				return common.WrapErr("get default warehouse", err)
			}

			for _, line := range inv.Lines {
				stockOutInput := inventory.StockOutInput{
					CompanyID:       inv.CompanyID,
					TransactionNo:   fmt.Sprintf("OUT-%s", inv.InvoiceNo),
					ProductID:       line.ProductID,
					WarehouseID:     warehouse.ID,
					Quantity:        line.Quantity,
					Reference:       inv.InvoiceNo,
					Notes:           "Sales Deduct",
					TransactionDate: inv.InvoiceDate,
				}
				if _, err := uc.inventoryUC.StockOut(ctx, stockOutInput); err != nil {
					return common.WrapErr("deduct inventory", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return inv, nil
}

// ListInvoices lists sales invoices
func (uc *SalesUsecase) ListInvoices(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.SalesInvoice, int64, error) {
	filter := repository.SalesFilter{Page: page, PageSize: pageSize}
	invoices, total, err := uc.invoiceRepo.List(ctx, companyID, filter)
	if err != nil {
		return nil, 0, common.WrapErr("list invoices", err)
	}
	return invoices, total, nil
}

// VoidInvoice cancels a sales invoice
func (uc *SalesUsecase) VoidInvoice(ctx context.Context, companyID, invoiceID uuid.UUID) error {
	if _, err := uc.invoiceRepo.GetByID(ctx, companyID, invoiceID); err != nil {
		return err
	}
	if err := uc.invoiceRepo.UpdateStatus(ctx, companyID, invoiceID, entity.SalesStatusCancelled); err != nil {
		return common.WrapErr("void invoice", err)
	}
	return nil
}

// GetInvoice gets a sales invoice by ID
func (uc *SalesUsecase) GetInvoice(ctx context.Context, companyID, invoiceID uuid.UUID) (*entity.SalesInvoice, error) {
	return uc.invoiceRepo.GetByID(ctx, companyID, invoiceID)
}

// CalculateOrderTotal calculates totals for a sales order
func CalculateOrderTotal(lines []entity.SalesOrderLine) (subtotal, tax, discount, total decimal.Decimal) {
	for _, line := range lines {
		lineSubtotal := line.Quantity.Mul(line.UnitPrice)
		lineDiscount := lineSubtotal.Mul(line.DiscountPct).Div(decimal.NewFromInt(100))
		lineTax := lineSubtotal.Sub(lineDiscount).Mul(line.TaxPct).Div(decimal.NewFromInt(100))

		subtotal = subtotal.Add(lineSubtotal)
		discount = discount.Add(lineDiscount)
		tax = tax.Add(lineTax)
	}
	total = subtotal.Sub(discount).Add(tax)
	return
}

// GenerateJournalEntry generates journal entry for a sales invoice
func (uc *SalesUsecase) GenerateJournalEntry(inv *entity.SalesInvoice, arAccountID, salesAccountID, taxAccountID, periodID, createdBy uuid.UUID) *entity.JournalEntry {
	journal := entity.NewJournalEntry(inv.CompanyID, periodID, createdBy, inv.InvoiceDate, fmt.Sprintf("Sales Invoice %s", inv.InvoiceNo))

	journal.Lines = append(journal.Lines, entity.JournalLine{
		ID:           uuid.New(),
		JournalID:    journal.ID,
		AccountID:    arAccountID,
		DebitAmount:  inv.Total,
		CreditAmount: decimal.Zero,
	})

	journal.Lines = append(journal.Lines, entity.JournalLine{
		ID:           uuid.New(),
		JournalID:    journal.ID,
		AccountID:    salesAccountID,
		DebitAmount:  decimal.Zero,
		CreditAmount: inv.Subtotal.Sub(inv.DiscountAmount),
	})

	if inv.TaxAmount.GreaterThan(decimal.Zero) {
		journal.Lines = append(journal.Lines, entity.JournalLine{
			ID:           uuid.New(),
			JournalID:    journal.ID,
			AccountID:    taxAccountID,
			DebitAmount:  decimal.Zero,
			CreditAmount: inv.TaxAmount,
		})
	}

	return journal
}
