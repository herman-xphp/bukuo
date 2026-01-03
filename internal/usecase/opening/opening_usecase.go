package opening

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

// OpeningBalanceUsecase handles opening balance business logic
type OpeningBalanceUsecase struct {
	journalRepo repository.JournalRepository
	accountRepo repository.AccountRepository
	periodRepo  repository.PeriodRepository
}

// NewOpeningBalanceUsecase creates a new OpeningBalanceUsecase
func NewOpeningBalanceUsecase(jr repository.JournalRepository, ar repository.AccountRepository, pr repository.PeriodRepository) *OpeningBalanceUsecase {
	return &OpeningBalanceUsecase{journalRepo: jr, accountRepo: ar, periodRepo: pr}
}

// OpeningBalanceInput represents a single account opening balance
type OpeningBalanceInput struct {
	AccountID uuid.UUID       `json:"account_id"`
	Balance   decimal.Decimal `json:"balance"`
}

// ImportOpeningBalanceInput represents input for importing opening balances
type ImportOpeningBalanceInput struct {
	CompanyID   uuid.UUID             `json:"company_id"`
	PeriodID    uuid.UUID             `json:"period_id"`
	UserID      uuid.UUID             `json:"user_id"`
	BalanceDate time.Time             `json:"balance_date"`
	Balances    []OpeningBalanceInput `json:"balances"`
}

// ImportOpeningBalanceOutput represents output of importing opening balances
type ImportOpeningBalanceOutput struct {
	Journal      *entity.JournalEntry `json:"journal"`
	TotalDebit   decimal.Decimal      `json:"total_debit"`
	TotalCredit  decimal.Decimal      `json:"total_credit"`
	AccountCount int                  `json:"account_count"`
}

// ImportOpeningBalance creates an opening balance journal entry
func (uc *OpeningBalanceUsecase) ImportOpeningBalance(ctx context.Context, input ImportOpeningBalanceInput) (*ImportOpeningBalanceOutput, error) {
	period, err := uc.periodRepo.GetByID(ctx, input.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("period not found: %w", err)
	}

	if period.Status != entity.PeriodStatusOpen {
		return nil, fmt.Errorf("period is not open")
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, input.CompanyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	accountMap := make(map[uuid.UUID]*entity.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	journal := entity.NewJournalEntry(input.CompanyID, input.PeriodID, input.UserID, input.BalanceDate, "Opening Balance / Saldo Awal")
	journal.SourceType = "OPENING_BALANCE"

	var totalDebit, totalCredit decimal.Decimal

	for _, b := range input.Balances {
		acc, ok := accountMap[b.AccountID]
		if !ok {
			return nil, fmt.Errorf("account not found: %s", b.AccountID)
		}

		if b.Balance.IsZero() {
			continue
		}

		var debit, credit decimal.Decimal
		switch acc.NormalBalance() {
		case "DEBIT":
			if b.Balance.IsPositive() {
				debit = b.Balance
				totalDebit = totalDebit.Add(debit)
			} else {
				credit = b.Balance.Abs()
				totalCredit = totalCredit.Add(credit)
			}
		case "CREDIT":
			if b.Balance.IsPositive() {
				credit = b.Balance
				totalCredit = totalCredit.Add(credit)
			} else {
				debit = b.Balance.Abs()
				totalDebit = totalDebit.Add(debit)
			}
		}

		journal.AddLine(b.AccountID, fmt.Sprintf("Opening balance: %s", acc.Name), debit, credit)
	}

	if err := journal.Validate(); err != nil {
		return nil, fmt.Errorf("opening balance not balanced: %w (debit: %s, credit: %s)", err, totalDebit, totalCredit)
	}

	count, _ := uc.journalRepo.CountByYear(ctx, input.CompanyID, input.BalanceDate.Year())
	journal.EntryNumber = fmt.Sprintf("OB-%d-%04d", input.BalanceDate.Year(), count+1)

	if err := journal.Post(input.UserID); err != nil {
		return nil, common.WrapErr("post opening balance", err)
	}

	if err := uc.journalRepo.Create(ctx, journal); err != nil {
		return nil, common.WrapErr("save opening balance", err)
	}

	return &ImportOpeningBalanceOutput{
		Journal: journal, TotalDebit: totalDebit, TotalCredit: totalCredit, AccountCount: len(journal.Lines),
	}, nil
}

// GetOpeningBalanceTemplate returns accounts with their expected balance type
func (uc *OpeningBalanceUsecase) GetOpeningBalanceTemplate(ctx context.Context, companyID uuid.UUID) ([]OpeningBalanceTemplate, error) {
	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	templates := make([]OpeningBalanceTemplate, 0)
	for _, acc := range accounts {
		if acc.Type == entity.AccountTypeRevenue || acc.Type == entity.AccountTypeExpense {
			continue
		}
		templates = append(templates, OpeningBalanceTemplate{
			AccountID: acc.ID, AccountCode: acc.Code, AccountName: acc.Name,
			AccountType: string(acc.Type), NormalBalance: string(acc.NormalBalance()),
		})
	}

	return templates, nil
}

// OpeningBalanceTemplate represents an account template for opening balance
type OpeningBalanceTemplate struct {
	AccountID     uuid.UUID `json:"account_id"`
	AccountCode   string    `json:"account_code"`
	AccountName   string    `json:"account_name"`
	AccountType   string    `json:"account_type"`
	NormalBalance string    `json:"normal_balance"`
}
