package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DepreciationMethod represents depreciation calculation method
type DepreciationMethod string

const (
	DepreciationStraightLine     DepreciationMethod = "STRAIGHT_LINE"
	DepreciationDecliningBalance DepreciationMethod = "DECLINING_BALANCE"
	DepreciationDoubleDeclining  DepreciationMethod = "DOUBLE_DECLINING"
	DepreciationSumOfYears       DepreciationMethod = "SUM_OF_YEARS"
)

// AssetStatus represents the status of an asset
type AssetStatus string

const (
	AssetStatusActive     AssetStatus = "ACTIVE"
	AssetStatusDisposed   AssetStatus = "DISPOSED"
	AssetStatusSold       AssetStatus = "SOLD"
	AssetStatusWrittenOff AssetStatus = "WRITTEN_OFF"
)

// FixedAsset represents a fixed asset
type FixedAsset struct {
	ID                 uuid.UUID          `json:"id"`
	CompanyID          uuid.UUID          `json:"company_id"`
	AssetCode          string             `json:"asset_code"`
	Name               string             `json:"name"`
	Description        string             `json:"description,omitempty"`
	CategoryID         uuid.UUID          `json:"category_id"`
	AcquisitionDate    time.Time          `json:"acquisition_date"`
	AcquisitionCost    decimal.Decimal    `json:"acquisition_cost"`
	ResidualValue      decimal.Decimal    `json:"residual_value"`
	UsefulLifeMonths   int                `json:"useful_life_months"`
	DepreciationMethod DepreciationMethod `json:"depreciation_method"`
	AccumulatedDepr    decimal.Decimal    `json:"accumulated_depreciation"`
	NetBookValue       decimal.Decimal    `json:"net_book_value"`
	AssetAccountID     uuid.UUID          `json:"asset_account_id"`
	DeprAccountID      uuid.UUID          `json:"depreciation_account_id"`
	AccumDeprAccountID uuid.UUID          `json:"accum_depreciation_account_id"`
	Status             AssetStatus        `json:"status"`
	DisposalDate       *time.Time         `json:"disposal_date,omitempty"`
	DisposalAmount     decimal.Decimal    `json:"disposal_amount"`
	Location           string             `json:"location,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// NewFixedAsset creates a new fixed asset
func NewFixedAsset(companyID uuid.UUID, code, name string, cost, residual decimal.Decimal, lifeMonths int, method DepreciationMethod) *FixedAsset {
	return &FixedAsset{
		ID:                 uuid.New(),
		CompanyID:          companyID,
		AssetCode:          code,
		Name:               name,
		AcquisitionCost:    cost,
		ResidualValue:      residual,
		UsefulLifeMonths:   lifeMonths,
		DepreciationMethod: method,
		AccumulatedDepr:    decimal.Zero,
		NetBookValue:       cost,
		Status:             AssetStatusActive,
		DisposalAmount:     decimal.Zero,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// CalculateMonthlyDepreciation calculates monthly depreciation
func (a *FixedAsset) CalculateMonthlyDepreciation() decimal.Decimal {
	depreciableCost := a.AcquisitionCost.Sub(a.ResidualValue)
	if a.UsefulLifeMonths <= 0 {
		return decimal.Zero
	}
	return depreciableCost.Div(decimal.NewFromInt(int64(a.UsefulLifeMonths)))
}

// AssetCategory represents a category of fixed assets
type AssetCategory struct {
	ID                 uuid.UUID          `json:"id"`
	CompanyID          uuid.UUID          `json:"company_id"`
	Name               string             `json:"name"`
	DefaultLifeMonths  int                `json:"default_life_months"`
	DefaultMethod      DepreciationMethod `json:"default_method"`
	AssetAccountID     uuid.UUID          `json:"asset_account_id"`
	DeprAccountID      uuid.UUID          `json:"depreciation_account_id"`
	AccumDeprAccountID uuid.UUID          `json:"accum_depreciation_account_id"`
	CreatedAt          time.Time          `json:"created_at"`
}

// DepreciationEntry represents a depreciation journal entry
type DepreciationEntry struct {
	ID                uuid.UUID       `json:"id"`
	CompanyID         uuid.UUID       `json:"company_id"`
	AssetID           uuid.UUID       `json:"asset_id"`
	PeriodID          uuid.UUID       `json:"period_id"`
	DeprDate          time.Time       `json:"depreciation_date"`
	DeprAmount        decimal.Decimal `json:"depreciation_amount"`
	AccumDeprAfter    decimal.Decimal `json:"accum_depreciation_after"`
	NetBookValueAfter decimal.Decimal `json:"net_book_value_after"`
	JournalID         uuid.UUID       `json:"journal_id"`
	CreatedAt         time.Time       `json:"created_at"`
}
