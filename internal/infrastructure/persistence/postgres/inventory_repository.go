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

var _ repository.WarehouseRepository = (*WarehouseRepository)(nil)
var _ repository.InventoryRepository = (*InventoryRepository)(nil)

type WarehouseRepository struct {
	db *pgxpool.Pool
}

func NewWarehouseRepository(db *pgxpool.Pool) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

func (r *WarehouseRepository) Create(ctx context.Context, w *entity.Warehouse) error {
	query := `INSERT INTO warehouses (id, company_id, code, name, address, is_default, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, w.ID, w.CompanyID, w.Code, w.Name, w.Address, w.IsDefault, w.IsActive, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *WarehouseRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Warehouse, error) {
	query := `SELECT id, company_id, code, name, address, is_default, is_active, created_at, updated_at
		FROM warehouses WHERE company_id = $1 AND id = $2`
	var w entity.Warehouse
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(&w.ID, &w.CompanyID, &w.Code, &w.Name, &w.Address, &w.IsDefault, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WarehouseRepository) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Warehouse, error) {
	query := `SELECT id, company_id, code, name, address, is_default, is_active, created_at, updated_at
		FROM warehouses WHERE company_id = $1 AND code = $2`
	var w entity.Warehouse
	err := r.db.QueryRow(ctx, query, companyID, code).Scan(&w.ID, &w.CompanyID, &w.Code, &w.Name, &w.Address, &w.IsDefault, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WarehouseRepository) GetDefault(ctx context.Context, companyID uuid.UUID) (*entity.Warehouse, error) {
	query := `SELECT id, company_id, code, name, address, is_default, is_active, created_at, updated_at
		FROM warehouses WHERE company_id = $1 AND is_default = true`
	var w entity.Warehouse
	err := r.db.QueryRow(ctx, query, companyID).Scan(&w.ID, &w.CompanyID, &w.Code, &w.Name, &w.Address, &w.IsDefault, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WarehouseRepository) List(ctx context.Context, companyID uuid.UUID) ([]entity.Warehouse, error) {
	query := `SELECT id, company_id, code, name, address, is_default, is_active, created_at, updated_at
		FROM warehouses WHERE company_id = $1 AND is_active = true ORDER BY is_default DESC, code`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var warehouses []entity.Warehouse
	for rows.Next() {
		var w entity.Warehouse
		if err := rows.Scan(&w.ID, &w.CompanyID, &w.Code, &w.Name, &w.Address, &w.IsDefault, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, nil
}

func (r *WarehouseRepository) Update(ctx context.Context, w *entity.Warehouse) error {
	w.UpdatedAt = time.Now()
	query := `UPDATE warehouses SET name = $3, address = $4, is_active = $5, updated_at = $6 WHERE company_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, w.CompanyID, w.ID, w.Name, w.Address, w.IsActive, w.UpdatedAt)
	return err
}

func (r *WarehouseRepository) SetDefault(ctx context.Context, companyID, warehouseID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `UPDATE warehouses SET is_default = false WHERE company_id = $1`, companyID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE warehouses SET is_default = true WHERE company_id = $1 AND id = $2`, companyID, warehouseID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *WarehouseRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	query := `UPDATE warehouses SET is_active = false, updated_at = $3 WHERE company_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, companyID, id, time.Now())
	return err
}

func (r *WarehouseRepository) ExistsByCode(ctx context.Context, companyID uuid.UUID, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM warehouses WHERE company_id = $1 AND code = $2)`, companyID, code).Scan(&exists)
	return exists, err
}

// InventoryRepository implements repository.InventoryRepository
type InventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) CreateTransaction(ctx context.Context, tx *entity.InventoryTransaction) error {
	query := `INSERT INTO inventory_transactions (id, company_id, transaction_no, type, product_id, warehouse_id, to_warehouse_id, quantity, unit_cost, total_cost, reference, notes, transaction_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := r.db.Exec(ctx, query, tx.ID, tx.CompanyID, tx.TransactionNo, tx.Type, tx.ProductID, tx.WarehouseID, tx.ToWarehouseID, tx.Quantity, tx.UnitCost, tx.TotalCost, tx.Reference, tx.Notes, tx.TransactionDate, tx.CreatedAt)
	return err
}

