package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JournalRepository struct {
	db *pgxpool.Pool
}

func NewJournalRepository(db *pgxpool.Pool) *JournalRepository {
	return &JournalRepository{db: db}
}

// CreateWithLines - Simpan journal + lines dalam 1 transaction
func (r *JournalRepository) CreateWithLines(ctx context.Context, journal *domain.JournalEntry) error {
	tx, err := r.db.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	// Insert header
	headerQuery := `
	INSERT INTO journal_entries (id, company_id, periode_id, entry_number, entry_date, description, status, created_by)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.Exec(ctx, headerQuery, journal.ID, journal.CompanyID, journal.PeriodID, journal.EntryNumber, journal.EntryDate, journal.Description, journal.Status, journal.CreatedBy)
	if err != nil {
		return err
	}

	// Insert lines
	lineQuery := `
	INSERT INTO journal_lines (id, journal_id, line_number, account_id, description, debit_amount, credit_amount)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, line := range journal.Lines {
		_, err = tx.Exec(ctx, lineQuery, line.ID, journal.ID, line.LineNumber, line.AccountID, line.Description, line.DebitAmount, line.CreditAmount)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r *JournalRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.JournalEntry, error) {
	// Get header
	headerQuery := `SELECT id, company_id, period_id, entry_number, entry_date, description, status, created_by
	                FROM journal_entries WHERE id = $1`
	var j domain.JournalEntry
	err := r.db.QueryRow(ctx, headerQuery, id).Scan(
		&j.ID, &j.CompanyID, &j.PeriodID, &j.EntryNumber,
		&j.EntryDate, &j.Description, &j.Status, &j.CreatedBy,
	)
	if err != nil {
		return nil, err
	}
	// Get lines
	lineQuery := `SELECT id, journal_id, line_number, account_id, description, debit_amount, credit_amount
	              FROM journal_lines WHERE journal_id = $1 ORDER BY line_number`
	rows, err := r.db.Query(ctx, lineQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var line domain.JournalLine
		err := rows.Scan(&line.ID, &line.JournalID, &line.LineNumber, &line.AccountID,
			&line.Description, &line.DebitAmount, &line.CreditAmount)
		if err != nil {
			return nil, err
		}
		j.Lines = append(j.Lines, line)
	}
	return &j, nil
}
