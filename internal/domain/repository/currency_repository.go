package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// CurrencyRepository defines the interface for currency data access
type CurrencyRepository interface {
	// Create creates a new currency
	Create(ctx context.Context, currency *entity.Currency) error

	// GetByID retrieves a currency by ID
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Currency, error)

	// GetByCode retrieves a currency by code
	GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Currency, error)

	// GetBaseCurrency retrieves the base currency for a company
	GetBaseCurrency(ctx context.Context, companyID uuid.UUID) (*entity.Currency, error)

	// List retrieves all active currencies for a company
	List(ctx context.Context, companyID uuid.UUID) ([]entity.Currency, error)

	// Update updates an existing currency
	Update(ctx context.Context, currency *entity.Currency) error

	// SetBaseCurrency sets a currency as the base currency (and unsets others)
	SetBaseCurrency(ctx context.Context, companyID, currencyID uuid.UUID) error

	// Delete soft-deletes a currency
	Delete(ctx context.Context, companyID, id uuid.UUID) error

	// ExistsByCode checks if a currency with the given code exists
	ExistsByCode(ctx context.Context, companyID uuid.UUID, code string) (bool, error)
}

// ExchangeRateRepository defines the interface for exchange rate data access
type ExchangeRateRepository interface {
	// Create creates a new exchange rate
	Create(ctx context.Context, rate *entity.ExchangeRate) error

	// GetByID retrieves an exchange rate by ID
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.ExchangeRate, error)

	// GetRate retrieves the exchange rate for a date
	GetRate(ctx context.Context, companyID, fromCurrencyID, toCurrencyID uuid.UUID, date time.Time) (*entity.ExchangeRate, error)

	// GetLatestRate retrieves the latest exchange rate
	GetLatestRate(ctx context.Context, companyID, fromCurrencyID, toCurrencyID uuid.UUID) (*entity.ExchangeRate, error)

	// List retrieves exchange rates for a company
	List(ctx context.Context, companyID uuid.UUID, fromCurrencyID, toCurrencyID *uuid.UUID) ([]entity.ExchangeRate, error)

	// Update updates an existing exchange rate
	Update(ctx context.Context, rate *entity.ExchangeRate) error

	// Delete deletes an exchange rate
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}
