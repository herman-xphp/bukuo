package journal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/shopspring/decimal"
)

// JournalUsecase handles journal entry business logic
type JournalUsecase struct {
	journalRepo  repository.JournalRepository
	accountRepo  repository.AccountRepository
	periodRepo   repository.PeriodRepository
	auditLogRepo repository.AuditLogRepository
	audit        *common.AuditLogger
}

// NewJournalUsecase creates a new JournalUsecase
func NewJournalUsecase(
	jr repository.JournalRepository,
	ar repository.AccountRepository,
	pr repository.PeriodRepository,
	alr repository.AuditLogRepository,
) *JournalUsecase {
	return &JournalUsecase{
		journalRepo: jr, accountRepo: ar, periodRepo: pr, auditLogRepo: alr,
		audit: common.NewAuditLogger(alr),
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
	period, err := uc.periodRepo.GetByDate(ctx, input.CompanyID, input.EntryDate)
	if err != nil {
		return nil, fmt.Errorf("period not found for date %s: %w", input.EntryDate.Format("2006-01-02"), err)
	}

	if err := period.CanPost(); err != nil {
		return nil, err
	}

	accountIDs := make([]uuid.UUID, len(input.Lines))
	for i, l := range input.Lines {
		accountIDs[i] = l.AccountID
	}

	accounts, err := uc.accountRepo.GetByIDs(ctx, accountIDs)
	if err != nil {
		return nil, common.WrapErr("fetch accounts", err)
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

	journal := entity.NewJournalEntry(input.CompanyID, period.ID, input.CreatedBy, input.EntryDate, input.Description)

	for _, line := range input.Lines {
		if err := journal.AddLine(line.AccountID, line.Description, line.DebitAmount, line.CreditAmount); err != nil {
			return nil, err
		}
	}

	if err := journal.Validate(); err != nil {
		return nil, err
	}

	count, _ := uc.journalRepo.CountByYear(ctx, input.CompanyID, input.EntryDate.Year())
	journal.EntryNumber = fmt.Sprintf("JE-%d-%04d", input.EntryDate.Year(), count+1)

	if err := uc.journalRepo.Create(ctx, journal); err != nil {
		return nil, common.WrapErr("save journal", err)
	}

	return journal, nil
}

// UpdateJournal updates an existing draft journal
func (uc *JournalUsecase) UpdateJournal(ctx context.Context, id uuid.UUID, input CreateJournalInput) (*entity.JournalEntry, error) {
	journal, err := uc.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if journal.Status != entity.JournalStatusDraft {
		return nil, fmt.Errorf("journal cannot be edited, status is %s", journal.Status)
	}

	if !journal.EntryDate.Equal(input.EntryDate) {
		period, err := uc.periodRepo.GetByDate(ctx, journal.CompanyID, input.EntryDate)
		if err != nil {
			return nil, fmt.Errorf("period not found for date %s: %w", input.EntryDate.Format("2006-01-02"), err)
		}
		if err := period.CanPost(); err != nil {
			return nil, err
		}
		journal.PeriodID = period.ID
	}

	accountIDs := make([]uuid.UUID, len(input.Lines))
	for i, l := range input.Lines {
		accountIDs[i] = l.AccountID
	}

	accounts, err := uc.accountRepo.GetByIDs(ctx, accountIDs)
	if err != nil {
		return nil, common.WrapErr("fetch accounts", err)
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

	journal.EntryDate = input.EntryDate
	journal.Description = input.Description
	journal.Lines = []entity.JournalLine{}

	for _, line := range input.Lines {
		if err := journal.AddLine(line.AccountID, line.Description, line.DebitAmount, line.CreditAmount); err != nil {
			return nil, err
		}
	}

	if err := journal.Validate(); err != nil {
		return nil, err
	}

	if err := uc.journalRepo.UpdateDetails(ctx, journal); err != nil {
		return nil, common.WrapErr("update journal", err)
	}

	uc.audit.LogUpdate(ctx, journal.CompanyID, &input.CreatedBy, "journal_entry", &journal.ID, fmt.Sprintf("Updated journal %s", journal.EntryNumber))

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
		return nil, common.WrapErr("update journal", err)
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
	original, err := uc.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !original.CanReverse() {
		return nil, fmt.Errorf("journal cannot be reversed, status: %s", original.Status)
	}

	period, err := uc.periodRepo.GetByDate(ctx, original.CompanyID, reversalDate)
	if err != nil {
		return nil, fmt.Errorf("period not found for reversal date: %w", err)
	}

	if err := period.CanPost(); err != nil {
		return nil, err
	}

	reversal := entity.NewJournalEntry(
		original.CompanyID, period.ID, userID, reversalDate,
		fmt.Sprintf("Reversal of %s: %s", original.EntryNumber, original.Description),
	)
	reversal.SourceType = "REVERSAL"
	reversal.SourceID = &original.ID

	for _, line := range original.Lines {
		reversal.AddLine(line.AccountID, line.Description, line.CreditAmount, line.DebitAmount)
	}

	count, _ := uc.journalRepo.CountByYear(ctx, original.CompanyID, reversalDate.Year())
	reversal.EntryNumber = fmt.Sprintf("JE-%d-%04d", reversalDate.Year(), count+1)
	reversal.Post(userID)

	if err := uc.journalRepo.CreateReversalWithTransaction(ctx, reversal, original.ID); err != nil {
		return nil, common.WrapErr("create reversal", err)
	}

	return reversal, nil
}

// Default approval threshold
var DefaultApprovalThreshold = decimal.NewFromInt(10000000)

// SubmitForApproval submits a journal for approval
func (uc *JournalUsecase) SubmitForApproval(ctx context.Context, id, userID uuid.UUID) (*entity.JournalEntry, error) {
	journal, err := uc.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := journal.SubmitForApproval(); err != nil {
		return nil, err
	}

	if err := uc.journalRepo.Update(ctx, journal); err != nil {
		return nil, common.WrapErr("update journal", err)
	}

	return journal, nil
}

// ApproveJournal approves a pending journal
func (uc *JournalUsecase) ApproveJournal(ctx context.Context, id, approverID uuid.UUID) (*entity.JournalEntry, error) {
	journal, err := uc.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := journal.Approve(approverID); err != nil {
		return nil, err
	}

	if err := uc.journalRepo.Update(ctx, journal); err != nil {
		return nil, common.WrapErr("update journal", err)
	}

	uc.audit.LogUpdate(ctx, journal.CompanyID, &approverID, "journal_entry", &journal.ID, fmt.Sprintf("Approved journal %s", journal.EntryNumber))

	return journal, nil
}

// RejectJournal rejects a pending journal with reason
func (uc *JournalUsecase) RejectJournal(ctx context.Context, id, rejectorID uuid.UUID, reason string) (*entity.JournalEntry, error) {
	journal, err := uc.journalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := journal.Reject(rejectorID, reason); err != nil {
		return nil, err
	}

	if err := uc.journalRepo.Update(ctx, journal); err != nil {
		return nil, common.WrapErr("update journal", err)
	}

	uc.audit.LogUpdate(ctx, journal.CompanyID, &rejectorID, "journal_entry", &journal.ID, fmt.Sprintf("Rejected journal %s: %s", journal.EntryNumber, reason))

	return journal, nil
}

// GetPendingApprovals returns journals awaiting approval
func (uc *JournalUsecase) GetPendingApprovals(ctx context.Context, companyID uuid.UUID) ([]entity.JournalEntry, error) {
	return uc.journalRepo.GetByStatus(ctx, companyID, entity.JournalStatusPendingApproval)
}

// List returns all journals for a company with pagination
func (uc *JournalUsecase) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.JournalEntry, int, error) {
	p := common.ValidatePagination(1, limit)
	if limit <= 0 {
		limit = p.PageSize
	}

	journals, err := uc.journalRepo.GetByCompany(ctx, companyID, limit, offset, search)
	if err != nil {
		return nil, 0, common.WrapErr("list journals", err)
	}

	total, err := uc.journalRepo.Count(ctx, companyID, search)
	if err != nil {
		return nil, 0, common.WrapErr("count journals", err)
	}

	return journals, total, nil
}
