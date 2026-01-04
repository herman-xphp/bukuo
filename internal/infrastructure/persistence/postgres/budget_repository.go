package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/txhelper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type BudgetRepository struct {
	db *pgxpool.Pool
}

func NewBudgetRepository(db *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(ctx context.Context, budget *entity.Budget, lines []entity.BudgetLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO budgets (
				id, company_id, name, description, period_id, start_date, end_date,
				total_amount, status, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		if _, err := tx.Exec(ctx, query,
			budget.ID, budget.CompanyID, budget.Name, budget.Description, budget.PeriodID,
			budget.StartDate, budget.EndDate, budget.TotalAmount, budget.Status, budget.CreatedAt, budget.UpdatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO budget_lines (id, budget_id, account_id, budget_amount, notes)
			VALUES ($1, $2, $3, $4, $5)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery, line.ID, line.BudgetID, line.AccountID, line.BudgetAmount, line.Notes); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *BudgetRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Budget, error) {
	query := `
		SELECT id, company_id, name, description, period_id, start_date, end_date,
			   total_amount, status, created_at, updated_at
		FROM budgets
		WHERE company_id = $1 AND id = $2
	`
	var b entity.Budget
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&b.ID, &b.CompanyID, &b.Name, &b.Description, &b.PeriodID, &b.StartDate, &b.EndDate,
		&b.TotalAmount, &b.Status, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	linesQuery := `
		SELECT bl.id, bl.budget_id, bl.account_id, a.code, a.name, bl.budget_amount, bl.notes
		FROM budget_lines bl
		JOIN accounts a ON a.id = bl.account_id
		WHERE bl.budget_id = $1
	`
	rows, err := r.db.Query(ctx, linesQuery, b.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.BudgetLine
		if err := rows.Scan(&line.ID, &line.BudgetID, &line.AccountID, &line.AccountCode, &line.AccountName, &line.BudgetAmount, &line.Notes); err != nil {
			return nil, err
		}
		b.Lines = append(b.Lines, line)
	}
	return &b, nil
}

func (r *BudgetRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.Budget, int64, error) {
	countQuery := "SELECT COUNT(*) FROM budgets WHERE company_id = $1"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, company_id, name, description, period_id, start_date, end_date,
			   total_amount, status, created_at, updated_at
		FROM budgets
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var budgets []entity.Budget
	for rows.Next() {
		var b entity.Budget
		if err := rows.Scan(&b.ID, &b.CompanyID, &b.Name, &b.Description, &b.PeriodID, &b.StartDate, &b.EndDate, &b.TotalAmount, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, err
		}
		budgets = append(budgets, b)
	}
	return budgets, total, nil
}

func (r *BudgetRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.BudgetStatus) error {
	query := "UPDATE budgets SET status = $1, updated_at = $2 WHERE company_id = $3 AND id = $4"
	_, err := r.db.Exec(ctx, query, status, time.Now(), companyID, id)
	return err
}

// GetBudgetVsActual calculates variance between budget and actual spending
func (r *BudgetRepository) GetBudgetVsActual(ctx context.Context, companyID, budgetID uuid.UUID) (*entity.BudgetVsActual, error) {
	budgetQuery := `SELECT id, name, period_id FROM budgets WHERE company_id = $1 AND id = $2`
	var bva entity.BudgetVsActual
	var periodID uuid.UUID
	if err := r.db.QueryRow(ctx, budgetQuery, companyID, budgetID).Scan(&bva.BudgetID, &bva.BudgetName, &periodID); err != nil {
		return nil, err
	}

	// Get period name
	periodQuery := "SELECT name FROM periods WHERE id = $1"
	_ = r.db.QueryRow(ctx, periodQuery, periodID).Scan(&bva.PeriodName)

	// Get budget lines with actuals from journal entries
	linesQuery := `
		SELECT bl.id, bl.budget_id, bl.account_id, a.code, a.name, bl.budget_amount,
			   COALESCE(
				   (SELECT SUM(CASE WHEN a.normal_balance = 'DEBIT' THEN jl.debit - jl.credit ELSE jl.credit - jl.debit END)
				    FROM journal_lines jl
				    JOIN journals j ON j.id = jl.journal_id
				    WHERE jl.account_id = bl.account_id AND j.company_id = $1 AND j.period_id = $2 AND j.status = 'POSTED'),
			   0) as actual_amount,
			   bl.notes, bl.budget_amount
		FROM budget_lines bl
		JOIN accounts a ON a.id = bl.account_id
		WHERE bl.budget_id = $3
	`
	rows, err := r.db.Query(ctx, linesQuery, companyID, periodID, budgetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.BudgetLine
		var actual decimal.Decimal
		if err := rows.Scan(&line.ID, &line.BudgetID, &line.AccountID, &line.AccountCode, &line.AccountName, &line.BudgetAmount, &actual, &line.Notes, &line.BudgetAmount); err != nil {
			return nil, err
		}
		line.ActualAmount = actual
		line.Variance = line.BudgetAmount.Sub(actual)
		if !line.BudgetAmount.IsZero() {
			line.VariancePct = line.Variance.Div(line.BudgetAmount).Mul(decimal.NewFromInt(100))
		}

		bva.TotalBudget = bva.TotalBudget.Add(line.BudgetAmount)
		bva.TotalActual = bva.TotalActual.Add(line.ActualAmount)
		bva.Lines = append(bva.Lines, line)
	}

	bva.TotalVariance = bva.TotalBudget.Sub(bva.TotalActual)
	return &bva, nil
}
