package exchangerate

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

// Errors
var (
	ErrExchangeRateNotFound = errors.New("exchange rate not found")
	ErrSameCurrency         = errors.New("from and to currency cannot be the same")
	ErrInvalidRate          = errors.New("exchange rate must be positive")
)

// ExchangeRateUsecase handles exchange rate business logic
type ExchangeRateUsecase struct {
	exchangeRateRepo repository.ExchangeRateRepository
	currencyRepo     repository.CurrencyRepository
}

// NewExchangeRateUsecase creates a new ExchangeRateUsecase
func NewExchangeRateUsecase(er repository.ExchangeRateRepository, cr repository.CurrencyRepository) *ExchangeRateUsecase {
	return &ExchangeRateUsecase{exchangeRateRepo: er, currencyRepo: cr}
}

// CreateExchangeRateInput represents input for creating an exchange rate
type CreateExchangeRateInput struct {
	CompanyID      uuid.UUID
	FromCurrencyID uuid.UUID
	ToCurrencyID   uuid.UUID
	Rate           decimal.Decimal
	EffectiveDate  time.Time
}

// CreateExchangeRate creates a new exchange rate
func (uc *ExchangeRateUsecase) CreateExchangeRate(ctx context.Context, input CreateExchangeRateInput) (*entity.ExchangeRate, error) {
	if input.FromCurrencyID == input.ToCurrencyID {
		return nil, ErrSameCurrency
	}
	if input.Rate.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidRate
	}

	rate := entity.NewExchangeRate(input.CompanyID, input.FromCurrencyID, input.ToCurrencyID, input.Rate, input.EffectiveDate)

	if err := uc.exchangeRateRepo.Create(ctx, rate); err != nil {
		return nil, common.WrapErr("create exchange rate", err)
	}

	return rate, nil
}

// GetByID retrieves an exchange rate by ID
func (uc *ExchangeRateUsecase) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.ExchangeRate, error) {
	rate, err := uc.exchangeRateRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrExchangeRateNotFound
	}
	return rate, nil
}

// GetRate retrieves the exchange rate for a specific date
func (uc *ExchangeRateUsecase) GetRate(ctx context.Context, companyID, fromCurrencyID, toCurrencyID uuid.UUID, date time.Time) (*entity.ExchangeRate, error) {
	return uc.exchangeRateRepo.GetRate(ctx, companyID, fromCurrencyID, toCurrencyID, date)
}

// GetLatestRate retrieves the latest exchange rate
func (uc *ExchangeRateUsecase) GetLatestRate(ctx context.Context, companyID, fromCurrencyID, toCurrencyID uuid.UUID) (*entity.ExchangeRate, error) {
	return uc.exchangeRateRepo.GetLatestRate(ctx, companyID, fromCurrencyID, toCurrencyID)
}

// List retrieves exchange rates
func (uc *ExchangeRateUsecase) List(ctx context.Context, companyID uuid.UUID, fromCurrencyID, toCurrencyID *uuid.UUID) ([]entity.ExchangeRate, error) {
	rates, err := uc.exchangeRateRepo.List(ctx, companyID, fromCurrencyID, toCurrencyID)
	if err != nil {
		return nil, common.WrapErr("list exchange rates", err)
	}
	return rates, nil
}

// ConvertInput represents input for currency conversion
type ConvertInput struct {
	CompanyID      uuid.UUID
	FromCurrencyID uuid.UUID
	ToCurrencyID   uuid.UUID
	Amount         decimal.Decimal
	Date           time.Time
}

// ConvertOutput represents output for currency conversion
type ConvertOutput struct {
	FromAmount   decimal.Decimal `json:"from_amount"`
	ToAmount     decimal.Decimal `json:"to_amount"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
	Date         time.Time       `json:"date"`
}

// Convert converts an amount from one currency to another
func (uc *ExchangeRateUsecase) Convert(ctx context.Context, input ConvertInput) (*ConvertOutput, error) {
	if input.FromCurrencyID == input.ToCurrencyID {
		return &ConvertOutput{
			FromAmount:   input.Amount,
			ToAmount:     input.Amount,
			ExchangeRate: decimal.NewFromInt(1),
			Date:         input.Date,
		}, nil
	}

	rate, err := uc.exchangeRateRepo.GetRate(ctx, input.CompanyID, input.FromCurrencyID, input.ToCurrencyID, input.Date)
	if err != nil {
		return nil, ErrExchangeRateNotFound
	}

	return &ConvertOutput{
		FromAmount:   input.Amount,
		ToAmount:     rate.Convert(input.Amount),
		ExchangeRate: rate.Rate,
		Date:         rate.EffectiveDate,
	}, nil
}

// UpdateExchangeRate updates an existing exchange rate
func (uc *ExchangeRateUsecase) UpdateExchangeRate(ctx context.Context, companyID, id uuid.UUID, rate decimal.Decimal) (*entity.ExchangeRate, error) {
	if rate.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidRate
	}

	exchangeRate, err := uc.exchangeRateRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrExchangeRateNotFound
	}

	exchangeRate.Rate = rate

	if err := uc.exchangeRateRepo.Update(ctx, exchangeRate); err != nil {
		return nil, common.WrapErr("update exchange rate", err)
	}

	return exchangeRate, nil
}

// DeleteExchangeRate deletes an exchange rate
func (uc *ExchangeRateUsecase) DeleteExchangeRate(ctx context.Context, companyID, id uuid.UUID) error {
	if _, err := uc.exchangeRateRepo.GetByID(ctx, companyID, id); err != nil {
		return ErrExchangeRateNotFound
	}
	if err := uc.exchangeRateRepo.Delete(ctx, companyID, id); err != nil {
		return common.WrapErr("delete exchange rate", err)
	}
	return nil
}
