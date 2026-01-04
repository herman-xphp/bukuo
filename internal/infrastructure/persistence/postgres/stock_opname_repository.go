package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/txhelper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockOpnameRepository struct {
	db *pgxpool.Pool
}

func NewStockOpnameRepository(db *pgxpool.Pool) *StockOpnameRepository {
	return &StockOpnameRepository{db: db}
}

func (r *StockOpnameRepository) Create(ctx context.Context, opname *entity.StockOpname, lines []entity.StockOpnameLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO stock_opnames (
				id, company_id, opname_no, warehouse_id, opname_date,
				status, notes, created_by, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		if _, err := tx.Exec(ctx, query,
			opname.ID, opname.CompanyID, opname.OpnameNo, opname.WarehouseID, opname.OpnameDate,
			opname.Status, opname.Notes, opname.CreatedBy, opname.CreatedAt, opname.UpdatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO stock_opname_lines (
				id, opname_id, product_id, system_qty, actual_qty, difference_qty, notes
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery,
				line.ID, line.OpnameID, line.ProductID, line.SystemQty, line.ActualQty, line.DifferenceQty, line.Notes,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *StockOpnameRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.StockOpname, error) {
	query := `
		SELECT id, company_id, opname_no, warehouse_id, opname_date,
			   status, notes, created_by, approved_by, approved_at, created_at, updated_at
		FROM stock_opnames
		WHERE company_id = $1 AND id = $2
	`
	var o entity.StockOpname
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&o.ID, &o.CompanyID, &o.OpnameNo, &o.WarehouseID, &o.OpnameDate,
		&o.Status, &o.Notes, &o.CreatedBy, &o.ApprovedBy, &o.ApprovedAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	linesQuery := `
		SELECT id, opname_id, product_id, system_qty, actual_qty, difference_qty, notes
		FROM stock_opname_lines
		WHERE opname_id = $1
	`
	rows, err := r.db.Query(ctx, linesQuery, o.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.StockOpnameLine
		if err := rows.Scan(
			&line.ID, &line.OpnameID, &line.ProductID, &line.SystemQty, &line.ActualQty, &line.DifferenceQty, &line.Notes,
		); err != nil {
			return nil, err
		}
		o.Lines = append(o.Lines, line)
	}

	return &o, nil
}

func (r *StockOpnameRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.StockOpname, int64, error) {
	countQuery := "SELECT COUNT(*) FROM stock_opnames WHERE company_id = $1"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, company_id, opname_no, warehouse_id, opname_date,
			   status, notes, created_by, approved_by, approved_at, created_at, updated_at
		FROM stock_opnames
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`)

	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var opnames []entity.StockOpname
	for rows.Next() {
		var o entity.StockOpname
		if err := rows.Scan(
			&o.ID, &o.CompanyID, &o.OpnameNo, &o.WarehouseID, &o.OpnameDate,
			&o.Status, &o.Notes, &o.CreatedBy, &o.ApprovedBy, &o.ApprovedAt, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		opnames = append(opnames, o)
	}

	return opnames, total, nil
}

func (r *StockOpnameRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.StockOpnameStatus, approvedBy *uuid.UUID) error {
	var approvedAt *time.Time
	if status == entity.StockOpnameStatusApproved {
		now := time.Now()
		approvedAt = &now
	}
	query := "UPDATE stock_opnames SET status = $1, approved_by = $2, approved_at = $3, updated_at = $4 WHERE company_id = $5 AND id = $6"
	_, err := r.db.Exec(ctx, query, status, approvedBy, approvedAt, time.Now(), companyID, id)
	return err
}
