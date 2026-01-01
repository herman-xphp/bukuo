package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Verify interface implementation at compile time
var _ repository.JournalRepository = (*JournalRepository)(nil)

// JournalRepository implements repository.JournalRepository for PostgreSQL
type JournalRepository struct {
	db *pgxpool.Pool
}

// NewJournalRepository creates a new JournalRepository
func NewJournalRepository(db *pgxpool.Pool) *JournalRepository {
	return &JournalRepository{db: db}
}

func (r *JournalRepository) Create(ctx context.Context, journal *entity.JournalEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert header
	headerQuery := `
		INSERT INTO journal_entries 
		(id, company_id, period_id, entry_number, entry_date, description, status, source_type, source_id, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.Exec(ctx, headerQuery,
		journal.ID, journal.CompanyID, journal.PeriodID, journal.EntryNumber,
		journal.EntryDate, journal.Description, journal.Status, journal.SourceType,
		journal.SourceID, journal.CreatedBy, journal.CreatedAt,
	)
	if err != nil {
		return err
	}

	// Insert lines
	lineQuery := `
		INSERT INTO journal_lines 
		(id, journal_id, line_number, account_id, description, debit_amount, credit_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, line := range journal.Lines {
		_, err = tx.Exec(ctx, lineQuery,
			line.ID, journal.ID, line.LineNumber, line.AccountID,
			line.Description, line.DebitAmount, line.CreditAmount,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *JournalRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.JournalEntry, error) {
	// Get header
	headerQuery := `
		SELECT id, company_id, period_id, entry_number, entry_date, description, status, 
		       source_type, source_id, created_by, created_at, posted_at, posted_by
		FROM journal_entries WHERE id = $1
	`
	var j entity.JournalEntry
	err := r.db.QueryRow(ctx, headerQuery, id).Scan(
		&j.ID, &j.CompanyID, &j.PeriodID, &j.EntryNumber, &j.EntryDate,
		&j.Description, &j.Status, &j.SourceType, &j.SourceID,
		&j.CreatedBy, &j.CreatedAt, &j.PostedAt, &j.PostedBy,
	)
	if err != nil {
		return nil, err
	}

	// Get lines
	lineQuery := `
		SELECT id, journal_id, line_number, account_id, description, debit_amount, credit_amount
		FROM journal_lines WHERE journal_id = $1 ORDER BY line_number
	`
	rows, err := r.db.Query(ctx, lineQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line entity.JournalLine
		err := rows.Scan(
			&line.ID, &line.JournalID, &line.LineNumber, &line.AccountID,
			&line.Description, &line.DebitAmount, &line.CreditAmount,
		)
		if err != nil {
			return nil, err
		}
		j.Lines = append(j.Lines, line)
	}

	return &j, nil
}

func (r *JournalRepository) GetByEntryNumber(ctx context.Context, companyID uuid.UUID, number string) (*entity.JournalEntry, error) {
	query := `SELECT id FROM journal_entries WHERE company_id = $1 AND entry_number = $2`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, companyID, number).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *JournalRepository) GetByPeriod(ctx context.Context, periodID uuid.UUID) ([]entity.JournalEntry, error) {
	query := `SELECT id FROM journal_entries WHERE period_id = $1 ORDER BY entry_date, entry_number`
	rows, err := r.db.Query(ctx, query, periodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var journals []entity.JournalEntry
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		journal, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		journals = append(journals, *journal)
	}
	return journals, nil
}

func (r *JournalRepository) GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.JournalEntry, error) {
	query := `SELECT id FROM journal_entries WHERE company_id = $1 AND entry_date BETWEEN $2 AND $3 ORDER BY entry_date, entry_number`
	rows, err := r.db.Query(ctx, query, companyID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var journals []entity.JournalEntry
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		journal, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		journals = append(journals, *journal)
	}
	return journals, nil
}

func (r *JournalRepository) Update(ctx context.Context, journal *entity.JournalEntry) error {
	query := `
		UPDATE journal_entries 
		SET status = $2, posted_at = $3, posted_by = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		journal.ID, journal.Status, journal.PostedAt, journal.PostedBy,
	)
	return err
}

func (r *JournalRepository) CountByYear(ctx context.Context, companyID uuid.UUID, year int) (int, error) {
	query := `
		SELECT COUNT(*) FROM journal_entries 
		WHERE company_id = $1 AND EXTRACT(YEAR FROM entry_date) = $2
	`
	var count int
	err := r.db.QueryRow(ctx, query, companyID, year).Scan(&count)
	return count, err
}
