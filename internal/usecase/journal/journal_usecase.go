package journal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/shopspring/decimal"
)

// JournalUsecase handles journal entry business logic
type JournalUsecase struct {
	journalRepo repository.JournalRepository
	accountRepo repository.AccountRepository
	periodRepo  repository.PeriodRepository
}

// NewJournalUsecase creates a new JournalUsecase
func NewJournalUsecase(
	jr repository.JournalRepository,
	ar repository.AccountRepository,
	pr repository.PeriodRepository,
) *JournalUsecase {
	return &JournalUsecase{
		journalRepo: jr,
		accountRepo: ar,
		periodRepo:  pr,
	}
}

// CreateJournalInput represents input for creating a journal
type CreateJournalInput struct {
	CompanyID   uuid.UUID
	EntryDate   time.Time
	Description string
	Lines       []JournalLineInput
	CreatedBy   uuid.UUID
}

// JournalLineInput represents a journal line input
type JournalLineInput struct {
	AccountID    uuid.UUID
	Description  string
	DebitAmount  decimal.Decimal
	CreditAmount decimal.Decimal
}

// CreateJournal creates a new journal entry
func (uc *JournalUsecase) CreateJournal(ctx context.Context, input CreateJournalInput) (*entity.JournalEntry, error) {
	// 1. Get period by date
	period, err := uc.periodRepo.GetByDate(ctx, input.CompanyID, input.EntryDate)
	if err != nil {
		return nil, fmt.Errorf("period not found for date %s: %w", input.EntryDate.Format("2006-01-02"), err)
	}

	// 2. Check period can post
	if err := period.CanPost(); err != nil {
		return nil, err
	}

	// 3. Validate accounts exist and are postable
	accountIDs := make([]uuid.UUID, len(input.Lines))
	for i, l := range input.Lines {
		accountIDs[i] = l.AccountID
	}
	accounts, err := uc.accountRepo.GetByIDs(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}
	for _, id := range accountIDs {
		acc, ok := accounts[id]
		if !ok {
			return nil, fmt.Errorf("account %s not found", id)
		}
		if err := acc.CanPost(); err != nil {
			return nil, fmt.Errorf("account %s: %w", acc.Code, err)
		}
	}

	// 4. Create journal entity
	journal := entity.NewJournalEntry(input.CompanyID, period.ID, input.CreatedBy, input.EntryDate, input.Description)

	// 5. Add lines
	for _, line := range input.Lines {
		if err := journal.AddLine(line.AccountID, line.Description, line.DebitAmount, line.CreditAmount); err != nil {
			return nil, err
		}
	}

	// 6. Validate (debit = credit)
	if err := journal.Validate(); err != nil {
		return nil, err
	}

	// 7. Generate entry number
	count, _ := uc.journalRepo.CountByYear(ctx, input.CompanyID, input.EntryDate.Year())
	journal.EntryNumber = fmt.Sprintf("JE-%d-%04d", input.EntryDate.Year(), count+1)

	// 8. Save
	if err := uc.journalRepo.Create(ctx, journal); err != nil {
		return nil, fmt.Errorf("failed to save journal: %w", err)
	}

	return journal, nil
}

// GetByID retrieves a journal by ID
func (uc *JournalUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.JournalEntry, error) {
	return uc.journalRepo.GetByID(ctx, id)
}

// PostJournal posts a draft journal
func (uc *JournalUsecase) PostJournal(ctx context.Context, id, userID uuid.UUID) (*entity.JournalEntry, error) {
	journal, err := uc.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := journal.Post(userID); err != nil {
		return nil, err
	}

	if err := uc.journalRepo.Update(ctx, journal); err != nil {
		return nil, err
	}

	return journal, nil
}

// GetByPeriod retrieves all journals for a period
func (uc *JournalUsecase) GetByPeriod(ctx context.Context, periodID uuid.UUID) ([]entity.JournalEntry, error) {
	return uc.journalRepo.GetByPeriod(ctx, periodID)
}

// GetByDateRange retrieves journals within a date range
func (uc *JournalUsecase) GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.JournalEntry, error) {
	return uc.journalRepo.GetByDateRange(ctx, companyID, start, end)
}

// ReverseJournal creates a reversal entry for a posted journal
func (uc *JournalUsecase) ReverseJournal(ctx context.Context, id, userID uuid.UUID, reversalDate time.Time) (*entity.JournalEntry, error) {
	// 1. Get original journal
	original, err := uc.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Verify it can be reversed
	if !original.CanReverse() {
		return nil, fmt.Errorf("journal cannot be reversed, status: %s", original.Status)
	}

	// 3. Get period for reversal date
	period, err := uc.periodRepo.GetByDate(ctx, original.CompanyID, reversalDate)
	if err != nil {
		return nil, fmt.Errorf("period not found for reversal date: %w", err)
	}

	if err := period.CanPost(); err != nil {
		return nil, err
	}

	// 4. Create reversal journal (swap debit/credit)
	reversal := entity.NewJournalEntry(
		original.CompanyID,
		period.ID,
		userID,
		reversalDate,
		fmt.Sprintf("Reversal of %s: %s", original.EntryNumber, original.Description),
	)
	reversal.SourceType = "REVERSAL"
	reversal.SourceID = &original.ID

	// 5. Add reversed lines (swap debit and credit)
	for _, line := range original.Lines {
		reversal.AddLine(line.AccountID, line.Description, line.CreditAmount, line.DebitAmount)
	}

	// 6. Generate entry number
	count, _ := uc.journalRepo.CountByYear(ctx, original.CompanyID, reversalDate.Year())
	reversal.EntryNumber = fmt.Sprintf("JE-%d-%04d", reversalDate.Year(), count+1)

	// 7. Auto-post the reversal
	reversal.Post(userID)

	// 8. Save reversal
	if err := uc.journalRepo.Create(ctx, reversal); err != nil {
		return nil, err
	}

	// 9. Update original status to REVERSED
	original.Status = entity.JournalStatusReversed
	uc.journalRepo.Update(ctx, original)

	return reversal, nil
}
