package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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
	return r.getJournalsWithLines(ctx,
		`WHERE j.period_id = $1 ORDER BY j.entry_date, j.entry_number, jl.line_number`,
		periodID,
	)
}

func (r *JournalRepository) GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.JournalEntry, error) {
	return r.getJournalsWithLines(ctx,
		`WHERE j.company_id = $1 AND j.entry_date BETWEEN $2 AND $3 ORDER BY j.entry_date, j.entry_number, jl.line_number`,
		companyID, start, end,
	)
}

func (r *JournalRepository) GetByStatus(ctx context.Context, companyID uuid.UUID, status entity.JournalStatus) ([]entity.JournalEntry, error) {
	return r.getJournalsWithLines(ctx,
		`WHERE j.company_id = $1 AND j.status = $2 ORDER BY j.created_at DESC, jl.line_number`,
		companyID, status,
	)
}

// getJournalsWithLines fetches journals with their lines in a single query (fixes N+1)
func (r *JournalRepository) getJournalsWithLines(ctx context.Context, whereClause string, args ...interface{}) ([]entity.JournalEntry, error) {
	query := `
		SELECT 
			j.id, j.company_id, j.period_id, j.entry_number, j.entry_date, 
			j.description, j.status, j.source_type, j.source_id, 
			j.created_by, j.created_at, j.posted_at, j.posted_by,
			jl.id, jl.line_number, jl.account_id, jl.description, 
			jl.debit_amount, jl.credit_amount
		FROM journal_entries j
		LEFT JOIN journal_lines jl ON j.id = jl.journal_id
		` + whereClause

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to aggregate lines by journal
	journalMap := make(map[uuid.UUID]*entity.JournalEntry)
	var journalOrder []uuid.UUID

	for rows.Next() {
		var j entity.JournalEntry
		var lineID, lineAccountID *uuid.UUID
		var lineNumber *int
		var lineDesc *string
		var lineDebit, lineCredit *string

		err := rows.Scan(
			&j.ID, &j.CompanyID, &j.PeriodID, &j.EntryNumber, &j.EntryDate,
			&j.Description, &j.Status, &j.SourceType, &j.SourceID,
			&j.CreatedBy, &j.CreatedAt, &j.PostedAt, &j.PostedBy,
			&lineID, &lineNumber, &lineAccountID, &lineDesc,
			&lineDebit, &lineCredit,
		)
		if err != nil {
			return nil, err
		}

		// Get or create journal in map
		existing, ok := journalMap[j.ID]
		if !ok {
			j.Lines = make([]entity.JournalLine, 0)
			journalMap[j.ID] = &j
			journalOrder = append(journalOrder, j.ID)
			existing = &j
		}

		// Add line if exists (LEFT JOIN may have null lines)
		if lineID != nil {
			line := entity.JournalLine{
				ID:          *lineID,
				JournalID:   j.ID,
				LineNumber:  *lineNumber,
				AccountID:   *lineAccountID,
				Description: *lineDesc,
			}
			if lineDebit != nil {
				line.DebitAmount, _ = decimal.NewFromString(*lineDebit)
			}
			if lineCredit != nil {
				line.CreditAmount, _ = decimal.NewFromString(*lineCredit)
			}
			existing.Lines = append(existing.Lines, line)
		}
	}

	// Convert map to slice maintaining order
	journals := make([]entity.JournalEntry, 0, len(journalOrder))
	for _, id := range journalOrder {
		journals = append(journals, *journalMap[id])
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

// CreateReversalWithTransaction creates reversal and updates original in single transaction
func (r *JournalRepository) CreateReversalWithTransaction(ctx context.Context, reversal *entity.JournalEntry, originalID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Insert reversal header
	headerQuery := `
		INSERT INTO journal_entries 
		(id, company_id, period_id, entry_number, entry_date, description, status, source_type, source_id, created_by, created_at, posted_at, posted_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, headerQuery,
		reversal.ID, reversal.CompanyID, reversal.PeriodID, reversal.EntryNumber,
		reversal.EntryDate, reversal.Description, reversal.Status, reversal.SourceType,
		reversal.SourceID, reversal.CreatedBy, reversal.CreatedAt, reversal.PostedAt, reversal.PostedBy,
	)
	if err != nil {
		return err
	}

	// 2. Insert reversal lines
	lineQuery := `
		INSERT INTO journal_lines 
		(id, journal_id, line_number, account_id, description, debit_amount, credit_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, line := range reversal.Lines {
		_, err = tx.Exec(ctx, lineQuery,
			line.ID, reversal.ID, line.LineNumber, line.AccountID,
			line.Description, line.DebitAmount, line.CreditAmount,
		)
		if err != nil {
			return err
		}
	}

	// 3. Update original journal status to REVERSED
	updateQuery := `UPDATE journal_entries SET status = 'REVERSED' WHERE id = $1`
	_, err = tx.Exec(ctx, updateQuery, originalID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ClosePeriodWithTransaction creates closing journal and updates period status in single transaction
func (r *JournalRepository) ClosePeriodWithTransaction(ctx context.Context, closingJournal *entity.JournalEntry, period *entity.AccountingPeriod) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Insert closing journal header
	headerQuery := `
		INSERT INTO journal_entries 
		(id, company_id, period_id, entry_number, entry_date, description, status, source_type, source_id, created_by, created_at, posted_at, posted_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, headerQuery,
		closingJournal.ID, closingJournal.CompanyID, closingJournal.PeriodID, closingJournal.EntryNumber,
		closingJournal.EntryDate, closingJournal.Description, closingJournal.Status, closingJournal.SourceType,
		closingJournal.SourceID, closingJournal.CreatedBy, closingJournal.CreatedAt, closingJournal.PostedAt, closingJournal.PostedBy,
	)
	if err != nil {
		return err
	}

	// 2. Insert closing journal lines
	lineQuery := `
		INSERT INTO journal_lines 
		(id, journal_id, line_number, account_id, description, debit_amount, credit_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for i, line := range closingJournal.Lines {
		_, err = tx.Exec(ctx, lineQuery,
			uuid.New(), closingJournal.ID, i+1, line.AccountID,
			line.Description, line.DebitAmount, line.CreditAmount,
		)
		if err != nil {
			return err
		}
	}

	// 3. Update period status
	periodQuery := `
		UPDATE accounting_periods
		SET status = $1, closed_at = $2, closed_by = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err = tx.Exec(ctx, periodQuery,
		period.Status, period.ClosedAt, period.ClosedBy, period.ID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
