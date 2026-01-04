package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/txhelper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FixedAssetRepository struct {
	db *pgxpool.Pool
}

func NewFixedAssetRepository(db *pgxpool.Pool) *FixedAssetRepository {
	return &FixedAssetRepository{db: db}
}

func (r *FixedAssetRepository) Create(ctx context.Context, asset *entity.FixedAsset) error {
	query := `
		INSERT INTO fixed_assets (
			id, company_id, asset_code, name, description, category_id,
			acquisition_date, acquisition_cost, residual_value, useful_life_months,
			depreciation_method, accumulated_depreciation, net_book_value,
			asset_account_id, depreciation_account_id, accum_depreciation_account_id,
			status, location, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`
	_, err := r.db.Exec(ctx, query,
		asset.ID, asset.CompanyID, asset.AssetCode, asset.Name, asset.Description, asset.CategoryID,
		asset.AcquisitionDate, asset.AcquisitionCost, asset.ResidualValue, asset.UsefulLifeMonths,
		asset.DepreciationMethod, asset.AccumulatedDepr, asset.NetBookValue,
		asset.AssetAccountID, asset.DeprAccountID, asset.AccumDeprAccountID,
		asset.Status, asset.Location, asset.CreatedAt, asset.UpdatedAt,
	)
	return err
}

func (r *FixedAssetRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.FixedAsset, error) {
	query := `
		SELECT id, company_id, asset_code, name, description, category_id,
			   acquisition_date, acquisition_cost, residual_value, useful_life_months,
			   depreciation_method, accumulated_depreciation, net_book_value,
			   asset_account_id, depreciation_account_id, accum_depreciation_account_id,
			   status, disposal_date, disposal_amount, location, created_at, updated_at
		FROM fixed_assets
		WHERE company_id = $1 AND id = $2
	`
	var a entity.FixedAsset
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&a.ID, &a.CompanyID, &a.AssetCode, &a.Name, &a.Description, &a.CategoryID,
		&a.AcquisitionDate, &a.AcquisitionCost, &a.ResidualValue, &a.UsefulLifeMonths,
		&a.DepreciationMethod, &a.AccumulatedDepr, &a.NetBookValue,
		&a.AssetAccountID, &a.DeprAccountID, &a.AccumDeprAccountID,
		&a.Status, &a.DisposalDate, &a.DisposalAmount, &a.Location, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *FixedAssetRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.FixedAsset, int64, error) {
	countQuery := "SELECT COUNT(*) FROM fixed_assets WHERE company_id = $1"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, company_id, asset_code, name, description, category_id,
			   acquisition_date, acquisition_cost, residual_value, useful_life_months,
			   depreciation_method, accumulated_depreciation, net_book_value,
			   asset_account_id, depreciation_account_id, accum_depreciation_account_id,
			   status, disposal_date, disposal_amount, location, created_at, updated_at
		FROM fixed_assets
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var assets []entity.FixedAsset
	for rows.Next() {
		var a entity.FixedAsset
		if err := rows.Scan(
			&a.ID, &a.CompanyID, &a.AssetCode, &a.Name, &a.Description, &a.CategoryID,
			&a.AcquisitionDate, &a.AcquisitionCost, &a.ResidualValue, &a.UsefulLifeMonths,
			&a.DepreciationMethod, &a.AccumulatedDepr, &a.NetBookValue,
			&a.AssetAccountID, &a.DeprAccountID, &a.AccumDeprAccountID,
			&a.Status, &a.DisposalDate, &a.DisposalAmount, &a.Location, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		assets = append(assets, a)
	}

	return assets, total, nil
}

func (r *FixedAssetRepository) Update(ctx context.Context, asset *entity.FixedAsset) error {
	asset.UpdatedAt = time.Now()
	query := `
		UPDATE fixed_assets
		SET name = $3, description = $4, location = $5, 
		    accumulated_depreciation = $6, net_book_value = $7, status = $8,
		    disposal_date = $9, disposal_amount = $10, updated_at = $11
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query,
		asset.CompanyID, asset.ID, asset.Name, asset.Description, asset.Location,
		asset.AccumulatedDepr, asset.NetBookValue, asset.Status,
		asset.DisposalDate, asset.DisposalAmount, asset.UpdatedAt,
	)
	return err
}

func (r *FixedAssetRepository) GetActiveAssets(ctx context.Context, companyID uuid.UUID) ([]entity.FixedAsset, error) {
	query := `
		SELECT id, company_id, asset_code, name, description, category_id,
			   acquisition_date, acquisition_cost, residual_value, useful_life_months,
			   depreciation_method, accumulated_depreciation, net_book_value,
			   asset_account_id, depreciation_account_id, accum_depreciation_account_id,
			   status, disposal_date, disposal_amount, location, created_at, updated_at
		FROM fixed_assets
		WHERE company_id = $1 AND status = 'ACTIVE'
		ORDER BY asset_code
	`

	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []entity.FixedAsset
	for rows.Next() {
		var a entity.FixedAsset
		if err := rows.Scan(
			&a.ID, &a.CompanyID, &a.AssetCode, &a.Name, &a.Description, &a.CategoryID,
			&a.AcquisitionDate, &a.AcquisitionCost, &a.ResidualValue, &a.UsefulLifeMonths,
			&a.DepreciationMethod, &a.AccumulatedDepr, &a.NetBookValue,
			&a.AssetAccountID, &a.DeprAccountID, &a.AccumDeprAccountID,
			&a.Status, &a.DisposalDate, &a.DisposalAmount, &a.Location, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}

	return assets, nil
}

func (r *FixedAssetRepository) CreateDepreciationEntry(ctx context.Context, entry *entity.DepreciationEntry) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		// Insert depreciation entry
		query := `
			INSERT INTO depreciation_entries (
				id, company_id, asset_id, period_id, depreciation_date,
				depreciation_amount, accum_depreciation_after, net_book_value_after,
				journal_id, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		if _, err := tx.Exec(ctx, query,
			entry.ID, entry.CompanyID, entry.AssetID, entry.PeriodID, entry.DeprDate,
			entry.DeprAmount, entry.AccumDeprAfter, entry.NetBookValueAfter,
			entry.JournalID, entry.CreatedAt,
		); err != nil {
			return err
		}

		// Update asset accumulated depreciation and net book value
		updateQuery := `
			UPDATE fixed_assets
			SET accumulated_depreciation = $3, net_book_value = $4, updated_at = $5
			WHERE company_id = $1 AND id = $2
		`
		_, err := tx.Exec(ctx, updateQuery,
			entry.CompanyID, entry.AssetID, entry.AccumDeprAfter, entry.NetBookValueAfter, time.Now(),
		)
		return err
	})
}

