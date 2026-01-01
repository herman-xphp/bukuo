package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Verify interface implementation at compile time
var _ repository.PeriodRepository = (*PeriodRepository)(nil)

// PeriodRepository implements repository.PeriodRepository for PostgreSQL
type PeriodRepository struct {
	db *pgxpool.Pool
}

// NewPeriodRepository creates a new PeriodRepository
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
		err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.StartDate, &p.EndDate,
			&p.Status, &p.ClosedAt, &p.ClosedBy, &p.CreatedAt,
		)
		if err != nil {
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
		err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.StartDate, &p.EndDate,
			&p.Status, &p.ClosedAt, &p.ClosedBy, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		periods = append(periods, p)
	}
	return periods, nil
}

func (r *PeriodRepository) Update(ctx context.Context, period *entity.AccountingPeriod) error {
	query := `
		UPDATE accounting_periods 
		SET status = $2, closed_at = $3, closed_by = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		period.ID, period.Status, period.ClosedAt, period.ClosedBy,
	)
	return err
}
