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

type StockTransferRepository struct {
	db *pgxpool.Pool
}

func NewStockTransferRepository(db *pgxpool.Pool) *StockTransferRepository {
	return &StockTransferRepository{db: db}
}

func (r *StockTransferRepository) Create(ctx context.Context, transfer *entity.StockTransfer, lines []entity.StockTransferLine) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO stock_transfers (
				id, company_id, transfer_no, from_warehouse_id, to_warehouse_id,
				transfer_date, status, notes, created_by, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		if _, err := tx.Exec(ctx, query,
			transfer.ID, transfer.CompanyID, transfer.TransferNo, transfer.FromWarehouseID, transfer.ToWarehouseID,
			transfer.TransferDate, transfer.Status, transfer.Notes, transfer.CreatedBy, transfer.CreatedAt, transfer.UpdatedAt,
		); err != nil {
			return err
		}

		lineQuery := `
			INSERT INTO stock_transfer_lines (id, transfer_id, product_id, quantity)
			VALUES ($1, $2, $3, $4)
		`
		for _, line := range lines {
			if _, err := tx.Exec(ctx, lineQuery,
				line.ID, line.TransferID, line.ProductID, line.Quantity,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *StockTransferRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.StockTransfer, error) {
	query := `
		SELECT id, company_id, transfer_no, from_warehouse_id, to_warehouse_id,
			   transfer_date, status, notes, created_by, created_at, updated_at
		FROM stock_transfers
		WHERE company_id = $1 AND id = $2
	`
	var t entity.StockTransfer
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&t.ID, &t.CompanyID, &t.TransferNo, &t.FromWarehouseID, &t.ToWarehouseID,
		&t.TransferDate, &t.Status, &t.Notes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	linesQuery := `
		SELECT id, transfer_id, product_id, quantity
		FROM stock_transfer_lines
		WHERE transfer_id = $1
	`
	rows, err := r.db.Query(ctx, linesQuery, t.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.StockTransferLine
		if err := rows.Scan(&line.ID, &line.TransferID, &line.ProductID, &line.Quantity); err != nil {
			return nil, err
		}
		t.Lines = append(t.Lines, line)
	}

	return &t, nil
}

func (r *StockTransferRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.StockTransfer, int64, error) {
	countQuery := "SELECT COUNT(*) FROM stock_transfers WHERE company_id = $1"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, company_id, transfer_no, from_warehouse_id, to_warehouse_id,
			   transfer_date, status, notes, created_by, created_at, updated_at
		FROM stock_transfers
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`)

	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transfers []entity.StockTransfer
	for rows.Next() {
		var t entity.StockTransfer
		if err := rows.Scan(
			&t.ID, &t.CompanyID, &t.TransferNo, &t.FromWarehouseID, &t.ToWarehouseID,
			&t.TransferDate, &t.Status, &t.Notes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		transfers = append(transfers, t)
	}

	return transfers, total, nil
}

func (r *StockTransferRepository) UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.StockTransferStatus) error {
	query := "UPDATE stock_transfers SET status = $1, updated_at = $2 WHERE company_id = $3 AND id = $4"
	_, err := r.db.Exec(ctx, query, status, time.Now(), companyID, id)
	return err
}
