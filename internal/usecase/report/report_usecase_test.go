package report_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/test/mock"
	"github.com/herman-xphp/bukuo/internal/usecase/report"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestReportUsecase_GetTrialBalance(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New()
	periodID := uuid.New()

	mockJournalRepo := new(mock.JournalRepositoryMock)
	mockAccountRepo := new(mock.AccountRepositoryMock)
	mockPeriodRepo := new(mock.PeriodRepositoryMock)

	uc := report.NewReportUsecase(mockJournalRepo, mockAccountRepo, mockPeriodRepo)

	// Setup fake data
	period := &entity.AccountingPeriod{
		ID:        periodID,
		CompanyID: companyID,
		StartDate: time.Now().AddDate(0, 0, -30),
		EndDate:   time.Now(),
	}

	acc1ID := uuid.New()
	acc2ID := uuid.New()
	accounts := []entity.Account{
		{
			ID: acc1ID, Code: "1-1001", Name: "Cash", Type: entity.AccountTypeAsset,
			CompanyID: companyID, IsActive: true,
		},
		{
			ID: acc2ID, Code: "4-1001", Name: "Sales", Type: entity.AccountTypeRevenue,
			CompanyID: companyID, IsActive: true,
		},
	}

	journals := []entity.JournalEntry{
		{
			ID:        uuid.New(),
			CompanyID: companyID,
			PeriodID:  periodID,
			Status:    entity.JournalStatusPosted,
			Lines: []entity.JournalLine{
				{AccountID: acc1ID, DebitAmount: decimal.NewFromInt(100), CreditAmount: decimal.Zero},
				{AccountID: acc2ID, DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(100)},
			},
		},
		// Unposted journal should be ignored
		{
			ID:        uuid.New(),
			CompanyID: companyID,
			PeriodID:  periodID,
			Status:    entity.JournalStatusDraft,
			Lines: []entity.JournalLine{
				{AccountID: acc1ID, DebitAmount: decimal.NewFromInt(50), CreditAmount: decimal.Zero},
			},
		},
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
		verify  func(*report.TrialBalanceReport)
	}{
		{
			name: "Success",
			setup: func() {
				mockPeriodRepo.On("GetByID", ctx, periodID).Return(period, nil).Once()
				mockJournalRepo.On("GetByPeriod", ctx, periodID).Return(journals, nil).Once()
				mockAccountRepo.On("GetByCompany", ctx, companyID).Return(accounts, nil).Once()
			},
			wantErr: false,
			verify: func(res *report.TrialBalanceReport) {
				assert.NotNil(t, res)
				assert.True(t, res.IsBalanced)
				assert.Equal(t, decimal.NewFromInt(100), res.TotalDebit)
				assert.Equal(t, decimal.NewFromInt(100), res.TotalCredit)
				assert.Len(t, res.Items, 2)

				// Verify items
				var item1 *report.TrialBalanceItem
				for i := range res.Items {
					if res.Items[i].AccountID == acc1ID {
						item1 = &res.Items[i]
						break
					}
				}
				assert.NotNil(t, item1)
				assert.Equal(t, decimal.NewFromInt(100), item1.Debit)
			},
		},
		{
			name: "Accounts Error",
			setup: func() {
				mockPeriodRepo.On("GetByID", ctx, periodID).Return(period, nil).Once()
				mockJournalRepo.On("GetByPeriod", ctx, periodID).Return(journals, nil).Once()
				mockAccountRepo.On("GetByCompany", ctx, companyID).Return([]entity.Account(nil), errors.New("db fail")).Once()
			},
			wantErr: true,
			verify:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			res, err := uc.GetTrialBalance(ctx, companyID, periodID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.verify != nil {
					tt.verify(res)
				}
			}
		})
	}
}
