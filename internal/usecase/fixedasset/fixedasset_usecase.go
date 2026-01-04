package fixedasset

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

// FixedAssetRepository interface
type FixedAssetRepository interface {
	Create(ctx context.Context, asset *entity.FixedAsset) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.FixedAsset, error)
	List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.FixedAsset, int64, error)
	Update(ctx context.Context, asset *entity.FixedAsset) error
	GetActiveAssets(ctx context.Context, companyID uuid.UUID) ([]entity.FixedAsset, error)
	CreateDepreciationEntry(ctx context.Context, entry *entity.DepreciationEntry) error
}

// AssetCategoryRepository interface
type AssetCategoryRepository interface {
	Create(ctx context.Context, cat *entity.AssetCategory) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.AssetCategory, error)
	List(ctx context.Context, companyID uuid.UUID) ([]entity.AssetCategory, error)
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

// FixedAssetUsecase handles fixed asset business logic
type FixedAssetUsecase struct {
	assetRepo    FixedAssetRepository
	categoryRepo AssetCategoryRepository
	journalRepo  repository.JournalRepository
	periodRepo   repository.PeriodRepository
	txManager    repository.TransactionManager
}

// NewFixedAssetUsecase creates a new FixedAssetUsecase
func NewFixedAssetUsecase(
	ar FixedAssetRepository,
	cr AssetCategoryRepository,
	jr repository.JournalRepository,
	pr repository.PeriodRepository,
	tm repository.TransactionManager,
) *FixedAssetUsecase {
	return &FixedAssetUsecase{
		assetRepo:    ar,
		categoryRepo: cr,
		journalRepo:  jr,
		periodRepo:   pr,
		txManager:    tm,
	}
}

// CreateAssetInput holds input for asset creation
type CreateAssetInput struct {
	AssetCode          string
	Name               string
	Description        string
	CategoryID         uuid.UUID
	AcquisitionDate    time.Time
	AcquisitionCost    decimal.Decimal
	ResidualValue      decimal.Decimal
	UsefulLifeMonths   int
	DepreciationMethod entity.DepreciationMethod
	AssetAccountID     uuid.UUID
	DeprAccountID      uuid.UUID
	AccumDeprAccountID uuid.UUID
	Location           string
}

