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

type PayrollRepository struct {
	db *pgxpool.Pool
}

func NewPayrollRepository(db *pgxpool.Pool) *PayrollRepository {
	return &PayrollRepository{db: db}
}

// -- Employee --

func (r *PayrollRepository) CreateEmployee(ctx context.Context, e *entity.Employee) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO employees (
				id, company_id, first_name, last_name, email, phone, join_date, status,
				job_title, department, basic_salary, bank_name, bank_account, tax_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		`
		if _, err := tx.Exec(ctx, query,
			e.ID, e.CompanyID, e.FirstName, e.LastName, e.Email, e.Phone, e.JoinDate, e.Status,
			e.JobTitle, e.Department, e.BasicSalary, e.BankName, e.BankAccount, e.TaxID, e.CreatedAt, e.UpdatedAt,
		); err != nil {
			return err
		}

		// Insert Components
		compQuery := `
			INSERT INTO employee_salary_components (id, employee_id, component_id, amount)
			VALUES ($1, $2, $3, $4)
		`
		for _, c := range e.Components {
			if _, err := tx.Exec(ctx, compQuery, c.ID, e.ID, c.ComponentID, c.Amount); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PayrollRepository) GetEmployee(ctx context.Context, companyID, id uuid.UUID) (*entity.Employee, error) {
	query := `
		SELECT id, company_id, first_name, last_name, email, phone, join_date, resign_date,
		       status, job_title, department, basic_salary, bank_name, bank_account, tax_id
		FROM employees WHERE company_id = $1 AND id = $2
	`
	var e entity.Employee
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&e.ID, &e.CompanyID, &e.FirstName, &e.LastName, &e.Email, &e.Phone, &e.JoinDate, &e.ResignDate,
		&e.Status, &e.JobTitle, &e.Department, &e.BasicSalary, &e.BankName, &e.BankAccount, &e.TaxID,
	)
	if err != nil {
		return nil, err
	}

	// Get Components
	compQuery := `
		SELECT esc.id, esc.component_id, sc.name, sc.type, esc.amount
		FROM employee_salary_components esc
		JOIN salary_components sc ON sc.id = esc.component_id
		WHERE esc.employee_id = $1
	`
	rows, err := r.db.Query(ctx, compQuery, e.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c entity.EmployeeSalaryComponent
		if err := rows.Scan(&c.ID, &c.ComponentID, &c.Component.Name, &c.Component.Type, &c.Amount); err != nil {
			return nil, err
		}
		c.EmployeeID = e.ID
		e.Components = append(e.Components, c)
	}
	return &e, nil
}

func (r *PayrollRepository) ListEmployees(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.Employee, int64, error) {
	countQuery := "SELECT COUNT(*) FROM employees WHERE company_id = $1"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, first_name, last_name, email, status, job_title, department
		FROM employees WHERE company_id = $1 ORDER BY first_name ASC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var employees []entity.Employee
	for rows.Next() {
		var e entity.Employee
		if err := rows.Scan(&e.ID, &e.FirstName, &e.LastName, &e.Email, &e.Status, &e.JobTitle, &e.Department); err != nil {
			return nil, 0, err
		}
		employees = append(employees, e)
	}
	return employees, total, nil
}

// -- Salary Component --

func (r *PayrollRepository) CreateComponent(ctx context.Context, c *entity.SalaryComponent) error {
	query := `
		INSERT INTO salary_components (id, company_id, name, type, amount, is_taxable, account_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query, c.ID, c.CompanyID, c.Name, c.Type, c.Amount, c.IsTaxable, c.AccountID, c.CreatedAt)
	return err
}

func (r *PayrollRepository) ListComponents(ctx context.Context, companyID uuid.UUID) ([]entity.SalaryComponent, error) {
	query := `
		SELECT id, name, type, amount, is_taxable, account_id
		FROM salary_components WHERE company_id = $1 ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var components []entity.SalaryComponent
	for rows.Next() {
		var c entity.SalaryComponent
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Amount, &c.IsTaxable, &c.AccountID); err != nil {
			return nil, err
		}
		components = append(components, c)
	}
	return components, nil
}

// -- Pay Run --

func (r *PayrollRepository) CreatePayRun(ctx context.Context, pr *entity.PayRun) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO pay_runs (
				id, company_id, period_id, start_date, end_date, payment_date,
				total_gross, total_net, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		if _, err := tx.Exec(ctx, query,
			pr.ID, pr.CompanyID, pr.PeriodID, pr.StartDate, pr.EndDate, pr.PaymentDate,
			pr.TotalGross, pr.TotalNet, pr.Status, pr.CreatedAt, pr.UpdatedAt,
		); err != nil {
			return err
		}

		for _, slip := range pr.Slips {
			slipQuery := `
				INSERT INTO pay_slips (id, pay_run_id, employee_id, basic_salary, gross_pay, deductions, net_pay)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`
			if _, err := tx.Exec(ctx, slipQuery, slip.ID, pr.ID, slip.EmployeeID, slip.BasicSalary, slip.GrossPay, slip.Deductions, slip.NetPay); err != nil {
				return err
			}

			for _, item := range slip.Items {
				itemQuery := `
					INSERT INTO pay_slip_items (id, pay_slip_id, component_id, name, type, amount)
					VALUES ($1, $2, $3, $4, $5, $6)
				`
				// Handle nullable UUID for component_id
				var compID interface{}
				if item.ComponentID != uuid.Nil {
					compID = item.ComponentID
				}
				if _, err := tx.Exec(ctx, itemQuery, item.ID, slip.ID, compID, item.Name, item.Type, item.Amount); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *PayrollRepository) GetPayRun(ctx context.Context, companyID, id uuid.UUID) (*entity.PayRun, error) {
	query := `
		SELECT id, company_id, period_id, start_date, end_date, payment_date,
		       total_gross, total_net, status, journal_id, created_at
		FROM pay_runs WHERE company_id = $1 AND id = $2
	`
	var pr entity.PayRun
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&pr.ID, &pr.CompanyID, &pr.PeriodID, &pr.StartDate, &pr.EndDate, &pr.PaymentDate,
		&pr.TotalGross, &pr.TotalNet, &pr.Status, &pr.JournalID, &pr.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PayrollRepository) UpdatePayRunStatus(ctx context.Context, companyID, id uuid.UUID, status entity.PayRunStatus, journalID *uuid.UUID) error {
	query := "UPDATE pay_runs SET status = $1, journal_id = $2, updated_at = $3 WHERE company_id = $4 AND id = $5"
	_, err := r.db.Exec(ctx, query, status, journalID, time.Now(), companyID, id)
	return err
}

func (r *PayrollRepository) GetPaySlips(ctx context.Context, payRunID uuid.UUID) ([]entity.PaySlip, error) {
	query := `
		SELECT ps.id, ps.employee_id, e.first_name || ' ' || e.last_name as name,
		       ps.basic_salary, ps.gross_pay, ps.deductions, ps.net_pay
		FROM pay_slips ps
		JOIN employees e ON e.id = ps.employee_id
		WHERE ps.pay_run_id = $1
	`
	rows, err := r.db.Query(ctx, query, payRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slips []entity.PaySlip
	for rows.Next() {
		var ps entity.PaySlip
		if err := rows.Scan(&ps.ID, &ps.EmployeeID, &ps.EmployeeName, &ps.BasicSalary, &ps.GrossPay, &ps.Deductions, &ps.NetPay); err != nil {
			return nil, err
		}
		slips = append(slips, ps)
	}
	return slips, nil
}
