package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type ARAPRepository struct {
	db *pgxpool.Pool
}

func NewARAPRepository(db *pgxpool.Pool) *ARAPRepository {
	return &ARAPRepository{db: db}
}

// GetARAgingReport generates accounts receivable aging report
func (r *ARAPRepository) GetARAgingReport(ctx context.Context, companyID uuid.UUID) (*entity.ARAgingReport, error) {
	now := time.Now()
	report := &entity.ARAgingReport{
		CompanyID:     companyID,
		ReportDate:    now,
		Current:       entity.ARAgingBucket{Label: "0-30 days"},
		ThirtyDays:    entity.ARAgingBucket{Label: "31-60 days"},
		SixtyDays:     entity.ARAgingBucket{Label: "61-90 days"},
		NinetyDays:    entity.ARAgingBucket{Label: "91-120 days"},
		OverOneTwenty: entity.ARAgingBucket{Label: ">120 days"},
	}

	// Query outstanding sales invoices
	query := `
		SELECT si.id, si.invoice_no, si.contact_id, c.name, si.invoice_date, si.due_date, 
		       si.total, si.paid_amount, (si.total - si.paid_amount) as balance_due
		FROM sales_invoices si
		JOIN contacts c ON c.id = si.contact_id
		WHERE si.company_id = $1 AND si.status != 'PAID' AND si.status != 'CANCELLED'
		  AND si.total > si.paid_amount
		ORDER BY si.due_date
	`

	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customerMap := make(map[uuid.UUID]*entity.CustomerAging)

	for rows.Next() {
		var inv entity.OutstandingInvoice
		if err := rows.Scan(
			&inv.InvoiceID, &inv.InvoiceNo, &inv.ContactID, &inv.ContactName,
			&inv.InvoiceDate, &inv.DueDate, &inv.TotalAmount, &inv.PaidAmount, &inv.BalanceDue,
		); err != nil {
			return nil, err
		}

		daysOverdue := int(now.Sub(inv.DueDate).Hours() / 24)
		if daysOverdue < 0 {
			daysOverdue = 0
		}

		// Get or create customer aging
		ca, exists := customerMap[inv.ContactID]
		if !exists {
			ca = &entity.CustomerAging{
				ContactID:   inv.ContactID,
				ContactName: inv.ContactName,
			}
			customerMap[inv.ContactID] = ca
		}
		ca.TotalDue = ca.TotalDue.Add(inv.BalanceDue)

		// Categorize by aging bucket
		switch {
		case daysOverdue <= 30:
			report.Current.Amount = report.Current.Amount.Add(inv.BalanceDue)
			report.Current.InvoiceCount++
			ca.Current = ca.Current.Add(inv.BalanceDue)
		case daysOverdue <= 60:
			report.ThirtyDays.Amount = report.ThirtyDays.Amount.Add(inv.BalanceDue)
			report.ThirtyDays.InvoiceCount++
			ca.ThirtyDays = ca.ThirtyDays.Add(inv.BalanceDue)
		case daysOverdue <= 90:
			report.SixtyDays.Amount = report.SixtyDays.Amount.Add(inv.BalanceDue)
			report.SixtyDays.InvoiceCount++
			ca.SixtyDays = ca.SixtyDays.Add(inv.BalanceDue)
		case daysOverdue <= 120:
			report.NinetyDays.Amount = report.NinetyDays.Amount.Add(inv.BalanceDue)
			report.NinetyDays.InvoiceCount++
			ca.NinetyDays = ca.NinetyDays.Add(inv.BalanceDue)
		default:
			report.OverOneTwenty.Amount = report.OverOneTwenty.Amount.Add(inv.BalanceDue)
			report.OverOneTwenty.InvoiceCount++
			ca.OverOneTwenty = ca.OverOneTwenty.Add(inv.BalanceDue)
		}
	}

	report.TotalAR = report.Current.Amount.Add(report.ThirtyDays.Amount).Add(report.SixtyDays.Amount).Add(report.NinetyDays.Amount).Add(report.OverOneTwenty.Amount)

	for _, ca := range customerMap {
		report.ByCustomer = append(report.ByCustomer, *ca)
	}

	return report, nil
}