// CreateAsset creates a new fixed asset
func (uc *FixedAssetUsecase) CreateAsset(ctx context.Context, companyID uuid.UUID, input CreateAssetInput) (*entity.FixedAsset, error) {
	if input.AcquisitionCost.LessThanOrEqual(decimal.Zero) {
		return nil, common.NewValidationError("acquisition cost must be positive")
	}
	if input.UsefulLifeMonths <= 0 {
		return nil, common.NewValidationError("useful life must be positive")
	}

	asset := &entity.FixedAsset{
		ID:                 uuid.New(),
		CompanyID:          companyID,
		AssetCode:          input.AssetCode,
		Name:               input.Name,
		Description:        input.Description,
		CategoryID:         input.CategoryID,
		AcquisitionDate:    input.AcquisitionDate,
		AcquisitionCost:    input.AcquisitionCost,
		ResidualValue:      input.ResidualValue,
		UsefulLifeMonths:   input.UsefulLifeMonths,
		DepreciationMethod: input.DepreciationMethod,
		AccumulatedDepr:    decimal.Zero,
		NetBookValue:       input.AcquisitionCost,
		AssetAccountID:     input.AssetAccountID,
		DeprAccountID:      input.DeprAccountID,
		AccumDeprAccountID: input.AccumDeprAccountID,
		Status:             entity.AssetStatusActive,
		DisposalAmount:     decimal.Zero,
		Location:           input.Location,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := uc.assetRepo.Create(ctx, asset); err != nil {
		return nil, common.WrapErr("create asset", err)
	}

	return asset, nil
}

// GetAsset retrieves an asset by ID
func (uc *FixedAssetUsecase) GetAsset(ctx context.Context, companyID, id uuid.UUID) (*entity.FixedAsset, error) {
	return uc.assetRepo.GetByID(ctx, companyID, id)
}

// ListAssets lists assets with pagination
func (uc *FixedAssetUsecase) ListAssets(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.FixedAsset, int64, error) {
	offset := (page - 1) * pageSize
	return uc.assetRepo.List(ctx, companyID, pageSize, offset)
}

// CalculateDepreciation runs depreciation for all active assets for a period
func (uc *FixedAssetUsecase) CalculateDepreciation(ctx context.Context, companyID, periodID, userID uuid.UUID) ([]entity.DepreciationEntry, error) {
	// Get period
	period, err := uc.periodRepo.GetByID(ctx, periodID)
	if err != nil {
		return nil, common.WrapErr("get period", err)
	}
	if period.Status != entity.PeriodStatusOpen {
		return nil, common.NewValidationError("period is not open")
	}

	// Get active assets
	assets, err := uc.assetRepo.GetActiveAssets(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get active assets", err)
	}

	var entries []entity.DepreciationEntry

	for _, asset := range assets {
		if asset.NetBookValue.LessThanOrEqual(asset.ResidualValue) {
			continue // Fully depreciated
		}

		monthlyDepr := asset.CalculateMonthlyDepreciation()
		if monthlyDepr.IsZero() {
			continue
		}

		// Don't depreciate below residual value
		maxDepr := asset.NetBookValue.Sub(asset.ResidualValue)
		if monthlyDepr.GreaterThan(maxDepr) {
			monthlyDepr = maxDepr
		}

		newAccumDepr := asset.AccumulatedDepr.Add(monthlyDepr)
		newNetBookValue := asset.AcquisitionCost.Sub(newAccumDepr)

		entry := entity.DepreciationEntry{
			ID:                uuid.New(),
			CompanyID:         companyID,
			AssetID:           asset.ID,
			PeriodID:          periodID,
			DeprDate:          time.Now(),
			DeprAmount:        monthlyDepr,
			AccumDeprAfter:    newAccumDepr,
			NetBookValueAfter: newNetBookValue,
			CreatedAt:         time.Now(),
		}

		// Create journal entry for depreciation
		jrnCount, _ := uc.journalRepo.CountByYear(ctx, companyID, time.Now().Year())
		journal := &entity.JournalEntry{
			ID:          uuid.New(),
			CompanyID:   companyID,
			PeriodID:    periodID,
			EntryNumber: fmt.Sprintf("JV-%d-%04d", time.Now().Year(), jrnCount+1),
			EntryDate:   time.Now(),
			Description: fmt.Sprintf("Depreciation - %s", asset.Name),
			Status:      entity.JournalStatusPosted,
			SourceType:  "DEPRECIATION",
			SourceID:    &asset.ID,
			CreatedBy:   userID,
			CreatedAt:   time.Now(),
			Lines: []entity.JournalLine{
				{ID: uuid.New(), LineNumber: 1, AccountID: asset.DeprAccountID, Description: "Depreciation Expense", DebitAmount: monthlyDepr, CreditAmount: decimal.Zero},
				{ID: uuid.New(), LineNumber: 2, AccountID: asset.AccumDeprAccountID, Description: "Accumulated Depreciation", DebitAmount: decimal.Zero, CreditAmount: monthlyDepr},
			},
		}

		if err := uc.journalRepo.Create(ctx, journal); err != nil {
			return nil, common.WrapErr("create depreciation journal", err)
		}

		entry.JournalID = journal.ID

		if err := uc.assetRepo.CreateDepreciationEntry(ctx, &entry); err != nil {
			return nil, common.WrapErr("create depreciation entry", err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// DisposeAsset disposes of an asset
func (uc *FixedAssetUsecase) DisposeAsset(ctx context.Context, companyID, assetID uuid.UUID, disposalAmount decimal.Decimal) error {
	asset, err := uc.assetRepo.GetByID(ctx, companyID, assetID)
	if err != nil {
		return common.WrapErr("get asset", err)
	}

	if asset.Status != entity.AssetStatusActive {
		return common.NewValidationError("can only dispose active assets")
	}

	now := time.Now()
	asset.Status = entity.AssetStatusDisposed
	asset.DisposalDate = &now
	asset.DisposalAmount = disposalAmount

	return uc.assetRepo.Update(ctx, asset)
}

// Category methods
func (uc *FixedAssetUsecase) CreateCategory(ctx context.Context, companyID uuid.UUID, name string, lifeMonths int, method entity.DepreciationMethod, assetAccID, deprAccID, accumDeprAccID uuid.UUID) (*entity.AssetCategory, error) {
	cat := &entity.AssetCategory{
		ID:                 uuid.New(),
		CompanyID:          companyID,
		Name:               name,
		DefaultLifeMonths:  lifeMonths,
		DefaultMethod:      method,
		AssetAccountID:     assetAccID,
		DeprAccountID:      deprAccID,
		AccumDeprAccountID: accumDeprAccID,
		CreatedAt:          time.Now(),
	}

	if err := uc.categoryRepo.Create(ctx, cat); err != nil {
		return nil, common.WrapErr("create category", err)
	}

	return cat, nil
}

func (uc *FixedAssetUsecase) ListCategories(ctx context.Context, companyID uuid.UUID) ([]entity.AssetCategory, error) {
	return uc.categoryRepo.List(ctx, companyID)
}
