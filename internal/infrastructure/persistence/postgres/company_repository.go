package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Verify interface implementation at compile time
var _ repository.CompanyRepository = (*CompanyRepository)(nil)

// CompanyRepository implements repository.CompanyRepository for PostgreSQL
type CompanyRepository struct {
	db *pgxpool.Pool
}

// NewCompanyRepository creates a new CompanyRepository
func NewCompanyRepository(db *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) Create(ctx context.Context, company *entity.Company) error {
	query := `
		INSERT INTO companies (id, name, tax_id, address, phone, email, fiscal_year_start, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		company.ID, company.Name, company.TaxID, company.Address,
		company.Phone, company.Email, company.FiscalYearStart,
		company.CreatedAt, company.UpdatedAt,
	)
	return err
}

func (r *CompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	query := `
		SELECT id, name, tax_id, address, phone, email, fiscal_year_start, created_at, updated_at
		FROM companies WHERE id = $1
	`
	var c entity.Company
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.TaxID, &c.Address, &c.Phone, &c.Email,
		&c.FiscalYearStart, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CompanyRepository) Update(ctx context.Context, company *entity.Company) error {
	query := `
		UPDATE companies 
		SET name = $2, tax_id = $3, address = $4, phone = $5, email = $6, fiscal_year_start = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		company.ID, company.Name, company.TaxID, company.Address,
		company.Phone, company.Email, company.FiscalYearStart, company.UpdatedAt,
	)
	return err
}

func (r *CompanyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM companies WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
