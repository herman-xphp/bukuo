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

type SalesInvoiceRepository struct {
	db *pgxpool.Pool
}

func NewSalesInvoiceRepository(db *pgxpool.Pool) *SalesInvoiceRepository {
	return &SalesInvoiceRepository{db: db}
}

func (r *SalesInvoiceRepository) Create(ctx context.Context, inv *entity.SalesInvoice, lines []entity.SalesInvoiceLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO sales_invoices (
				id, company_id, invoice_no, customer_id, order_id, 
				invoice_date, due_date, status, subtotal, tax_amount, 
				discount_amount, total, paid_amount, notes, journal_id, 
				created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`
		if _, err := tx.Exec(ctx, query,
			inv.ID, inv.CompanyID, inv.InvoiceNo, inv.CustomerID, inv.OrderID,
			inv.InvoiceDate, inv.DueDate, inv.Status, inv.Subtotal, inv.TaxAmount,
			inv.DiscountAmount, inv.Total, inv.PaidAmount, inv.Notes, inv.JournalID,
			inv.CreatedAt, inv.UpdatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO sales_invoice_lines (
				id, invoice_id, product_id, description, quantity, 
				unit_price, discount_pct, tax_pct, line_total
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery,
				line.ID, line.InvoiceID, line.ProductID, line.Description, line.Quantity,
				line.UnitPrice, line.DiscountPct, line.TaxPct, line.LineTotal,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SalesInvoiceRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesInvoice, error) {
	query := `
		SELECT id, company_id, invoice_no, customer_id, order_id, 
			   invoice_date, due_date, status, subtotal, tax_amount, 
			   discount_amount, total, paid_amount, notes, journal_id, 
			   created_at, updated_at
		FROM sales_invoices 
		WHERE company_id = $1 AND id = $2
	`
	var inv entity.SalesInvoice
	var notes *string

	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&inv.ID, &inv.CompanyID, &inv.InvoiceNo, &inv.CustomerID, &inv.OrderID,
		&inv.InvoiceDate, &inv.DueDate, &inv.Status, &inv.Subtotal, &inv.TaxAmount,
		&inv.DiscountAmount, &inv.Total, &inv.PaidAmount, &notes, &inv.JournalID,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if notes != nil {
		inv.Notes = *notes
	}

	linesQuery := `
		SELECT id, invoice_id, product_id, description, quantity, 
			   unit_price, discount_pct, tax_pct, line_total
		FROM sales_invoice_lines
		WHERE invoice_id = $1
	`
	rows, err := r.db.Query(ctx, linesQuery, inv.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.SalesInvoiceLine
		var description *string
		if err := rows.Scan(
			&line.ID, &line.InvoiceID, &line.ProductID, &description, &line.Quantity,
			&line.UnitPrice, &line.DiscountPct, &line.TaxPct, &line.LineTotal,
		); err != nil {
			return nil, err
		}
		if description != nil {
			line.Description = *description
		}
		inv.Lines = append(inv.Lines, line)
	}

	return &inv, nil
}

func (r *SalesInvoiceRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.SalesStatus) error {
	query := "UPDATE sales_invoices SET status = $1, updated_at = $2 WHERE company_id = $3 AND id = $4"
	_, err := r.db.Exec(ctx, query, status, time.Now(), companyID, id)
	return err
}

func (r *SalesInvoiceRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.SalesFilter) ([]entity.SalesInvoice, int64, error) {
	qb := querybuilder.New()
	qb.AddCondition("i.company_id = $%d", companyID)

	if filter.CustomerID != nil {
		qb.AddCondition("i.customer_id = $%d", *filter.CustomerID)
	}

	whereClause := qb.WhereClause()

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sales_invoices i WHERE %s", whereClause)
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
		SELECT i.id, i.company_id, i.invoice_no, i.customer_id, i.order_id, 
			   i.invoice_date, i.due_date, i.status, i.subtotal, i.tax_amount, 
			   i.discount_amount, i.total, i.paid_amount, i.notes, i.journal_id, 
			   i.created_at, i.updated_at,
			   c.name as customer_name
		FROM sales_invoices i
		LEFT JOIN contacts c ON i.customer_id = c.id
		WHERE %s 
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []entity.SalesInvoice
	for rows.Next() {
		var inv entity.SalesInvoice
		var customerName *string

		if err := rows.Scan(
			&inv.ID, &inv.CompanyID, &inv.InvoiceNo, &inv.CustomerID, &inv.OrderID,
			&inv.InvoiceDate, &inv.DueDate, &inv.Status, &inv.Subtotal, &inv.TaxAmount,
			&inv.DiscountAmount, &inv.Total, &inv.PaidAmount, &inv.Notes, &inv.JournalID,
			&inv.CreatedAt, &inv.UpdatedAt,
			&customerName,
		); err != nil {
			return nil, 0, err
		}

		if customerName != nil {
			inv.Customer = &entity.Customer{
				Contact: entity.Contact{ID: inv.CustomerID, Name: *customerName},
			}
		}

		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *SalesInvoiceRepository) Update(ctx context.Context, inv *entity.SalesInvoice) error {
	inv.UpdatedAt = time.Now()
	query := `
		UPDATE sales_invoices
		SET invoice_no = $3, customer_id = $4, order_id = $5, 
			invoice_date = $6, due_date = $7, status = $8, subtotal = $9, 
			tax_amount = $10, discount_amount = $11, total = $12, 
			paid_amount = $13, notes = $14, journal_id = $15, updated_at = $16
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query,
		inv.CompanyID, inv.ID, inv.InvoiceNo, inv.CustomerID, inv.OrderID,
		inv.InvoiceDate, inv.DueDate, inv.Status, inv.Subtotal, inv.TaxAmount,
		inv.DiscountAmount, inv.Total, inv.PaidAmount, inv.Notes, inv.JournalID,
		inv.UpdatedAt,
	)
	return err
}

func (r *SalesInvoiceRepository) UpdatePaidAmount(ctx context.Context, companyID, id uuid.UUID, paidAmount interface{}) error {
	return nil
}
