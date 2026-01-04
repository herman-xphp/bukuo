package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/txhelper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PurchaseOrderRepository struct {
	db *pgxpool.Pool
}

func NewPurchaseOrderRepository(db *pgxpool.Pool) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

func (r *PurchaseOrderRepository) Create(ctx context.Context, order *entity.PurchaseOrder, lines []entity.PurchaseOrderLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO purchase_orders (
				id, company_id, order_no, supplier_id, request_id, order_date, delivery_date,
				status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`
		if _, err := tx.Exec(ctx, query,
			order.ID, order.CompanyID, order.OrderNo, order.SupplierID, order.RequestID, order.OrderDate, order.DeliveryDate,
			order.Status, order.Subtotal, order.TaxAmount, order.DiscountAmount, order.Total, order.Notes, order.CreatedAt, order.UpdatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO purchase_order_lines (
				id, order_id, product_id, description, quantity, received_qty, unit_price, discount_pct, tax_pct, line_total
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery,
				line.ID, line.OrderID, line.ProductID, line.Description, line.Quantity, line.ReceivedQty, line.UnitPrice, line.DiscountPct, line.TaxPct, line.LineTotal,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PurchaseOrderRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.PurchaseOrder, error) {
	query := `
		SELECT id, company_id, order_no, supplier_id, request_id, order_date, delivery_date,
			   status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
		FROM purchase_orders
		WHERE company_id = $1 AND id = $2
	`
	var o entity.PurchaseOrder
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&o.ID, &o.CompanyID, &o.OrderNo, &o.SupplierID, &o.RequestID, &o.OrderDate, &o.DeliveryDate,
		&o.Status, &o.Subtotal, &o.TaxAmount, &o.DiscountAmount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	linesQuery := `
		SELECT id, order_id, product_id, description, quantity, received_qty, unit_price, discount_pct, tax_pct, line_total
		FROM purchase_order_lines WHERE order_id = $1
	`
	rows, err := r.db.Query(ctx, linesQuery, o.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.PurchaseOrderLine
		if err := rows.Scan(&line.ID, &line.OrderID, &line.ProductID, &line.Description, &line.Quantity, &line.ReceivedQty, &line.UnitPrice, &line.DiscountPct, &line.TaxPct, &line.LineTotal); err != nil {
			return nil, err
		}
		o.Lines = append(o.Lines, line)
	}
	return &o, nil
}

func (r *PurchaseOrderRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.PurchaseOrder, int64, error) {
	countQuery := "SELECT COUNT(*) FROM purchase_orders WHERE company_id = $1"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, company_id, order_no, supplier_id, request_id, order_date, delivery_date,
			   status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
		FROM purchase_orders
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []entity.PurchaseOrder
	for rows.Next() {
		var o entity.PurchaseOrder
		if err := rows.Scan(
			&o.ID, &o.CompanyID, &o.OrderNo, &o.SupplierID, &o.RequestID, &o.OrderDate, &o.DeliveryDate,
			&o.Status, &o.Subtotal, &o.TaxAmount, &o.DiscountAmount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *PurchaseOrderRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.PurchaseStatus) error {
	query := "UPDATE purchase_orders SET status = $1, updated_at = $2 WHERE company_id = $3 AND id = $4"
	_, err := r.db.Exec(ctx, query, status, time.Now(), companyID, id)
	return err
}

// PurchaseInvoiceRepository
type PurchaseInvoiceRepository struct {
	db *pgxpool.Pool
}

func NewPurchaseInvoiceRepository(db *pgxpool.Pool) *PurchaseInvoiceRepository {
	return &PurchaseInvoiceRepository{db: db}
}

func (r *PurchaseInvoiceRepository) Create(ctx context.Context, inv *entity.PurchaseInvoice, lines []entity.PurchaseInvoiceLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO purchase_invoices (
				id, company_id, invoice_no, supplier_id, order_id, invoice_date, due_date,
				status, subtotal, tax_amount, discount_amount, total, paid_amount, notes, journal_id, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`
		if _, err := tx.Exec(ctx, query,
			inv.ID, inv.CompanyID, inv.InvoiceNo, inv.SupplierID, inv.OrderID, inv.InvoiceDate, inv.DueDate,
			inv.Status, inv.Subtotal, inv.TaxAmount, inv.DiscountAmount, inv.Total, inv.PaidAmount, inv.Notes, inv.JournalID, inv.CreatedAt, inv.UpdatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO purchase_invoice_lines (
				id, invoice_id, product_id, description, quantity, unit_price, discount_pct, tax_pct, line_total
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery,
				line.ID, line.InvoiceID, line.ProductID, line.Description, line.Quantity, line.UnitPrice, line.DiscountPct, line.TaxPct, line.LineTotal,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PurchaseInvoiceRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.PurchaseInvoice, error) {
	query := `
		SELECT id, company_id, invoice_no, supplier_id, order_id, invoice_date, due_date,
			   status, subtotal, tax_amount, discount_amount, total, paid_amount, notes, journal_id, created_at, updated_at
		FROM purchase_invoices
		WHERE company_id = $1 AND id = $2
	`
	var inv entity.PurchaseInvoice
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&inv.ID, &inv.CompanyID, &inv.InvoiceNo, &inv.SupplierID, &inv.OrderID, &inv.InvoiceDate, &inv.DueDate,
		&inv.Status, &inv.Subtotal, &inv.TaxAmount, &inv.DiscountAmount, &inv.Total, &inv.PaidAmount, &inv.Notes, &inv.JournalID, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *PurchaseInvoiceRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.PurchaseInvoice, int64, error) {
	countQuery := "SELECT COUNT(*) FROM purchase_invoices WHERE company_id = $1"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, company_id, invoice_no, supplier_id, order_id, invoice_date, due_date,
			   status, subtotal, tax_amount, discount_amount, total, paid_amount, notes, journal_id, created_at, updated_at
		FROM purchase_invoices
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []entity.PurchaseInvoice
	for rows.Next() {
		var inv entity.PurchaseInvoice
		if err := rows.Scan(
			&inv.ID, &inv.CompanyID, &inv.InvoiceNo, &inv.SupplierID, &inv.OrderID, &inv.InvoiceDate, &inv.DueDate,
			&inv.Status, &inv.Subtotal, &inv.TaxAmount, &inv.DiscountAmount, &inv.Total, &inv.PaidAmount, &inv.Notes, &inv.JournalID, &inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, total, nil
}

func (r *PurchaseInvoiceRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.PurchaseStatus) error {
	query := "UPDATE purchase_invoices SET status = $1, updated_at = $2 WHERE company_id = $3 AND id = $4"
	_, err := r.db.Exec(ctx, query, status, time.Now(), companyID, id)
	return err
}
