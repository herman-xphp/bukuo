package journal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/test/mock"
	"github.com/herman-xphp/bukuo/internal/usecase/journal"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

func TestJournalUsecase_ApproveJournal(t *testing.T) {
	ctx := context.Background()
	journalID := uuid.New()
	approverID := uuid.New()

	mockJournalRepo := new(mock.JournalRepositoryMock)
	mockAccountRepo := new(mock.AccountRepositoryMock)
	mockPeriodRepo := new(mock.PeriodRepositoryMock)
	mockAuditRepo := new(mock.AuditLogRepositoryMock)

	uc := journal.NewJournalUsecase(mockJournalRepo, mockAccountRepo, mockPeriodRepo, mockAuditRepo)

	pendingJournal := &entity.JournalEntry{
		ID:     journalID,
		Status: entity.JournalStatusPendingApproval,
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "Success",
			setup: func() {
				mockJournalRepo.On("GetByID", ctx, journalID).Return(pendingJournal, nil).Once()
				mockJournalRepo.On("Update", ctx, tmock.AnythingOfType("*entity.JournalEntry")).Return(nil).Once()
				mockAuditRepo.On("Create", ctx, tmock.AnythingOfType("*entity.AuditLog")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Journal Not Found",
			setup: func() {
				mockJournalRepo.On("GetByID", ctx, journalID).Return(nil, errors.New("not found")).Once()
			},
			wantErr: true,
		},
		{
			name: "Invalid Status",
			setup: func() {
				draftJournal := &entity.JournalEntry{ID: journalID, Status: entity.JournalStatusDraft}
				mockJournalRepo.On("GetByID", ctx, journalID).Return(draftJournal, nil).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			res, err := uc.ApproveJournal(ctx, journalID, approverID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, entity.JournalStatusApproved, res.Status)
				assert.NotNil(t, res.ApprovedAt)
				assert.Equal(t, approverID, *res.ApprovedBy)
			}
		})
	}
}

func TestJournalUsecase_RejectJournal(t *testing.T) {
	ctx := context.Background()
	journalID := uuid.New()
	rejectorID := uuid.New()
	reason := "Duplicate entry"

	mockJournalRepo := new(mock.JournalRepositoryMock)
	mockAccountRepo := new(mock.AccountRepositoryMock)
	mockPeriodRepo := new(mock.PeriodRepositoryMock)
	mockAuditRepo := new(mock.AuditLogRepositoryMock)

	uc := journal.NewJournalUsecase(mockJournalRepo, mockAccountRepo, mockPeriodRepo, mockAuditRepo)

	pendingJournal := &entity.JournalEntry{
		ID:     journalID,
		Status: entity.JournalStatusPendingApproval,
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "Success",
			setup: func() {
				mockJournalRepo.On("GetByID", ctx, journalID).Return(pendingJournal, nil).Once()
				mockJournalRepo.On("Update", ctx, tmock.AnythingOfType("*entity.JournalEntry")).Return(nil).Once()
				mockAuditRepo.On("Create", ctx, tmock.AnythingOfType("*entity.AuditLog")).Return(nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			res, err := uc.RejectJournal(ctx, journalID, rejectorID, reason)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, entity.JournalStatusRejected, res.Status)
				assert.NotNil(t, res.RejectedAt)
				assert.Equal(t, rejectorID, *res.RejectedBy)
				assert.Equal(t, reason, res.RejectReason)
			}
		})
	}
}

