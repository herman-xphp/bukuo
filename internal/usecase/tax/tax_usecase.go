package tax

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

// TaxRateRepository interface
type TaxRateRepository interface {
	Create(ctx context.Context, rate *entity.TaxRate) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.TaxRate, error)
	List(ctx context.Context, companyID uuid.UUID) ([]entity.TaxRate, error)
	Update(ctx context.Context, rate *entity.TaxRate) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

// TaxReturnRepository interface
type TaxReturnRepository interface {
	Create(ctx context.Context, ret *entity.TaxReturn) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.TaxReturn, error)
	List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.TaxReturn, int64, error)
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.TaxReturnStatus, journalID *uuid.UUID) error
}

// TaxUsecase handles tax business logic
type TaxUsecase struct {
	rateRepo    TaxRateRepository
	returnRepo  TaxReturnRepository
	journalRepo repository.JournalRepository
	txManager   repository.TransactionManager
}

// NewTaxUsecase creates a new TaxUsecase
func NewTaxUsecase(
	rr TaxRateRepository,
	retr TaxReturnRepository,
	jr repository.JournalRepository,
	tm repository.TransactionManager,
) *TaxUsecase {
	return &TaxUsecase{
		rateRepo:    rr,
		returnRepo:  retr,
		journalRepo: jr,
		txManager:   tm,
	}
}

// CreateTaxRateInput holds input for tax rate creation
type CreateTaxRateInput struct {
	Name              string
	Code              string
	Type              entity.TaxType
	Rate              decimal.Decimal
	SalesAccountID    uuid.UUID
	PurchaseAccountID uuid.UUID
	Description       string
}

// CreateTaxRate creates a new tax rate
func (uc *TaxUsecase) CreateTaxRate(ctx context.Context, companyID uuid.UUID, input CreateTaxRateInput) (*entity.TaxRate, error) {
	if input.Rate.IsNegative() || input.Rate.GreaterThan(decimal.NewFromInt(100)) {
		return nil, common.NewValidationError("rate must be between 0 and 100")
	}

	rate := &entity.TaxRate{
		ID:                uuid.New(),
		CompanyID:         companyID,
		Name:              input.Name,
		Code:              input.Code,
		Type:              input.Type,
		Rate:              input.Rate,
		SalesAccountID:    input.SalesAccountID,
		PurchaseAccountID: input.PurchaseAccountID,
		Description:       input.Description,
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := uc.rateRepo.Create(ctx, rate); err != nil {
		return nil, common.WrapErr("create tax rate", err)
	}

	return rate, nil
}

// ListTaxRates lists all tax rates
func (uc *TaxUsecase) ListTaxRates(ctx context.Context, companyID uuid.UUID) ([]entity.TaxRate, error) {
	return uc.rateRepo.List(ctx, companyID)
}

// GetTaxRate retrieves a tax rate by ID
func (uc *TaxUsecase) GetTaxRate(ctx context.Context, companyID, id uuid.UUID) (*entity.TaxRate, error) {
	return uc.rateRepo.GetByID(ctx, companyID, id)
}

// UpdateTaxRate updates a tax rate
func (uc *TaxUsecase) UpdateTaxRate(ctx context.Context, companyID, id uuid.UUID, name, code, description string, rate decimal.Decimal, isActive bool) error {
	taxRate, err := uc.rateRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return common.WrapErr("get tax rate", err)
	}

	taxRate.Name = name
	taxRate.Code = code
	taxRate.Description = description
	taxRate.Rate = rate
	taxRate.IsActive = isActive

	return uc.rateRepo.Update(ctx, taxRate)
}

// DeleteTaxRate deletes a tax rate
func (uc *TaxUsecase) DeleteTaxRate(ctx context.Context, companyID, id uuid.UUID) error {
	return uc.rateRepo.Delete(ctx, companyID, id)
}

// CreateTaxReturnInput holds input for tax return creation
type CreateTaxReturnInput struct {
	PeriodID      uuid.UUID
	TaxRateID     uuid.UUID
	ReturnDate    time.Time
	TaxableAmount decimal.Decimal // DPP
	TaxAmount     decimal.Decimal // PPN
	Credits       decimal.Decimal // Input VAT
	Notes         string
}

