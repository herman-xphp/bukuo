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

var _ repository.SalesQuotationRepository = (*SalesQuotationRepository)(nil)

type SalesQuotationRepository struct {
	db *pgxpool.Pool
}

func NewSalesQuotationRepository(db *pgxpool.Pool) *SalesQuotationRepository {
	return &SalesQuotationRepository{db: db}
}

func (r *SalesQuotationRepository) Create(ctx context.Context, q *entity.SalesQuotation, lines []entity.SalesQuotationLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO sales_quotations (
				id, company_id, quotation_no, customer_id, quotation_date, valid_until,
				status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`
		if _, err := tx.Exec(ctx, query,
			q.ID, q.CompanyID, q.QuotationNo, q.CustomerID, q.QuotationDate, q.ValidUntil,
			q.Status, q.Subtotal, q.TaxAmount, q.DiscountAmount, q.Total, q.Notes, q.CreatedAt, q.UpdatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO sales_quotation_lines (
				id, quotation_id, product_id, description, quantity,
				unit_price, discount_pct, tax_pct, line_total
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery,
				line.ID, line.QuotationID, line.ProductID, line.Description, line.Quantity,
				line.UnitPrice, line.DiscountPct, line.TaxPct, line.LineTotal,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SalesQuotationRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesQuotation, error) {
	query := `
		SELECT id, company_id, quotation_no, customer_id, quotation_date, valid_until,
			   status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
		FROM sales_quotations
		WHERE company_id = $1 AND id = $2
	`
	var q entity.SalesQuotation
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&q.ID, &q.CompanyID, &q.QuotationNo, &q.CustomerID, &q.QuotationDate, &q.ValidUntil,
		&q.Status, &q.Subtotal, &q.TaxAmount, &q.DiscountAmount, &q.Total, &q.Notes, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	linesQuery := `
		SELECT id, quotation_id, product_id, description, quantity,
			   unit_price, discount_pct, tax_pct, line_total
		FROM sales_quotation_lines
		WHERE quotation_id = $1
	`
	rows, err := r.db.Query(ctx, linesQuery, q.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.SalesQuotationLine
		if err := rows.Scan(
			&line.ID, &line.QuotationID, &line.ProductID, &line.Description, &line.Quantity,
			&line.UnitPrice, &line.DiscountPct, &line.TaxPct, &line.LineTotal,
		); err != nil {
			return nil, err
		}
		q.Lines = append(q.Lines, line)
	}

	return &q, nil
}

func (r *SalesQuotationRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.SalesFilter) ([]entity.SalesQuotation, int64, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if filter.CustomerID != nil {
		qb.AddCondition("customer_id = $%d", *filter.CustomerID)
	}
	if filter.Status != nil {
		qb.AddCondition("status = $%d", *filter.Status)
	}

	whereClause := qb.WhereClause()

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sales_quotations WHERE %s", whereClause)
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
		SELECT id, company_id, quotation_no, customer_id, quotation_date, valid_until,
			   status, subtotal, tax_amount, discount_amount, total, notes, created_at, updated_at
		FROM sales_quotations
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var quotations []entity.SalesQuotation
	for rows.Next() {
		var q entity.SalesQuotation
		if err := rows.Scan(
			&q.ID, &q.CompanyID, &q.QuotationNo, &q.CustomerID, &q.QuotationDate, &q.ValidUntil,
			&q.Status, &q.Subtotal, &q.TaxAmount, &q.DiscountAmount, &q.Total, &q.Notes, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		quotations = append(quotations, q)
	}

	return quotations, total, nil
}

func (r *SalesQuotationRepository) Update(ctx context.Context, q *entity.SalesQuotation) error {
	q.UpdatedAt = time.Now()
	query := `
		UPDATE sales_quotations
		SET quotation_no = $3, customer_id = $4, quotation_date = $5, valid_until = $6,
			status = $7, subtotal = $8, tax_amount = $9, discount_amount = $10, 
			total = $11, notes = $12, updated_at = $13
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query,
		q.CompanyID, q.ID, q.QuotationNo, q.CustomerID, q.QuotationDate, q.ValidUntil,
		q.Status, q.Subtotal, q.TaxAmount, q.DiscountAmount, q.Total, q.Notes, q.UpdatedAt,
	)
	return err
}

func (r *SalesQuotationRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.SalesStatus) error {
	query := "UPDATE sales_quotations SET status = $1, updated_at = $2 WHERE company_id = $3 AND id = $4"
	_, err := r.db.Exec(ctx, query, status, time.Now(), companyID, id)
	return err
}

func (r *SalesQuotationRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "DELETE FROM sales_quotation_lines WHERE quotation_id = $1", id); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "DELETE FROM sales_quotations WHERE company_id = $1 AND id = $2", companyID, id)
		return err
	})
}
