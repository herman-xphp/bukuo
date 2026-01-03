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

var _ repository.ProductRepository = (*ProductRepository)(nil)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *entity.Product) error {
	query := `
		INSERT INTO products (
			id, company_id, code, name, type, category_id, unit_id,
			description, image_url, sales_price, purchase_price, sales_account_id,
			purchase_account_id, inventory_account_id, is_active, min_stock,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err := r.db.Exec(ctx, query,
		product.ID, product.CompanyID, product.Code, product.Name, product.Type,
		product.CategoryID, product.UnitID, product.Description, product.ImageURL,
		product.SalesPrice, product.PurchasePrice, product.SalesAccountID,
		product.PurchaseAccountID, product.InventoryAccountID, product.IsActive,
		product.MinStock, product.CreatedAt, product.UpdatedAt,
	)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Product, error) {
	query := `
		SELECT id, company_id, code, name, type, category_id, unit_id,
			   description, image_url, sales_price, purchase_price, sales_account_id,
			   purchase_account_id, inventory_account_id, is_active, min_stock,
			   created_at, updated_at
		FROM products 
		WHERE company_id = $1 AND id = $2
	`
	var product entity.Product
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&product.ID, &product.CompanyID, &product.Code, &product.Name, &product.Type,
		&product.CategoryID, &product.UnitID, &product.Description, &product.ImageURL,
		&product.SalesPrice, &product.PurchasePrice, &product.SalesAccountID,
		&product.PurchaseAccountID, &product.InventoryAccountID, &product.IsActive,
		&product.MinStock, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Product, error) {
	query := `
		SELECT id, company_id, code, name, type, category_id, unit_id,
			   description, image_url, sales_price, purchase_price, sales_account_id,
			   purchase_account_id, inventory_account_id, is_active, min_stock,
			   created_at, updated_at
		FROM products 
		WHERE company_id = $1 AND code = $2
	`
	var product entity.Product
	err := r.db.QueryRow(ctx, query, companyID, code).Scan(
		&product.ID, &product.CompanyID, &product.Code, &product.Name, &product.Type,
		&product.CategoryID, &product.UnitID, &product.Description, &product.ImageURL,
		&product.SalesPrice, &product.PurchasePrice, &product.SalesAccountID,
		&product.PurchaseAccountID, &product.InventoryAccountID, &product.IsActive,
		&product.MinStock, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.ProductFilter) ([]entity.Product, int64, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if filter.ProductType != nil {
		qb.AddCondition("type = $%d", *filter.ProductType)
	}
	if filter.CategoryID != nil {
		qb.AddCondition("category_id = $%d", *filter.CategoryID)
	}
	if filter.Search != "" {
		qb.AddSearch(filter.Search, "name", "code")
	}
	if filter.IsActive != nil {
		qb.AddCondition("is_active = $%d", *filter.IsActive)
	}

	whereClause := qb.WhereClause()

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, qb.Args()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	limitPos, offsetPos := qb.AddLimitOffset(pageSize, offset)

	query := fmt.Sprintf(`
		SELECT id, company_id, code, name, type, category_id, unit_id,
			   description, image_url, sales_price, purchase_price, sales_account_id,
			   purchase_account_id, inventory_account_id, is_active, min_stock,
			   created_at, updated_at
		FROM products 
		WHERE %s 
		ORDER BY code
		LIMIT $%d OFFSET $%d
	`, whereClause, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []entity.Product
	for rows.Next() {
		var product entity.Product
		if err := rows.Scan(
			&product.ID, &product.CompanyID, &product.Code, &product.Name, &product.Type,
			&product.CategoryID, &product.UnitID, &product.Description, &product.ImageURL,
			&product.SalesPrice, &product.PurchasePrice, &product.SalesAccountID,
			&product.PurchaseAccountID, &product.InventoryAccountID, &product.IsActive,
			&product.MinStock, &product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}
	return products, total, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *entity.Product) error {
	product.UpdatedAt = time.Now()
	query := `
		UPDATE products 
		SET name = $3, type = $4, category_id = $5, unit_id = $6,
			description = $7, image_url = $8, sales_price = $9, purchase_price = $10,
			sales_account_id = $11, purchase_account_id = $12,
			inventory_account_id = $13, is_active = $14, min_stock = $15,
			updated_at = $16
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query,
		product.CompanyID, product.ID, product.Name, product.Type,
		product.CategoryID, product.UnitID, product.Description, product.ImageURL,
		product.SalesPrice, product.PurchasePrice, product.SalesAccountID,
		product.PurchaseAccountID, product.InventoryAccountID, product.IsActive,
		product.MinStock, product.UpdatedAt,
	)
	return err
}

func (r *ProductRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	query := `UPDATE products SET is_active = false, updated_at = $3 WHERE company_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, companyID, id, time.Now())
	return err
}

func (r *ProductRepository) ExistsByCode(ctx context.Context, companyID uuid.UUID, code string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE company_id = $1 AND code = $2`
	args := []interface{}{companyID, code}

	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}
	query += `)`

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}