// CreateTaxReturn creates a new tax return (SPT)
func (uc *TaxUsecase) CreateTaxReturn(ctx context.Context, companyID uuid.UUID, input CreateTaxReturnInput) (*entity.TaxReturn, error) {
	payable := input.TaxAmount.Sub(input.Credits)

	ret := &entity.TaxReturn{
		ID:            uuid.New(),
		CompanyID:     companyID,
		PeriodID:      input.PeriodID,
		TaxRateID:     input.TaxRateID,
		ReturnNo:      fmt.Sprintf("SPT-%s", time.Now().Format("20060102150405")),
		ReturnDate:    input.ReturnDate,
		TaxableAmount: input.TaxableAmount,
		TaxAmount:     input.TaxAmount,
		Credits:       input.Credits,
		PayableAmount: payable,
		Status:        entity.TaxReturnDraft,
		Notes:         input.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := uc.returnRepo.Create(ctx, ret); err != nil {
		return nil, common.WrapErr("create tax return", err)
	}

	return ret, nil
}

// ListTaxReturns lists tax returns with pagination
func (uc *TaxUsecase) ListTaxReturns(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.TaxReturn, int64, error) {
	offset := (page - 1) * pageSize
	return uc.returnRepo.List(ctx, companyID, pageSize, offset)
}

// GetTaxReturn retrieves a tax return by ID
func (uc *TaxUsecase) GetTaxReturn(ctx context.Context, companyID, id uuid.UUID) (*entity.TaxReturn, error) {
	return uc.returnRepo.GetByID(ctx, companyID, id)
}

// FileTaxReturn marks tax return as filed
func (uc *TaxUsecase) FileTaxReturn(ctx context.Context, companyID, id uuid.UUID) error {
	ret, err := uc.returnRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return common.WrapErr("get tax return", err)
	}

	if ret.Status != entity.TaxReturnDraft {
		return common.NewValidationError("can only file draft returns")
	}

	return uc.returnRepo.UpdateStatus(ctx, companyID, id, entity.TaxReturnFiled, nil)
}

// PayTaxReturn records payment and creates journal entry
func (uc *TaxUsecase) PayTaxReturn(ctx context.Context, companyID, returnID, userID uuid.UUID, bankAccountID uuid.UUID) error {
	ret, err := uc.returnRepo.GetByID(ctx, companyID, returnID)
	if err != nil {
		return common.WrapErr("get tax return", err)
	}

	if ret.Status == entity.TaxReturnPaid {
		return common.NewValidationError("return already paid")
	}

	if ret.PayableAmount.IsNegative() || ret.PayableAmount.IsZero() {
		// Refund scenario or zero balance - just mark as paid
		return uc.returnRepo.UpdateStatus(ctx, companyID, returnID, entity.TaxReturnPaid, nil)
	}

	// Get tax rate for account IDs
	rate, err := uc.rateRepo.GetByID(ctx, companyID, ret.TaxRateID)
	if err != nil {
		return common.WrapErr("get tax rate", err)
	}

	// Create journal entry for tax payment
	jrnCount, _ := uc.journalRepo.CountByYear(ctx, companyID, time.Now().Year())
	journal := &entity.JournalEntry{
		ID:          uuid.New(),
		CompanyID:   companyID,
		PeriodID:    ret.PeriodID,
		EntryNumber: fmt.Sprintf("JV-%d-%04d", time.Now().Year(), jrnCount+1),
		EntryDate:   time.Now(),
		Description: fmt.Sprintf("Tax Payment - %s", ret.ReturnNo),
		Status:      entity.JournalStatusPosted,
		SourceType:  "TAX_PAYMENT",
		SourceID:    &returnID,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		Lines: []entity.JournalLine{
			{ID: uuid.New(), LineNumber: 1, AccountID: rate.SalesAccountID, Description: "Tax Payable", DebitAmount: ret.PayableAmount, CreditAmount: decimal.Zero},
			{ID: uuid.New(), LineNumber: 2, AccountID: bankAccountID, Description: "Bank Payment", DebitAmount: decimal.Zero, CreditAmount: ret.PayableAmount},
		},
	}

	if err := uc.journalRepo.Create(ctx, journal); err != nil {
		return common.WrapErr("create tax payment journal", err)
	}

	return uc.returnRepo.UpdateStatus(ctx, companyID, returnID, entity.TaxReturnPaid, &journal.ID)
}