// GetAPAgingReport generates accounts payable aging report
func (r *ARAPRepository) GetAPAgingReport(ctx context.Context, companyID uuid.UUID) (*entity.APAgingReport, error) {
	now := time.Now()
	report := &entity.APAgingReport{
		CompanyID:     companyID,
		ReportDate:    now,
		Current:       entity.ARAgingBucket{Label: "0-30 days"},
		ThirtyDays:    entity.ARAgingBucket{Label: "31-60 days"},
		SixtyDays:     entity.ARAgingBucket{Label: "61-90 days"},
		NinetyDays:    entity.ARAgingBucket{Label: "91-120 days"},
		OverOneTwenty: entity.ARAgingBucket{Label: ">120 days"},
	}

	query := `
		SELECT pi.id, pi.invoice_no, pi.supplier_id, c.name, pi.invoice_date, pi.due_date,
		       pi.total, pi.paid_amount, (pi.total - pi.paid_amount) as balance_due
		FROM purchase_invoices pi
		JOIN contacts c ON c.id = pi.supplier_id
		WHERE pi.company_id = $1 AND pi.status != 'PAID' AND pi.status != 'CANCELLED'
		  AND pi.total > pi.paid_amount
		ORDER BY pi.due_date
	`

	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	supplierMap := make(map[uuid.UUID]*entity.SupplierAging)

	for rows.Next() {
		var inv entity.OutstandingInvoice
		if err := rows.Scan(
			&inv.InvoiceID, &inv.InvoiceNo, &inv.ContactID, &inv.ContactName,
			&inv.InvoiceDate, &inv.DueDate, &inv.TotalAmount, &inv.PaidAmount, &inv.BalanceDue,
		); err != nil {
			return nil, err
		}

		daysOverdue := int(now.Sub(inv.DueDate).Hours() / 24)
		if daysOverdue < 0 {
			daysOverdue = 0
		}

		sa, exists := supplierMap[inv.ContactID]
		if !exists {
			sa = &entity.SupplierAging{
				ContactID:   inv.ContactID,
				ContactName: inv.ContactName,
			}
			supplierMap[inv.ContactID] = sa
		}
		sa.TotalDue = sa.TotalDue.Add(inv.BalanceDue)

		switch {
		case daysOverdue <= 30:
			report.Current.Amount = report.Current.Amount.Add(inv.BalanceDue)
			report.Current.InvoiceCount++
			sa.Current = sa.Current.Add(inv.BalanceDue)
		case daysOverdue <= 60:
			report.ThirtyDays.Amount = report.ThirtyDays.Amount.Add(inv.BalanceDue)
			report.ThirtyDays.InvoiceCount++
			sa.ThirtyDays = sa.ThirtyDays.Add(inv.BalanceDue)
		case daysOverdue <= 90:
			report.SixtyDays.Amount = report.SixtyDays.Amount.Add(inv.BalanceDue)
			report.SixtyDays.InvoiceCount++
			sa.SixtyDays = sa.SixtyDays.Add(inv.BalanceDue)
		case daysOverdue <= 120:
			report.NinetyDays.Amount = report.NinetyDays.Amount.Add(inv.BalanceDue)
			report.NinetyDays.InvoiceCount++
			sa.NinetyDays = sa.NinetyDays.Add(inv.BalanceDue)
		default:
			report.OverOneTwenty.Amount = report.OverOneTwenty.Amount.Add(inv.BalanceDue)
			report.OverOneTwenty.InvoiceCount++
			sa.OverOneTwenty = sa.OverOneTwenty.Add(inv.BalanceDue)
		}
	}

	report.TotalAP = report.Current.Amount.Add(report.ThirtyDays.Amount).Add(report.SixtyDays.Amount).Add(report.NinetyDays.Amount).Add(report.OverOneTwenty.Amount)

	for _, sa := range supplierMap {
		report.BySupplier = append(report.BySupplier, *sa)
	}

	return report, nil
}

