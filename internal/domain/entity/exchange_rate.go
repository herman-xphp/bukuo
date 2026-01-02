package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ExchangeRate represents an exchange rate between two currencies
type ExchangeRate struct {
	ID             uuid.UUID       `json:"id"`
	CompanyID      uuid.UUID       `json:"company_id"`
	FromCurrencyID uuid.UUID       `json:"from_currency_id"`
	ToCurrencyID   uuid.UUID       `json:"to_currency_id"`
	Rate           decimal.Decimal `json:"rate"` // 1 FROM = Rate TO
	EffectiveDate  time.Time       `json:"effective_date"`
	CreatedAt      time.Time       `json:"created_at"`
}

// NewExchangeRate creates a new exchange rate
func NewExchangeRate(companyID, fromCurrencyID, toCurrencyID uuid.UUID, rate decimal.Decimal, effectiveDate time.Time) *ExchangeRate {
	return &ExchangeRate{
		ID:             uuid.New(),
		CompanyID:      companyID,
		FromCurrencyID: fromCurrencyID,
		ToCurrencyID:   toCurrencyID,
		Rate:           rate,
		EffectiveDate:  effectiveDate,
		CreatedAt:      time.Now(),
	}
}

// Convert converts an amount from one currency to another
func (er *ExchangeRate) Convert(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(er.Rate)
}

// InverseRate returns the inverse rate (TO → FROM)
func (er *ExchangeRate) InverseRate() decimal.Decimal {
	if er.Rate.IsZero() {
		return decimal.Zero
	}
	return decimal.NewFromInt(1).Div(er.Rate)
}
