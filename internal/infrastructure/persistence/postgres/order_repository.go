package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/querybuilder"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/txhelper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.SalesOrderRepository = (*SalesOrderRepository)(nil)

type SalesOrderRepository struct {
	db *pgxpool.Pool
}

func NewSalesOrderRepository(db *pgxpool.Pool) *SalesOrderRepository {
	return &SalesOrderRepository{db: db}
}

func (r *SalesOrderRepository) Create(ctx context.Context, o *entity.SalesOrder, lines []entity.SalesOrderLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO sales_orders (
				id, company_id, order_no, customer_id, quotation_id, order_date, delivery_date,
				status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`
		if _, err := tx.Exec(ctx, query,
			o.ID, o.CompanyID, o.OrderNo, o.CustomerID, o.QuotationID, o.OrderDate, o.DeliveryDate,
			o.Status, o.Subtotal, o.TaxAmount, o.DiscountAmount, o.Total, o.Notes, o.CreatedAt, o.UpdatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO sales_order_lines (
				id, order_id, product_id, description, quantity, delivered_qty,
				unit_price, discount_pct, tax_pct, line_total
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery,
				line.ID, line.OrderID, line.ProductID, line.Description, line.Quantity, line.DeliveredQty,
				line.UnitPrice, line.DiscountPct, line.TaxPct, line.LineTotal,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SalesOrderRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesOrder, error) {
	query := `
		SELECT id, company_id, order_no, customer_id, quotation_id, order_date, delivery_date,
			   status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
		FROM sales_orders
		WHERE company_id = $1 AND id = $2
	`
	var o entity.SalesOrder
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&o.ID, &o.CompanyID, &o.OrderNo, &o.CustomerID, &o.QuotationID, &o.OrderDate, &o.DeliveryDate,
		&o.Status, &o.Subtotal, &o.TaxAmount, &o.DiscountAmount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	linesQuery := `
		SELECT id, order_id, product_id, description, quantity, delivered_qty,
			   unit_price, discount_pct, tax_pct, line_total
		FROM sales_order_lines
		WHERE order_id = $1
	`
	rows, err := r.db.Query(ctx, linesQuery, o.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.SalesOrderLine
		if err := rows.Scan(
			&line.ID, &line.OrderID, &line.ProductID, &line.Description, &line.Quantity, &line.DeliveredQty,
			&line.UnitPrice, &line.DiscountPct, &line.TaxPct, &line.LineTotal,
		); err != nil {
			return nil, err
		}
		o.Lines = append(o.Lines, line)
	}

	return &o, nil
}

func (r *SalesOrderRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.SalesFilter) ([]entity.SalesOrder, int64, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if filter.CustomerID != nil {
		qb.AddCondition("customer_id = $%d", *filter.CustomerID)
	}
	if filter.Status != nil {
		qb.AddCondition("status = $%d", *filter.Status)
	}

	whereClause := qb.WhereClause()

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sales_orders WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, qb.Args()...).Scan(&total); err != nil {
		return nil, 0, err
	}

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
		SELECT id, company_id, order_no, customer_id, quotation_id, order_date, delivery_date,
			   status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
		FROM sales_orders
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []entity.SalesOrder
	for rows.Next() {
		var o entity.SalesOrder
		if err := rows.Scan(
			&o.ID, &o.CompanyID, &o.OrderNo, &o.CustomerID, &o.QuotationID, &o.OrderDate, &o.DeliveryDate,
			&o.Status, &o.Subtotal, &o.TaxAmount, &o.DiscountAmount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

func (r *SalesOrderRepository) Update(ctx context.Context, o *entity.SalesOrder) error {
	o.UpdatedAt = time.Now()
	query := `
		UPDATE sales_orders
		SET order_no = $3, customer_id = $4, quotation_id = $5, order_date = $6, delivery_date = $7,
			status = $8, subtotal = $9, tax_amount = $10, discount_amount = $11, 
			total = $12, notes = $13, updated_at = $14
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query,
		o.CompanyID, o.ID, o.OrderNo, o.CustomerID, o.QuotationID, o.OrderDate, o.DeliveryDate,
		o.Status, o.Subtotal, o.TaxAmount, o.DiscountAmount, o.Total, o.Notes, o.UpdatedAt,
	)
	return err
}

func (r *SalesOrderRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.SalesStatus) error {
	query := "UPDATE sales_orders SET status = $1, updated_at = $2 WHERE company_id = $3 AND id = $4"
	_, err := r.db.Exec(ctx, query, status, time.Now(), companyID, id)
	return err
}