// AssetCategory Repository
type AssetCategoryRepository struct {
	db *pgxpool.Pool
}

func NewAssetCategoryRepository(db *pgxpool.Pool) *AssetCategoryRepository {
	return &AssetCategoryRepository{db: db}
}

func (r *AssetCategoryRepository) Create(ctx context.Context, cat *entity.AssetCategory) error {
	query := `
		INSERT INTO asset_categories (
			id, company_id, name, default_life_months, default_method,
			asset_account_id, depreciation_account_id, accum_depreciation_account_id, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		cat.ID, cat.CompanyID, cat.Name, cat.DefaultLifeMonths, cat.DefaultMethod,
		cat.AssetAccountID, cat.DeprAccountID, cat.AccumDeprAccountID, cat.CreatedAt,
	)
	return err
}

func (r *AssetCategoryRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.AssetCategory, error) {
	query := `
		SELECT id, company_id, name, default_life_months, default_method,
			   asset_account_id, depreciation_account_id, accum_depreciation_account_id, created_at
		FROM asset_categories
		WHERE company_id = $1 AND id = $2
	`
	var c entity.AssetCategory
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&c.ID, &c.CompanyID, &c.Name, &c.DefaultLifeMonths, &c.DefaultMethod,
		&c.AssetAccountID, &c.DeprAccountID, &c.AccumDeprAccountID, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *AssetCategoryRepository) List(ctx context.Context, companyID uuid.UUID) ([]entity.AssetCategory, error) {
	query := `
		SELECT id, company_id, name, default_life_months, default_method,
			   asset_account_id, depreciation_account_id, accum_depreciation_account_id, created_at
		FROM asset_categories
		WHERE company_id = $1
		ORDER BY name
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []entity.AssetCategory
	for rows.Next() {
		var c entity.AssetCategory
		if err := rows.Scan(
			&c.ID, &c.CompanyID, &c.Name, &c.DefaultLifeMonths, &c.DefaultMethod,
			&c.AssetAccountID, &c.DeprAccountID, &c.AccumDeprAccountID, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, nil
}

func (r *AssetCategoryRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	// Check if category has assets
	var count int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM fixed_assets WHERE company_id = $1 AND category_id = $2", companyID, id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete category with existing assets")
	}

	_, err := r.db.Exec(ctx, "DELETE FROM asset_categories WHERE company_id = $1 AND id = $2", companyID, id)
	return err
}
