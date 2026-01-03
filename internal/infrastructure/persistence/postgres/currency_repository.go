package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/querybuilder"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.CurrencyRepository = (*CurrencyRepository)(nil)
var _ repository.ExchangeRateRepository = (*ExchangeRateRepository)(nil)

type CurrencyRepository struct {
	db *pgxpool.Pool
}

func NewCurrencyRepository(db *pgxpool.Pool) *CurrencyRepository {
	return &CurrencyRepository{db: db}
}

func (r *CurrencyRepository) Create(ctx context.Context, currency *entity.Currency) error {
	query := `
		INSERT INTO currencies (id, company_id, code, name, symbol, decimal_places, is_base, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		currency.ID, currency.CompanyID, currency.Code, currency.Name, currency.Symbol,
		currency.DecimalPlaces, currency.IsBase, currency.IsActive, currency.CreatedAt,
	)
	return err
}

func (r *CurrencyRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Currency, error) {
	query := `
		SELECT id, company_id, code, name, symbol, decimal_places, is_base, is_active, created_at
		FROM currencies WHERE company_id = $1 AND id = $2
	`
	var c entity.Currency
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&c.ID, &c.CompanyID, &c.Code, &c.Name, &c.Symbol,
		&c.DecimalPlaces, &c.IsBase, &c.IsActive, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CurrencyRepository) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Currency, error) {
	query := `
		SELECT id, company_id, code, name, symbol, decimal_places, is_base, is_active, created_at
		FROM currencies WHERE company_id = $1 AND code = $2
	`
	var c entity.Currency
	err := r.db.QueryRow(ctx, query, companyID, code).Scan(
		&c.ID, &c.CompanyID, &c.Code, &c.Name, &c.Symbol,
		&c.DecimalPlaces, &c.IsBase, &c.IsActive, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CurrencyRepository) GetBaseCurrency(ctx context.Context, companyID uuid.UUID) (*entity.Currency, error) {
	query := `
		SELECT id, company_id, code, name, symbol, decimal_places, is_base, is_active, created_at
		FROM currencies WHERE company_id = $1 AND is_base = true
	`
	var c entity.Currency
	err := r.db.QueryRow(ctx, query, companyID).Scan(
		&c.ID, &c.CompanyID, &c.Code, &c.Name, &c.Symbol,
		&c.DecimalPlaces, &c.IsBase, &c.IsActive, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CurrencyRepository) List(ctx context.Context, companyID uuid.UUID) ([]entity.Currency, error) {
	query := `
		SELECT id, company_id, code, name, symbol, decimal_places, is_base, is_active, created_at
		FROM currencies WHERE company_id = $1 AND is_active = true ORDER BY is_base DESC, code
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []entity.Currency
	for rows.Next() {
		var c entity.Currency
		if err := rows.Scan(&c.ID, &c.CompanyID, &c.Code, &c.Name, &c.Symbol,
			&c.DecimalPlaces, &c.IsBase, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		currencies = append(currencies, c)
	}
	return currencies, nil
}

func (r *CurrencyRepository) Update(ctx context.Context, currency *entity.Currency) error {
	query := `
		UPDATE currencies SET name = $3, symbol = $4, decimal_places = $5, is_active = $6
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query, currency.CompanyID, currency.ID, currency.Name,
		currency.Symbol, currency.DecimalPlaces, currency.IsActive)
	return err
}

func (r *CurrencyRepository) SetBaseCurrency(ctx context.Context, companyID, currencyID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE currencies SET is_base = false WHERE company_id = $1`, companyID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE currencies SET is_base = true WHERE company_id = $1 AND id = $2`, companyID, currencyID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *CurrencyRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	query := `UPDATE currencies SET is_active = false WHERE company_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, companyID, id)
	return err
}

func (r *CurrencyRepository) ExistsByCode(ctx context.Context, companyID uuid.UUID, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM currencies WHERE company_id = $1 AND code = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, companyID, code).Scan(&exists)
	return exists, err
}

