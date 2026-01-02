package currency

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

// Errors
var (
	ErrCurrencyNotFound   = errors.New("currency not found")
	ErrCurrencyCodeExists = errors.New("currency code already exists")
	ErrCannotDeleteBase   = errors.New("cannot delete base currency")
)

// CurrencyUsecase handles currency business logic
type CurrencyUsecase struct {
	currencyRepo repository.CurrencyRepository
}

// NewCurrencyUsecase creates a new CurrencyUsecase
func NewCurrencyUsecase(cr repository.CurrencyRepository) *CurrencyUsecase {
	return &CurrencyUsecase{currencyRepo: cr}
}

// CreateCurrencyInput represents input for creating a currency
type CreateCurrencyInput struct {
	CompanyID     uuid.UUID
	Code          string
	Name          string
	Symbol        string
	DecimalPlaces int
}

// CreateCurrency creates a new currency
func (uc *CurrencyUsecase) CreateCurrency(ctx context.Context, input CreateCurrencyInput) (*entity.Currency, error) {
	// Check if code exists
	exists, err := uc.currencyRepo.ExistsByCode(ctx, input.CompanyID, input.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check code: %w", err)
	}
	if exists {
		return nil, ErrCurrencyCodeExists
	}

	currency := entity.NewCurrency(input.CompanyID, input.Code, input.Name, input.Symbol, input.DecimalPlaces)

	if err := uc.currencyRepo.Create(ctx, currency); err != nil {
		return nil, fmt.Errorf("failed to create currency: %w", err)
	}

	return currency, nil
}

// CreateFromPreset creates a currency from a preset
func (uc *CurrencyUsecase) CreateFromPreset(ctx context.Context, companyID uuid.UUID, code string) (*entity.Currency, error) {
	preset, ok := entity.CommonCurrencies[code]
	if !ok {
		return nil, fmt.Errorf("unknown currency code: %s", code)
	}

	return uc.CreateCurrency(ctx, CreateCurrencyInput{
		CompanyID:     companyID,
		Code:          code,
		Name:          preset.Name,
		Symbol:        preset.Symbol,
		DecimalPlaces: preset.DecimalPlaces,
	})
}

// GetByID retrieves a currency by ID
func (uc *CurrencyUsecase) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Currency, error) {
	currency, err := uc.currencyRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrCurrencyNotFound
	}
	return currency, nil
}

// GetBaseCurrency retrieves the base currency
func (uc *CurrencyUsecase) GetBaseCurrency(ctx context.Context, companyID uuid.UUID) (*entity.Currency, error) {
	return uc.currencyRepo.GetBaseCurrency(ctx, companyID)
}

// List retrieves all currencies
func (uc *CurrencyUsecase) List(ctx context.Context, companyID uuid.UUID) ([]entity.Currency, error) {
	return uc.currencyRepo.List(ctx, companyID)
}

// SetBaseCurrency sets a currency as the base currency
func (uc *CurrencyUsecase) SetBaseCurrency(ctx context.Context, companyID, currencyID uuid.UUID) error {
	_, err := uc.currencyRepo.GetByID(ctx, companyID, currencyID)
	if err != nil {
		return ErrCurrencyNotFound
	}

	return uc.currencyRepo.SetBaseCurrency(ctx, companyID, currencyID)
}

// UpdateCurrencyInput represents input for updating a currency
type UpdateCurrencyInput struct {
	CompanyID     uuid.UUID
	ID            uuid.UUID
	Name          string
	Symbol        string
	DecimalPlaces int
}

// UpdateCurrency updates an existing currency
func (uc *CurrencyUsecase) UpdateCurrency(ctx context.Context, input UpdateCurrencyInput) (*entity.Currency, error) {
	currency, err := uc.currencyRepo.GetByID(ctx, input.CompanyID, input.ID)
	if err != nil {
		return nil, ErrCurrencyNotFound
	}

	currency.Name = input.Name
	currency.Symbol = input.Symbol
	currency.DecimalPlaces = input.DecimalPlaces

	if err := uc.currencyRepo.Update(ctx, currency); err != nil {
		return nil, fmt.Errorf("failed to update currency: %w", err)
	}

	return currency, nil
}

// DeleteCurrency soft-deletes a currency
func (uc *CurrencyUsecase) DeleteCurrency(ctx context.Context, companyID, id uuid.UUID) error {
	currency, err := uc.currencyRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return ErrCurrencyNotFound
	}

	// Cannot delete base currency
	if currency.IsBase {
		return ErrCannotDeleteBase
	}

	return uc.currencyRepo.Delete(ctx, companyID, id)
}
