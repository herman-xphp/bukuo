package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/shopspring/decimal"
)

// JournalRepository defines the interface for journal entry persistence
type JournalRepository interface {
	Create(ctx context.Context, journal *entity.JournalEntry) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.JournalEntry, error)
	GetByEntryNumber(ctx context.Context, companyID uuid.UUID, number string) (*entity.JournalEntry, error)
	GetByPeriod(ctx context.Context, periodID uuid.UUID) ([]entity.JournalEntry, error)
	GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.JournalEntry, error)
	GetByStatus(ctx context.Context, companyID uuid.UUID, status entity.JournalStatus) ([]entity.JournalEntry, error)
	GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.JournalEntry, error)
	Count(ctx context.Context, companyID uuid.UUID, search string) (int, error)
	GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error)
	Update(ctx context.Context, journal *entity.JournalEntry) error
	UpdateDetails(ctx context.Context, journal *entity.JournalEntry) error
	CountByYear(ctx context.Context, companyID uuid.UUID, year int) (int, error)

	// CreateReversalWithTransaction creates reversal journal and updates original in single transaction
	CreateReversalWithTransaction(ctx context.Context, reversal *entity.JournalEntry, originalID uuid.UUID) error

	// ClosePeriodWithTransaction creates closing journal and updates period status in single transaction
	ClosePeriodWithTransaction(ctx context.Context, closingJournal *entity.JournalEntry, period *entity.AccountingPeriod) error
}