// ExchangeRateRepository
type ExchangeRateRepository struct {
	db *pgxpool.Pool
}

func NewExchangeRateRepository(db *pgxpool.Pool) *ExchangeRateRepository {
	return &ExchangeRateRepository{db: db}
}

func (r *ExchangeRateRepository) Create(ctx context.Context, rate *entity.ExchangeRate) error {
	query := `
		INSERT INTO exchange_rates (id, company_id, from_currency_id, to_currency_id, rate, effective_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		rate.ID, rate.CompanyID, rate.FromCurrencyID, rate.ToCurrencyID,
		rate.Rate, rate.EffectiveDate, rate.CreatedAt,
	)
	return err
}

func (r *ExchangeRateRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.ExchangeRate, error) {
	query := `
		SELECT id, company_id, from_currency_id, to_currency_id, rate, effective_date, created_at
		FROM exchange_rates WHERE company_id = $1 AND id = $2
	`
	var er entity.ExchangeRate
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&er.ID, &er.CompanyID, &er.FromCurrencyID, &er.ToCurrencyID,
		&er.Rate, &er.EffectiveDate, &er.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &er, nil
}

func (r *ExchangeRateRepository) GetRate(ctx context.Context, companyID, fromCurrencyID, toCurrencyID uuid.UUID, date time.Time) (*entity.ExchangeRate, error) {
	query := `
		SELECT id, company_id, from_currency_id, to_currency_id, rate, effective_date, created_at
		FROM exchange_rates 
		WHERE company_id = $1 AND from_currency_id = $2 AND to_currency_id = $3 AND effective_date <= $4
		ORDER BY effective_date DESC LIMIT 1
	`
	var er entity.ExchangeRate
	err := r.db.QueryRow(ctx, query, companyID, fromCurrencyID, toCurrencyID, date).Scan(
		&er.ID, &er.CompanyID, &er.FromCurrencyID, &er.ToCurrencyID,
		&er.Rate, &er.EffectiveDate, &er.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &er, nil
}

func (r *ExchangeRateRepository) GetLatestRate(ctx context.Context, companyID, fromCurrencyID, toCurrencyID uuid.UUID) (*entity.ExchangeRate, error) {
	return r.GetRate(ctx, companyID, fromCurrencyID, toCurrencyID, time.Now())
}

func (r *ExchangeRateRepository) List(ctx context.Context, companyID uuid.UUID, fromCurrencyID, toCurrencyID *uuid.UUID) ([]entity.ExchangeRate, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if fromCurrencyID != nil {
		qb.AddCondition("from_currency_id = $%d", *fromCurrencyID)
	}
	if toCurrencyID != nil {
		qb.AddCondition("to_currency_id = $%d", *toCurrencyID)
	}

	query := fmt.Sprintf(`
		SELECT id, company_id, from_currency_id, to_currency_id, rate, effective_date, created_at
		FROM exchange_rates WHERE %s ORDER BY effective_date DESC
	`, qb.WhereClause())

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []entity.ExchangeRate
	for rows.Next() {
		var er entity.ExchangeRate
		if err := rows.Scan(&er.ID, &er.CompanyID, &er.FromCurrencyID, &er.ToCurrencyID,
			&er.Rate, &er.EffectiveDate, &er.CreatedAt); err != nil {
			return nil, err
		}
		rates = append(rates, er)
	}
	return rates, nil
}

func (r *ExchangeRateRepository) Update(ctx context.Context, rate *entity.ExchangeRate) error {
	query := `
		UPDATE exchange_rates SET rate = $3, effective_date = $4
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query, rate.CompanyID, rate.ID, rate.Rate, rate.EffectiveDate)
	return err
}

func (r *ExchangeRateRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	query := `DELETE FROM exchange_rates WHERE company_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, companyID, id)
	return err
}