func (r *InventoryRepository) GetTransactionByID(ctx context.Context, companyID, id uuid.UUID) (*entity.InventoryTransaction, error) {
	query := `SELECT id, company_id, transaction_no, type, product_id, warehouse_id, to_warehouse_id, quantity, unit_cost, total_cost, reference, notes, transaction_date, created_at
		FROM inventory_transactions WHERE company_id = $1 AND id = $2`
	var tx entity.InventoryTransaction
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(&tx.ID, &tx.CompanyID, &tx.TransactionNo, &tx.Type, &tx.ProductID, &tx.WarehouseID, &tx.ToWarehouseID, &tx.Quantity, &tx.UnitCost, &tx.TotalCost, &tx.Reference, &tx.Notes, &tx.TransactionDate, &tx.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *InventoryRepository) ListTransactions(ctx context.Context, companyID uuid.UUID, filter repository.InventoryFilter) ([]entity.InventoryTransaction, int64, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if filter.ProductID != nil {
		qb.AddCondition("product_id = $%d", *filter.ProductID)
	}
	if filter.WarehouseID != nil {
		qb.AddCondition("warehouse_id = $%d", *filter.WarehouseID)
	}
	if filter.Type != nil {
		qb.AddCondition("type = $%d", *filter.Type)
	}

	whereClause := qb.WhereClause()

	// Get count
	var total int64
	r.db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM inventory_transactions WHERE %s", whereClause), qb.Args()...).Scan(&total)

	page, pageSize := filter.Page, filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	limitPos, offsetPos := qb.AddLimitOffset(pageSize, (page-1)*pageSize)

	query := fmt.Sprintf(`SELECT id, company_id, transaction_no, type, product_id, warehouse_id, to_warehouse_id, quantity, unit_cost, total_cost, reference, notes, transaction_date, created_at
		FROM inventory_transactions WHERE %s ORDER BY transaction_date DESC LIMIT $%d OFFSET $%d`, whereClause, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txs []entity.InventoryTransaction
	for rows.Next() {
		var tx entity.InventoryTransaction
		if err := rows.Scan(&tx.ID, &tx.CompanyID, &tx.TransactionNo, &tx.Type, &tx.ProductID, &tx.WarehouseID, &tx.ToWarehouseID, &tx.Quantity, &tx.UnitCost, &tx.TotalCost, &tx.Reference, &tx.Notes, &tx.TransactionDate, &tx.CreatedAt); err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}
	return txs, total, nil
}

func (r *InventoryRepository) GetStock(ctx context.Context, companyID, productID, warehouseID uuid.UUID) (*entity.ProductStock, error) {
	query := `SELECT id, company_id, product_id, warehouse_id, quantity, average_cost, updated_at FROM product_stocks WHERE company_id = $1 AND product_id = $2 AND warehouse_id = $3`
	var s entity.ProductStock
	err := r.db.QueryRow(ctx, query, companyID, productID, warehouseID).Scan(&s.ID, &s.CompanyID, &s.ProductID, &s.WarehouseID, &s.Quantity, &s.AverageCost, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *InventoryRepository) GetStockByProduct(ctx context.Context, companyID, productID uuid.UUID) ([]entity.ProductStock, error) {
	rows, err := r.db.Query(ctx, `SELECT id, company_id, product_id, warehouse_id, quantity, average_cost, updated_at FROM product_stocks WHERE company_id = $1 AND product_id = $2`, companyID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []entity.ProductStock
	for rows.Next() {
		var s entity.ProductStock
		if err := rows.Scan(&s.ID, &s.CompanyID, &s.ProductID, &s.WarehouseID, &s.Quantity, &s.AverageCost, &s.UpdatedAt); err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, nil
}

func (r *InventoryRepository) GetStockByWarehouse(ctx context.Context, companyID, warehouseID uuid.UUID) ([]entity.ProductStock, error) {
	rows, err := r.db.Query(ctx, `SELECT id, company_id, product_id, warehouse_id, quantity, average_cost, updated_at FROM product_stocks WHERE company_id = $1 AND warehouse_id = $2`, companyID, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []entity.ProductStock
	for rows.Next() {
		var s entity.ProductStock
		if err := rows.Scan(&s.ID, &s.CompanyID, &s.ProductID, &s.WarehouseID, &s.Quantity, &s.AverageCost, &s.UpdatedAt); err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, nil
}

func (r *InventoryRepository) UpdateStock(ctx context.Context, stock *entity.ProductStock) error {
	stock.UpdatedAt = time.Now()
	query := `INSERT INTO product_stocks (id, company_id, product_id, warehouse_id, quantity, average_cost, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (company_id, product_id, warehouse_id) DO UPDATE SET quantity = $5, average_cost = $6, updated_at = $7`
	if stock.ID == uuid.Nil {
		stock.ID = uuid.New()
	}
	_, err := r.db.Exec(ctx, query, stock.ID, stock.CompanyID, stock.ProductID, stock.WarehouseID, stock.Quantity, stock.AverageCost, stock.UpdatedAt)
	return err
}

func (r *InventoryRepository) GetTotalStock(ctx context.Context, companyID, productID uuid.UUID) (*entity.ProductStock, error) {
	var s entity.ProductStock
	s.CompanyID = companyID
	s.ProductID = productID
	err := r.db.QueryRow(ctx, `SELECT COALESCE(SUM(quantity), 0), COALESCE(AVG(average_cost), 0) FROM product_stocks WHERE company_id = $1 AND product_id = $2`, companyID, productID).Scan(&s.Quantity, &s.AverageCost)
	return &s, err
}

func (r *InventoryRepository) ListStocks(ctx context.Context, companyID uuid.UUID) ([]entity.ProductStock, error) {
	query := `
		SELECT 
			COALESCE(s.id, '00000000-0000-0000-0000-000000000000') as id,
			p.company_id,
			p.id as product_id,
			COALESCE(s.warehouse_id, '00000000-0000-0000-0000-000000000000') as warehouse_id,
			COALESCE(s.quantity, 0) as quantity,
			COALESCE(s.average_cost, 0) as average_cost,
			COALESCE(s.updated_at, p.updated_at) as updated_at,
			p.name as product_name,
			p.code as sku,
			COALESCE(w.name, 'Main Warehouse') as warehouse_name
		FROM products p
		LEFT JOIN product_stocks s ON p.id = s.product_id
		LEFT JOIN warehouses w ON s.warehouse_id = w.id
		WHERE p.company_id = $1 AND p.is_active = true
		ORDER BY p.name`

	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []entity.ProductStock
	for rows.Next() {
		var s entity.ProductStock
		if err := rows.Scan(&s.ID, &s.CompanyID, &s.ProductID, &s.WarehouseID, &s.Quantity, &s.AverageCost, &s.UpdatedAt, &s.ProductName, &s.ProductCode, &s.WarehouseName); err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, nil
}
