package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Verify interface implementation at compile time
var _ repository.UnitRepository = (*UnitRepository)(nil)
var _ repository.CategoryRepository = (*CategoryRepository)(nil)

// UnitRepository implements repository.UnitRepository for PostgreSQL
type UnitRepository struct {
	db *pgxpool.Pool
}

// NewUnitRepository creates a new UnitRepository
func NewUnitRepository(db *pgxpool.Pool) *UnitRepository {
	return &UnitRepository{db: db}
}

func (r *UnitRepository) Create(ctx context.Context, unit *entity.UnitOfMeasure) error {
	query := `
		INSERT INTO units (id, company_id, code, name, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		unit.ID, unit.CompanyID, unit.Code, unit.Name, unit.CreatedAt,
	)
	return err
}

func (r *UnitRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.UnitOfMeasure, error) {
	query := `
		SELECT id, company_id, code, name, created_at
		FROM units WHERE company_id = $1 AND id = $2
	`
	var unit entity.UnitOfMeasure
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&unit.ID, &unit.CompanyID, &unit.Code, &unit.Name, &unit.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (r *UnitRepository) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.UnitOfMeasure, error) {
	query := `
		SELECT id, company_id, code, name, created_at
		FROM units WHERE company_id = $1 AND code = $2
	`
	var unit entity.UnitOfMeasure
	err := r.db.QueryRow(ctx, query, companyID, code).Scan(
		&unit.ID, &unit.CompanyID, &unit.Code, &unit.Name, &unit.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (r *UnitRepository) List(ctx context.Context, companyID uuid.UUID) ([]entity.UnitOfMeasure, error) {
	query := `
		SELECT id, company_id, code, name, created_at
		FROM units WHERE company_id = $1 ORDER BY code
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []entity.UnitOfMeasure
	for rows.Next() {
		var unit entity.UnitOfMeasure
		err := rows.Scan(&unit.ID, &unit.CompanyID, &unit.Code, &unit.Name, &unit.CreatedAt)
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	return units, nil
}

func (r *UnitRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	query := `DELETE FROM units WHERE company_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, companyID, id)
	return err
}

func (r *UnitRepository) ExistsByCode(ctx context.Context, companyID uuid.UUID, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM units WHERE company_id = $1 AND code = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, companyID, code).Scan(&exists)
	return exists, err
}

// CategoryRepository implements repository.CategoryRepository for PostgreSQL
type CategoryRepository struct {
	db *pgxpool.Pool
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *entity.ProductCategory) error {
	query := `
		INSERT INTO product_categories (id, company_id, name, parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		category.ID, category.CompanyID, category.Name, category.ParentID, category.CreatedAt,
	)
	return err
}

func (r *CategoryRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.ProductCategory, error) {
	query := `
		SELECT id, company_id, name, parent_id, created_at
		FROM product_categories WHERE company_id = $1 AND id = $2
	`
	var cat entity.ProductCategory
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&cat.ID, &cat.CompanyID, &cat.Name, &cat.ParentID, &cat.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepository) List(ctx context.Context, companyID uuid.UUID) ([]entity.ProductCategory, error) {
	query := `
		SELECT id, company_id, name, parent_id, created_at
		FROM product_categories WHERE company_id = $1 ORDER BY name
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []entity.ProductCategory
	for rows.Next() {
		var cat entity.ProductCategory
		err := rows.Scan(&cat.ID, &cat.CompanyID, &cat.Name, &cat.ParentID, &cat.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *entity.ProductCategory) error {
	query := `
		UPDATE product_categories SET name = $3, parent_id = $4
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query, category.CompanyID, category.ID, category.Name, category.ParentID)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	query := `DELETE FROM product_categories WHERE company_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, companyID, id)
	return err
}
