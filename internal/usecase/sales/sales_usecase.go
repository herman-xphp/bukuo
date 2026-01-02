package sales

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/shopspring/decimal"
)

var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrOrderNotFound   = errors.New("order not found")
)

// SalesUsecase handles sales business logic
type SalesUsecase struct {
	// Simplified - in production would have all repos
}

// NewSalesUsecase creates a new SalesUsecase
func NewSalesUsecase() *SalesUsecase {
	return &SalesUsecase{}
}

// CreateInvoiceInput represents input for creating an invoice
type CreateInvoiceInput struct {
	CompanyID   uuid.UUID
	InvoiceNo   string
	CustomerID  uuid.UUID
	InvoiceDate time.Time
	DueDate     time.Time
	Lines       []InvoiceLineInput
	Notes       string
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
	inv := &entity.SalesInvoice{
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
	inv.PaidAmount = decimal.Zero

	return inv, nil
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

	// Debit: AR
	journal.Lines = append(journal.Lines, entity.JournalLine{
		ID:           uuid.New(),
		JournalID:    journal.ID,
		AccountID:    arAccountID,
		DebitAmount:  inv.Total,
		CreditAmount: decimal.Zero,
	})

	// Credit: Sales Revenue
	journal.Lines = append(journal.Lines, entity.JournalLine{
		ID:           uuid.New(),
		JournalID:    journal.ID,
		AccountID:    salesAccountID,
		DebitAmount:  decimal.Zero,
		CreditAmount: inv.Subtotal.Sub(inv.DiscountAmount),
	})

	// Credit: Tax Payable
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
