package opening_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/test/mock"
	"github.com/herman-xphp/bukuo/internal/usecase/opening"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

func TestOpeningBalanceUsecase_ImportOpeningBalance(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New()
	periodID := uuid.New()
	userID := uuid.New()

	mockJournalRepo := new(mock.JournalRepositoryMock)
	mockAccountRepo := new(mock.AccountRepositoryMock)
	mockPeriodRepo := new(mock.PeriodRepositoryMock)

	uc := opening.NewOpeningBalanceUsecase(mockJournalRepo, mockAccountRepo, mockPeriodRepo)

	period := &entity.AccountingPeriod{
		ID:        periodID,
		CompanyID: companyID,
		Status:    entity.PeriodStatusOpen,
	}

	assetAccount := entity.Account{ID: uuid.New(), Type: entity.AccountTypeAsset, Name: "Cash"}
	equityAccount := entity.Account{ID: uuid.New(), Type: entity.AccountTypeEquity, Name: "Capital"}
	accounts := []entity.Account{assetAccount, equityAccount}

	tests := []struct {
		name    string
		input   opening.ImportOpeningBalanceInput
		setup   func()
		wantErr bool
	}{
		{
			name: "Success - Balanced Entry",
			input: opening.ImportOpeningBalanceInput{
				CompanyID:   companyID,
				PeriodID:    periodID,
				UserID:      userID,
				BalanceDate: time.Now(),
				Balances: []opening.OpeningBalanceInput{
					{AccountID: assetAccount.ID, Balance: decimal.NewFromInt(100)},  // Debit 100
					{AccountID: equityAccount.ID, Balance: decimal.NewFromInt(100)}, // Credit 100
				},
			},
			setup: func() {
				mockPeriodRepo.On("GetByID", ctx, periodID).Return(period, nil).Once()
				mockAccountRepo.On("GetByCompany", ctx, companyID).Return(accounts, nil).Once()
				mockJournalRepo.On("CountByYear", ctx, companyID, tmock.AnythingOfType("int")).Return(1, nil).Once()
				mockJournalRepo.On("Create", ctx, tmock.AnythingOfType("*entity.JournalEntry")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Failure - Unbalanced",
			input: opening.ImportOpeningBalanceInput{
				CompanyID:   companyID,
				PeriodID:    periodID,
				UserID:      userID,
				BalanceDate: time.Now(),
				Balances: []opening.OpeningBalanceInput{
					{AccountID: assetAccount.ID, Balance: decimal.NewFromInt(100)}, // Debit 100
					{AccountID: equityAccount.ID, Balance: decimal.NewFromInt(50)}, // Credit 50 (Unbalanced)
				},
			},
			setup: func() {
				mockPeriodRepo.On("GetByID", ctx, periodID).Return(period, nil).Once()
				mockAccountRepo.On("GetByCompany", ctx, companyID).Return(accounts, nil).Once()
			},
			wantErr: true,
		},
		{
			name: "Failure - Period Closed",
			input: opening.ImportOpeningBalanceInput{
				CompanyID: companyID,
				PeriodID:  periodID,
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
			_, err := uc.ImportOpeningBalance(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