// GetOutstandingReceivables gets list of unpaid sales invoices
func (r *ARAPRepository) GetOutstandingReceivables(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.OutstandingInvoice, int64, decimal.Decimal, error) {
	var total int64
	var totalAmount decimal.Decimal

	countQuery := `
		SELECT COUNT(*), COALESCE(SUM(total - paid_amount), 0)
		FROM sales_invoices
		WHERE company_id = $1 AND status != 'PAID' AND status != 'CANCELLED' AND total > paid_amount
	`
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total, &totalAmount); err != nil {
		return nil, 0, decimal.Zero, err
	}

	query := `
		SELECT si.id, si.invoice_no, si.contact_id, c.name, si.invoice_date, si.due_date,
		       si.total, si.paid_amount, (si.total - si.paid_amount) as balance_due
		FROM sales_invoices si
		JOIN contacts c ON c.id = si.contact_id
		WHERE si.company_id = $1 AND si.status != 'PAID' AND si.status != 'CANCELLED' AND si.total > si.paid_amount
		ORDER BY si.due_date
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, decimal.Zero, err
	}
	defer rows.Close()

	now := time.Now()
	var invoices []entity.OutstandingInvoice
	for rows.Next() {
		var inv entity.OutstandingInvoice
		if err := rows.Scan(
			&inv.InvoiceID, &inv.InvoiceNo, &inv.ContactID, &inv.ContactName,
			&inv.InvoiceDate, &inv.DueDate, &inv.TotalAmount, &inv.PaidAmount, &inv.BalanceDue,
		); err != nil {
			return nil, 0, decimal.Zero, err
		}
		inv.InvoiceType = "SALES"
		inv.DaysOverdue = int(now.Sub(inv.DueDate).Hours() / 24)
		if inv.DaysOverdue < 0 {
			inv.DaysOverdue = 0
		}
		invoices = append(invoices, inv)
	}

	return invoices, total, totalAmount, nil
}

// GetOutstandingPayables gets list of unpaid purchase invoices
func (r *ARAPRepository) GetOutstandingPayables(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.OutstandingInvoice, int64, decimal.Decimal, error) {
	var total int64
	var totalAmount decimal.Decimal

	countQuery := `
		SELECT COUNT(*), COALESCE(SUM(total - paid_amount), 0)
		FROM purchase_invoices
		WHERE company_id = $1 AND status != 'PAID' AND status != 'CANCELLED' AND total > paid_amount
	`
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total, &totalAmount); err != nil {
		return nil, 0, decimal.Zero, err
	}

	query := `
		SELECT pi.id, pi.invoice_no, pi.supplier_id, c.name, pi.invoice_date, pi.due_date,
		       pi.total, pi.paid_amount, (pi.total - pi.paid_amount) as balance_due
		FROM purchase_invoices pi
		JOIN contacts c ON c.id = pi.supplier_id
		WHERE pi.company_id = $1 AND pi.status != 'PAID' AND pi.status != 'CANCELLED' AND pi.total > pi.paid_amount
		ORDER BY pi.due_date
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, decimal.Zero, err
	}
	defer rows.Close()

	now := time.Now()
	var invoices []entity.OutstandingInvoice
	for rows.Next() {
		var inv entity.OutstandingInvoice
		if err := rows.Scan(
			&inv.InvoiceID, &inv.InvoiceNo, &inv.ContactID, &inv.ContactName,
			&inv.InvoiceDate, &inv.DueDate, &inv.TotalAmount, &inv.PaidAmount, &inv.BalanceDue,
		); err != nil {
			return nil, 0, decimal.Zero, err
		}
		inv.InvoiceType = "PURCHASE"
		inv.DaysOverdue = int(now.Sub(inv.DueDate).Hours() / 24)
		if inv.DaysOverdue < 0 {
			inv.DaysOverdue = 0
		}
		invoices = append(invoices, inv)
	}

	return invoices, total, totalAmount, nil
}