func TestJournalUsecase_UpdateJournal(t *testing.T) {
	ctx := context.Background()
	journalID := uuid.New()
	companyID := uuid.New()
	userID := uuid.New()
	periodID := uuid.New()

	mockJournalRepo := new(mock.JournalRepositoryMock)
	mockAccountRepo := new(mock.AccountRepositoryMock)
	mockPeriodRepo := new(mock.PeriodRepositoryMock)
	mockAuditRepo := new(mock.AuditLogRepositoryMock)

	uc := journal.NewJournalUsecase(mockJournalRepo, mockAccountRepo, mockPeriodRepo, mockAuditRepo)

	now := time.Now()

	// Existing Journal
	existing := &entity.JournalEntry{
		ID:        journalID,
		CompanyID: companyID,
		Status:    entity.JournalStatusDraft, // Can only update Draft
		EntryDate: now,
		PeriodID:  periodID,
	}

	// Input
	acc1 := uuid.New()
	acc2 := uuid.New()
	input := journal.CreateJournalInput{
		CompanyID:   companyID,
		EntryDate:   now,
		Description: "Updated desc",
		CreatedBy:   userID,
		Lines: []journal.JournalLineInput{
			{AccountID: acc1, DebitAmount: decimal.NewFromInt(200), CreditAmount: decimal.Zero},
			{AccountID: acc2, DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(200)},
		},
	}

	accountMap := map[uuid.UUID]*entity.Account{
		acc1: {ID: acc1, IsActive: true, IsPostable: true, CompanyID: companyID},
		acc2: {ID: acc2, IsActive: true, IsPostable: true, CompanyID: companyID},
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "Success",
			setup: func() {
				mockJournalRepo.On("GetByID", ctx, journalID).Return(existing, nil).Once()
				// Accounts check: Mock Expect Call with IDs
				mockAccountRepo.On("GetByIDs", ctx, tmock.MatchedBy(func(ids []uuid.UUID) bool {
					return len(ids) == 2
				})).Return(accountMap, nil).Once()

				// Update
				mockJournalRepo.On("UpdateDetails", ctx, tmock.AnythingOfType("*entity.JournalEntry")).Return(nil).Once()
				mockAuditRepo.On("Create", ctx, tmock.AnythingOfType("*entity.AuditLog")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Fail - Not Draft",
			setup: func() {
				posted := &entity.JournalEntry{ID: journalID, Status: entity.JournalStatusPosted}
				mockJournalRepo.On("GetByID", ctx, journalID).Return(posted, nil).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			res, err := uc.UpdateJournal(ctx, journalID, input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "Updated desc", res.Description)
			}
		})
	}
}

func TestJournalUsecase_ReverseJournal(t *testing.T) {
	ctx := context.Background()
	journalID := uuid.New()
	companyID := uuid.New()
	userID := uuid.New()
	periodID := uuid.New()

	mockJournalRepo := new(mock.JournalRepositoryMock)
	mockAccountRepo := new(mock.AccountRepositoryMock)
	mockPeriodRepo := new(mock.PeriodRepositoryMock)
	mockAuditRepo := new(mock.AuditLogRepositoryMock)

	uc := journal.NewJournalUsecase(mockJournalRepo, mockAccountRepo, mockPeriodRepo, mockAuditRepo)

	now := time.Now()

	original := &entity.JournalEntry{
		ID:          journalID,
		CompanyID:   companyID,
		Status:      entity.JournalStatusPosted, // Can only reverse Posted
		EntryDate:   now.AddDate(0, 0, -1),
		PeriodID:    periodID,
		EntryNumber: "JE-001",
		Lines: []entity.JournalLine{
			{AccountID: uuid.New(), DebitAmount: decimal.NewFromInt(100), CreditAmount: decimal.Zero},
			{AccountID: uuid.New(), DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(100)},
		},
	}

	period := &entity.AccountingPeriod{
		ID:        periodID,
		CompanyID: companyID,
		Status:    entity.PeriodStatusOpen,
		StartDate: now.AddDate(0, 0, -30),
		EndDate:   now.AddDate(0, 0, 30),
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "Success",
			setup: func() {
				mockJournalRepo.On("GetByID", ctx, journalID).Return(original, nil).Once()
				mockPeriodRepo.On("GetByDate", ctx, companyID, tmock.Anything).Return(period, nil).Once()
				mockJournalRepo.On("CountByYear", ctx, companyID, tmock.Anything).Return(1, nil).Once()
				// Transaction
				mockJournalRepo.On("CreateReversalWithTransaction", ctx, tmock.AnythingOfType("*entity.JournalEntry"), original.ID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Fail - Not Posted",
			setup: func() {
				draft := &entity.JournalEntry{ID: journalID, Status: entity.JournalStatusDraft}
				mockJournalRepo.On("GetByID", ctx, journalID).Return(draft, nil).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			res, err := uc.ReverseJournal(ctx, journalID, userID, now)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, entity.JournalStatusPosted, res.Status)
				// Verify amounts swapped (checking first line)
				assert.Equal(t, decimal.Zero, res.Lines[0].DebitAmount)
				assert.Equal(t, decimal.NewFromInt(100), res.Lines[0].CreditAmount)
			}
		})
	}
}
