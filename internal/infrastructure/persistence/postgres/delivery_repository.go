package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/querybuilder"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/txhelper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.DeliveryOrderRepository = (*DeliveryOrderRepository)(nil)

type DeliveryOrderRepository struct {
	db *pgxpool.Pool
}

func NewDeliveryOrderRepository(db *pgxpool.Pool) *DeliveryOrderRepository {
	return &DeliveryOrderRepository{db: db}
}

func (r *DeliveryOrderRepository) Create(ctx context.Context, d *entity.DeliveryOrder, lines []entity.DeliveryOrderLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO delivery_orders (
				id, company_id, delivery_no, customer_id, order_id, warehouse_id,
				delivery_date, status, notes, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		if _, err := tx.Exec(ctx, query,
			d.ID, d.CompanyID, d.DeliveryNo, d.CustomerID, d.OrderID, d.WarehouseID,
			d.DeliveryDate, d.Status, d.Notes, d.CreatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO delivery_order_lines (
				id, delivery_id, order_line_id, product_id, quantity
			)
			VALUES ($1, $2, $3, $4, $5)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery,
				line.ID, line.DeliveryID, line.OrderLineID, line.ProductID, line.Quantity,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *DeliveryOrderRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.DeliveryOrder, error) {
	query := `
		SELECT id, company_id, delivery_no, customer_id, order_id, warehouse_id,
			   delivery_date, status, notes, created_at
		FROM delivery_orders
		WHERE company_id = $1 AND id = $2
	`
	var d entity.DeliveryOrder
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&d.ID, &d.CompanyID, &d.DeliveryNo, &d.CustomerID, &d.OrderID, &d.WarehouseID,
		&d.DeliveryDate, &d.Status, &d.Notes, &d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	linesQuery := `
		SELECT id, delivery_id, order_line_id, product_id, quantity
		FROM delivery_order_lines
		WHERE delivery_id = $1
	`
	rows, err := r.db.Query(ctx, linesQuery, d.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.DeliveryOrderLine
		if err := rows.Scan(
			&line.ID, &line.DeliveryID, &line.OrderLineID, &line.ProductID, &line.Quantity,
		); err != nil {
			return nil, err
		}
		d.Lines = append(d.Lines, line)
	}

	return &d, nil
}

func (r *DeliveryOrderRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.SalesFilter) ([]entity.DeliveryOrder, int64, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if filter.CustomerID != nil {
		qb.AddCondition("customer_id = $%d", *filter.CustomerID)
	}
	if filter.Status != nil {
		qb.AddCondition("status = $%d", *filter.Status)
	}

	whereClause := qb.WhereClause()

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM delivery_orders WHERE %s", whereClause)
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
		SELECT id, company_id, delivery_no, customer_id, order_id, warehouse_id,
			   delivery_date, status, notes, created_at
		FROM delivery_orders
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deliveries []entity.DeliveryOrder
	for rows.Next() {
		var d entity.DeliveryOrder
		if err := rows.Scan(
			&d.ID, &d.CompanyID, &d.DeliveryNo, &d.CustomerID, &d.OrderID, &d.WarehouseID,
			&d.DeliveryDate, &d.Status, &d.Notes, &d.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		deliveries = append(deliveries, d)
	}

	return deliveries, total, nil
}
