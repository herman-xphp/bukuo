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

var _ repository.PeriodRepository = (*PeriodRepository)(nil)

type PeriodRepository struct {
	db *pgxpool.Pool
}

func NewPeriodRepository(db *pgxpool.Pool) *PeriodRepository {
	return &PeriodRepository{db: db}
}

func (r *PeriodRepository) Create(ctx context.Context, period *entity.AccountingPeriod) error {
	query := `
		INSERT INTO accounting_periods (id, company_id, name, start_date, end_date, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		period.ID, period.CompanyID, period.Name,
		period.StartDate, period.EndDate, period.Status, period.CreatedAt,
	)
	return err
}

func (r *PeriodRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccountingPeriod, error) {
	query := `
		SELECT id, company_id, name, start_date, end_date, status, closed_at, closed_by, created_at
		FROM accounting_periods WHERE id = $1
	`
	var p entity.AccountingPeriod
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.CompanyID, &p.Name, &p.StartDate, &p.EndDate,
		&p.Status, &p.ClosedAt, &p.ClosedBy, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PeriodRepository) GetByDate(ctx context.Context, companyID uuid.UUID, date time.Time) (*entity.AccountingPeriod, error) {
	query := `
		SELECT id, company_id, name, start_date, end_date, status, closed_at, closed_by, created_at
		FROM accounting_periods 
		WHERE company_id = $1 AND start_date <= $2 AND end_date >= $2
	`
	var p entity.AccountingPeriod
	err := r.db.QueryRow(ctx, query, companyID, date).Scan(
		&p.ID, &p.CompanyID, &p.Name, &p.StartDate, &p.EndDate,
		&p.Status, &p.ClosedAt, &p.ClosedBy, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PeriodRepository) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error) {
	query := `
		SELECT id, company_id, name, start_date, end_date, status, closed_at, closed_by, created_at
		FROM accounting_periods WHERE company_id = $1 ORDER BY start_date DESC
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []entity.AccountingPeriod
	for rows.Next() {
		var p entity.AccountingPeriod
		if err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.StartDate, &p.EndDate,
			&p.Status, &p.ClosedAt, &p.ClosedBy, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		periods = append(periods, p)
	}
	return periods, nil
}

func (r *PeriodRepository) GetOpenPeriods(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error) {
	query := `
		SELECT id, company_id, name, start_date, end_date, status, closed_at, closed_by, created_at
		FROM accounting_periods WHERE company_id = $1 AND status = 'OPEN' ORDER BY start_date
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []entity.AccountingPeriod
	for rows.Next() {
		var p entity.AccountingPeriod
		if err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.StartDate, &p.EndDate,
			&p.Status, &p.ClosedAt, &p.ClosedBy, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		periods = append(periods, p)
	}
	return periods, nil
}

func (r *PeriodRepository) Update(ctx context.Context, period *entity.AccountingPeriod) error {
	query := `
		UPDATE accounting_periods 
		SET name = $2, start_date = $3, end_date = $4, status = $5, closed_at = $6, closed_by = $7
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		period.ID, period.Name, period.StartDate, period.EndDate,
		period.Status, period.ClosedAt, period.ClosedBy,
	)
	return err
}

func (r *PeriodRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM accounting_periods WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PeriodRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.AccountingPeriod, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if search != "" {
		qb.AddSearch(search, "name", "status")
	}

	whereClause := qb.WhereClause()
	limitPos, offsetPos := qb.AddLimitOffset(limit, offset)

	query := fmt.Sprintf(`
		SELECT id, company_id, name, start_date, end_date, status, closed_at, closed_by, created_at
		FROM accounting_periods WHERE %s ORDER BY start_date DESC LIMIT $%d OFFSET $%d
	`, whereClause, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []entity.AccountingPeriod
	for rows.Next() {
		var p entity.AccountingPeriod
		if err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.StartDate, &p.EndDate,
			&p.Status, &p.ClosedAt, &p.ClosedBy, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		periods = append(periods, p)
	}
	return periods, nil
}

func (r *PeriodRepository) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if search != "" {
		qb.AddSearch(search, "name", "status")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM accounting_periods WHERE %s`, qb.WhereClause())
	var count int
	err := r.db.QueryRow(ctx, query, qb.Args()...).Scan(&count)
	return count, err
}
