package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// JournalRepository defines the interface for journal entry persistence
type JournalRepository interface {
	Create(ctx context.Context, journal *entity.JournalEntry) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.JournalEntry, error)
	GetByEntryNumber(ctx context.Context, companyID uuid.UUID, number string) (*entity.JournalEntry, error)
	GetByPeriod(ctx context.Context, periodID uuid.UUID) ([]entity.JournalEntry, error)
	GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.JournalEntry, error)
	Update(ctx context.Context, journal *entity.JournalEntry) error
	CountByYear(ctx context.Context, companyID uuid.UUID, year int) (int, error)

	// CreateReversalWithTransaction creates reversal journal and updates original in single transaction
	CreateReversalWithTransaction(ctx context.Context, reversal *entity.JournalEntry, originalID uuid.UUID) error
}
