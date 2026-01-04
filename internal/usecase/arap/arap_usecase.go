package arap

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/shopspring/decimal"
)

// ARAPRepository interface
type ARAPRepository interface {
	GetARAgingReport(ctx context.Context, companyID uuid.UUID) (*entity.ARAgingReport, error)
	GetAPAgingReport(ctx context.Context, companyID uuid.UUID) (*entity.APAgingReport, error)
	GetOutstandingReceivables(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.OutstandingInvoice, int64, decimal.Decimal, error)
	GetOutstandingPayables(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.OutstandingInvoice, int64, decimal.Decimal, error)
}

// ARAPUsecase handles AR/AP reporting business logic
type ARAPUsecase struct {
	repo ARAPRepository
}

// NewARAPUsecase creates a new ARAPUsecase
func NewARAPUsecase(r ARAPRepository) *ARAPUsecase {
	return &ARAPUsecase{repo: r}
}

// GetARAgingReport gets accounts receivable aging report
func (uc *ARAPUsecase) GetARAgingReport(ctx context.Context, companyID uuid.UUID) (*entity.ARAgingReport, error) {
	return uc.repo.GetARAgingReport(ctx, companyID)
}

// GetAPAgingReport gets accounts payable aging report
func (uc *ARAPUsecase) GetAPAgingReport(ctx context.Context, companyID uuid.UUID) (*entity.APAgingReport, error) {
	return uc.repo.GetAPAgingReport(ctx, companyID)
}

// GetOutstandingReceivables lists unpaid sales invoices
func (uc *ARAPUsecase) GetOutstandingReceivables(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.OutstandingInvoice, int64, decimal.Decimal, error) {
	offset := (page - 1) * pageSize
	return uc.repo.GetOutstandingReceivables(ctx, companyID, pageSize, offset)
}

// GetOutstandingPayables lists unpaid purchase invoices
func (uc *ARAPUsecase) GetOutstandingPayables(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.OutstandingInvoice, int64, decimal.Decimal, error) {
	offset := (page - 1) * pageSize
	return uc.repo.GetOutstandingPayables(ctx, companyID, pageSize, offset)
}

// GetARAPSummary provides a quick summary of AR and AP
func (uc *ARAPUsecase) GetARAPSummary(ctx context.Context, companyID uuid.UUID) (map[string]interface{}, error) {
	arReport, err := uc.repo.GetARAgingReport(ctx, companyID)
	if err != nil {
		return nil, err
	}

	apReport, err := uc.repo.GetAPAgingReport(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_receivables": arReport.TotalAR,
		"total_payables":    apReport.TotalAP,
		"net_position":      arReport.TotalAR.Sub(apReport.TotalAP),
		"ar_overdue":        arReport.ThirtyDays.Amount.Add(arReport.SixtyDays.Amount).Add(arReport.NinetyDays.Amount).Add(arReport.OverOneTwenty.Amount),
		"ap_overdue":        apReport.ThirtyDays.Amount.Add(apReport.SixtyDays.Amount).Add(apReport.NinetyDays.Amount).Add(apReport.OverOneTwenty.Amount),
		"customer_count":    len(arReport.ByCustomer),
		"supplier_count":    len(apReport.BySupplier),
	}, nil
}
