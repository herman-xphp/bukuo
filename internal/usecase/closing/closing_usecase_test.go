package closing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/test/mock"
	"github.com/herman-xphp/bukuo/internal/usecase/closing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

func TestClosingUsecase_ClosePeriod(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New()
	userID := uuid.New()
	periodID := uuid.New()
	retainedEarningsID := uuid.New()

	mockJournalRepo := new(mock.JournalRepositoryMock)
	mockAccountRepo := new(mock.AccountRepositoryMock)
	mockPeriodRepo := new(mock.PeriodRepositoryMock)
	mockAuditRepo := new(mock.AuditLogRepositoryMock)

	uc := closing.NewClosingUsecase(mockJournalRepo, mockAccountRepo, mockPeriodRepo, mockAuditRepo)

	// Setup fixtures
	period := &entity.AccountingPeriod{
		ID:        periodID,
		CompanyID: companyID,
		Name:      "Jan 2025",
		StartDate: time.Now().AddDate(0, -1, 0),
		EndDate:   time.Now(),
		Status:    entity.PeriodStatusOpen,
	}

	revenueAccount := entity.Account{ID: uuid.New(), Type: entity.AccountTypeRevenue, Name: "Sales"}
	expenseAccount := entity.Account{ID: uuid.New(), Type: entity.AccountTypeExpense, Name: "Cost"}

	accounts := []entity.Account{revenueAccount, expenseAccount}

	// Mock data: Revenue 100, Expense 40 => Net Income 60
	journals := []entity.JournalEntry{
		{
			Status: entity.JournalStatusPosted,
			Lines: []entity.JournalLine{
				{AccountID: revenueAccount.ID, CreditAmount: decimal.NewFromInt(100)}, // Sales Credit 100
				{AccountID: expenseAccount.ID, DebitAmount: decimal.NewFromInt(40)},   // Expense Debit 40
			},
		},
	}

	tests := []struct {
		name    string
		setup   func()
		input   closing.ClosePeriodInput
		wantErr bool
	}{
		{
			name: "Success",
			input: closing.ClosePeriodInput{
				PeriodID:           periodID,
				CompanyID:          companyID,
				UserID:             userID,
				RetainedEarningsID: retainedEarningsID,
			},
			setup: func() {
				mockPeriodRepo.On("GetByID", ctx, periodID).Return(period, nil).Once()
				mockAccountRepo.On("GetByCompany", ctx, companyID).Return(accounts, nil).Once()
				mockJournalRepo.On("GetByPeriod", ctx, periodID).Return(journals, nil).Once()
				mockJournalRepo.On("CountByYear", ctx, companyID, period.EndDate.Year()).Return(1, nil).Once()

				// Expect transaction call
				mockJournalRepo.On("ClosePeriodWithTransaction", ctx, tmock.AnythingOfType("*entity.JournalEntry"), period).Return(nil).Once()

				// Expect audit log
				mockAuditRepo.On("Create", ctx, tmock.AnythingOfType("*entity.AuditLog")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Period Not Found",
			input: closing.ClosePeriodInput{
				PeriodID: periodID,
			},
			setup: func() {
				mockPeriodRepo.On("GetByID", ctx, periodID).Return(nil, errors.New("not found")).Once()
			},
			wantErr: true,
		},
		{
			name: "Period Already Closed",
			input: closing.ClosePeriodInput{
				PeriodID: periodID,
			},
			setup: func() {
				closedPeriod := *period
				closedPeriod.Status = entity.PeriodStatusClosed
				mockPeriodRepo.On("GetByID", ctx, periodID).Return(&closedPeriod, nil).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := uc.ClosePeriod(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			// Clean up mocks for next run if needed, but On().Once() handles it well usually
		})
	}
}
