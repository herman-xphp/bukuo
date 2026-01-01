package journal_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/test/mock"
	"github.com/herman-xphp/bukuo/internal/usecase/journal"
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
