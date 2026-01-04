package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaxRateRepository struct {
	db *pgxpool.Pool
}

func NewTaxRateRepository(db *pgxpool.Pool) *TaxRateRepository {
	return &TaxRateRepository{db: db}
}

func (r *TaxRateRepository) Create(ctx context.Context, rate *entity.TaxRate) error {
	query := `
		INSERT INTO tax_rates (
			id, company_id, name, code, type, rate, 
			sales_account_id, purchase_account_id, description, 
			is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		rate.ID, rate.CompanyID, rate.Name, rate.Code, rate.Type, rate.Rate,
		rate.SalesAccountID, rate.PurchaseAccountID, rate.Description,
		rate.IsActive, rate.CreatedAt, rate.UpdatedAt,
	)
	return err
}

func (r *TaxRateRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.TaxRate, error) {
	query := `
		SELECT id, company_id, name, code, type, rate, 
			   sales_account_id, purchase_account_id, description, 
			   is_active, created_at, updated_at
		FROM tax_rates
		WHERE company_id = $1 AND id = $2
	`
	var rate entity.TaxRate
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&rate.ID, &rate.CompanyID, &rate.Name, &rate.Code, &rate.Type, &rate.Rate,
		&rate.SalesAccountID, &rate.PurchaseAccountID, &rate.Description,
		&rate.IsActive, &rate.CreatedAt, &rate.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *TaxRateRepository) List(ctx context.Context, companyID uuid.UUID) ([]entity.TaxRate, error) {
	query := `
		SELECT id, company_id, name, code, type, rate, 
			   sales_account_id, purchase_account_id, description, 
			   is_active, created_at, updated_at
		FROM tax_rates
		WHERE company_id = $1
		ORDER BY name
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []entity.TaxRate
	for rows.Next() {
		var rate entity.TaxRate
		if err := rows.Scan(
			&rate.ID, &rate.CompanyID, &rate.Name, &rate.Code, &rate.Type, &rate.Rate,
			&rate.SalesAccountID, &rate.PurchaseAccountID, &rate.Description,
			&rate.IsActive, &rate.CreatedAt, &rate.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}
	return rates, nil
}

func (r *TaxRateRepository) Update(ctx context.Context, rate *entity.TaxRate) error {
	rate.UpdatedAt = time.Now()
	query := `
		UPDATE tax_rates
		SET name = $3, code = $4, rate = $5, description = $6, is_active = $7, updated_at = $8
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query,
		rate.CompanyID, rate.ID, rate.Name, rate.Code, rate.Rate, rate.Description, rate.IsActive, rate.UpdatedAt,
	)
	return err
}

func (r *TaxRateRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM tax_rates WHERE company_id = $1 AND id = $2", companyID, id)
	return err
}

// TaxReturnRepository
type TaxReturnRepository struct {
	db *pgxpool.Pool
}

func NewTaxReturnRepository(db *pgxpool.Pool) *TaxReturnRepository {
	return &TaxReturnRepository{db: db}
}

func (r *TaxReturnRepository) Create(ctx context.Context, ret *entity.TaxReturn) error {
	query := `
		INSERT INTO tax_returns (
			id, company_id, period_id, tax_rate_id, return_no, return_date,
			taxable_amount, tax_amount, credits, payable_amount,
			status, notes, journal_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.Exec(ctx, query,
		ret.ID, ret.CompanyID, ret.PeriodID, ret.TaxRateID, ret.ReturnNo, ret.ReturnDate,
		ret.TaxableAmount, ret.TaxAmount, ret.Credits, ret.PayableAmount,
		ret.Status, ret.Notes, ret.JournalID, ret.CreatedAt, ret.UpdatedAt,
	)
	return err
}

func (r *TaxReturnRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.TaxReturn, error) {
	query := `
		SELECT id, company_id, period_id, tax_rate_id, return_no, return_date,
			   taxable_amount, tax_amount, credits, payable_amount,
			   status, notes, journal_id, created_at, updated_at
		FROM tax_returns
		WHERE company_id = $1 AND id = $2
	`
	var ret entity.TaxReturn
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&ret.ID, &ret.CompanyID, &ret.PeriodID, &ret.TaxRateID, &ret.ReturnNo, &ret.ReturnDate,
		&ret.TaxableAmount, &ret.TaxAmount, &ret.Credits, &ret.PayableAmount,
		&ret.Status, &ret.Notes, &ret.JournalID, &ret.CreatedAt, &ret.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (r *TaxReturnRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.TaxReturn, int64, error) {
	countQuery := "SELECT COUNT(*) FROM tax_returns WHERE company_id = $1"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, company_id, period_id, tax_rate_id, return_no, return_date,
			   taxable_amount, tax_amount, credits, payable_amount,
			   status, notes, journal_id, created_at, updated_at
		FROM tax_returns
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var returns []entity.TaxReturn
	for rows.Next() {
		var ret entity.TaxReturn
		if err := rows.Scan(
			&ret.ID, &ret.CompanyID, &ret.PeriodID, &ret.TaxRateID, &ret.ReturnNo, &ret.ReturnDate,
			&ret.TaxableAmount, &ret.TaxAmount, &ret.Credits, &ret.PayableAmount,
			&ret.Status, &ret.Notes, &ret.JournalID, &ret.CreatedAt, &ret.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		returns = append(returns, ret)
	}
	return returns, total, nil
}

func (r *TaxReturnRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.TaxReturnStatus, journalID *uuid.UUID) error {
	query := "UPDATE tax_returns SET status = $1, journal_id = $2, updated_at = $3 WHERE company_id = $4 AND id = $5"
	_, err := r.db.Exec(ctx, query, status, journalID, time.Now(), companyID, id)
	return err
}
