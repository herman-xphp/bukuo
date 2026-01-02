package entity

import (
	"time"

	"github.com/google/uuid"
)

// Currency represents a currency (ISO 4217)
type Currency struct {
	ID            uuid.UUID `json:"id"`
	CompanyID     uuid.UUID `json:"company_id"`
	Code          string    `json:"code"`           // ISO 4217: USD, IDR, EUR
	Name          string    `json:"name"`           // US Dollar, Indonesian Rupiah
	Symbol        string    `json:"symbol"`         // $, Rp, €
	DecimalPlaces int       `json:"decimal_places"` // 2 for USD, 0 for IDR
	IsBase        bool      `json:"is_base"`        // Base currency for company
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewCurrency creates a new currency
func NewCurrency(companyID uuid.UUID, code, name, symbol string, decimalPlaces int) *Currency {
	return &Currency{
		ID:            uuid.New(),
		CompanyID:     companyID,
		Code:          code,
		Name:          name,
		Symbol:        symbol,
		DecimalPlaces: decimalPlaces,
		IsBase:        false,
		IsActive:      true,
		CreatedAt:     time.Now(),
	}
}

// CommonCurrencies returns presets for common currencies
var CommonCurrencies = map[string]struct {
	Name          string
	Symbol        string
	DecimalPlaces int
}{
	"IDR": {"Indonesian Rupiah", "Rp", 0},
	"USD": {"US Dollar", "$", 2},
	"EUR": {"Euro", "€", 2},
	"SGD": {"Singapore Dollar", "S$", 2},
	"JPY": {"Japanese Yen", "¥", 0},
	"CNY": {"Chinese Yuan", "¥", 2},
	"GBP": {"British Pound", "£", 2},
	"AUD": {"Australian Dollar", "A$", 2},
	"MYR": {"Malaysian Ringgit", "RM", 2},
}
