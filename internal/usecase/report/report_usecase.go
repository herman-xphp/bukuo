package report

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/shopspring/decimal"
)

// ReportUsecase handles accounting reports
type ReportUsecase struct {
	journalRepo repository.JournalRepository
	accountRepo repository.AccountRepository
	periodRepo  repository.PeriodRepository
}

// NewReportUsecase creates a new ReportUsecase
func NewReportUsecase(
	jr repository.JournalRepository,
	ar repository.AccountRepository,
	pr repository.PeriodRepository,
) *ReportUsecase {
	return &ReportUsecase{
		journalRepo: jr,
		accountRepo: ar,
		periodRepo:  pr,
	}
}

// TrialBalanceItem represents one row in trial balance
type TrialBalanceItem struct {
	AccountID   uuid.UUID       `json:"account_id"`
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	AccountType string          `json:"account_type"`
	Debit       decimal.Decimal `json:"debit"`
	Credit      decimal.Decimal `json:"credit"`
}

// TrialBalanceReport represents the trial balance report
type TrialBalanceReport struct {
	CompanyID   uuid.UUID          `json:"company_id"`
	PeriodID    uuid.UUID          `json:"period_id"`
	StartDate   time.Time          `json:"start_date"`
	EndDate     time.Time          `json:"end_date"`
	Items       []TrialBalanceItem `json:"items"`
	TotalDebit  decimal.Decimal    `json:"total_debit"`
	TotalCredit decimal.Decimal    `json:"total_credit"`
	IsBalanced  bool               `json:"is_balanced"`
}

// GetTrialBalance generates trial balance report
func (uc *ReportUsecase) GetTrialBalance(ctx context.Context, companyID, periodID uuid.UUID) (*TrialBalanceReport, error) {
	period, err := uc.periodRepo.GetByID(ctx, periodID)
	if err != nil {
		return nil, err
	}

	journals, err := uc.journalRepo.GetByPeriod(ctx, periodID)
	if err != nil {
		return nil, err
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Build account map
	accountMap := make(map[uuid.UUID]*entity.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	// Calculate balances
	balances := make(map[uuid.UUID]struct {
		debit  decimal.Decimal
		credit decimal.Decimal
	})

	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			bal := balances[line.AccountID]
			bal.debit = bal.debit.Add(line.DebitAmount)
			bal.credit = bal.credit.Add(line.CreditAmount)
			balances[line.AccountID] = bal
		}
	}

	// Build report items
	var items []TrialBalanceItem
	totalDebit := decimal.Zero
	totalCredit := decimal.Zero

	for _, acc := range accounts {
		bal, ok := balances[acc.ID]
		if !ok {
			continue
		}

		item := TrialBalanceItem{
			AccountID:   acc.ID,
			AccountCode: acc.Code,
			AccountName: acc.Name,
			AccountType: string(acc.Type),
			Debit:       bal.debit,
			Credit:      bal.credit,
		}
		items = append(items, item)
		totalDebit = totalDebit.Add(bal.debit)
		totalCredit = totalCredit.Add(bal.credit)
	}

	return &TrialBalanceReport{
		CompanyID:   companyID,
		PeriodID:    periodID,
		StartDate:   period.StartDate,
		EndDate:     period.EndDate,
		Items:       items,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  totalDebit.Equal(totalCredit),
	}, nil
}

// LedgerEntry represents a ledger entry
type LedgerEntry struct {
	Date        time.Time       `json:"date"`
	EntryNumber string          `json:"entry_number"`
	Description string          `json:"description"`
	Debit       decimal.Decimal `json:"debit"`
	Credit      decimal.Decimal `json:"credit"`
	Balance     decimal.Decimal `json:"balance"`
}

// GeneralLedger represents the general ledger for an account
type GeneralLedger struct {
	AccountID      uuid.UUID       `json:"account_id"`
	AccountCode    string          `json:"account_code"`
	AccountName    string          `json:"account_name"`
	StartDate      time.Time       `json:"start_date"`
	EndDate        time.Time       `json:"end_date"`
	OpeningBalance decimal.Decimal `json:"opening_balance"`
	Entries        []LedgerEntry   `json:"entries"`
	ClosingBalance decimal.Decimal `json:"closing_balance"`
}

// GetGeneralLedger generates general ledger for an account
func (uc *ReportUsecase) GetGeneralLedger(ctx context.Context, accountID uuid.UUID, start, end time.Time) (*GeneralLedger, error) {
	account, err := uc.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	journals, err := uc.journalRepo.GetByDateRange(ctx, account.CompanyID, start, end)
	if err != nil {
		return nil, err
	}

	var entries []LedgerEntry
	balance := decimal.Zero
	isDebitAccount := account.NormalBalance() == "DEBIT"

	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			if line.AccountID != accountID {
				continue
			}

			if isDebitAccount {
				balance = balance.Add(line.DebitAmount).Sub(line.CreditAmount)
			} else {
				balance = balance.Add(line.CreditAmount).Sub(line.DebitAmount)
			}

			entries = append(entries, LedgerEntry{
				Date:        journal.EntryDate,
				EntryNumber: journal.EntryNumber,
				Description: line.Description,
				Debit:       line.DebitAmount,
				Credit:      line.CreditAmount,
				Balance:     balance,
			})
		}
	}

	return &GeneralLedger{
		AccountID:      accountID,
		AccountCode:    account.Code,
		AccountName:    account.Name,
		StartDate:      start,
		EndDate:        end,
		OpeningBalance: decimal.Zero,
		Entries:        entries,
		ClosingBalance: balance,
	}, nil
}

// IncomeStatementItem represents a line in income statement
type IncomeStatementItem struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Amount      decimal.Decimal `json:"amount"`
}

// IncomeStatement represents the income statement
type IncomeStatement struct {
	CompanyID    uuid.UUID             `json:"company_id"`
	StartDate    time.Time             `json:"start_date"`
	EndDate      time.Time             `json:"end_date"`
	Revenue      []IncomeStatementItem `json:"revenue"`
	Expenses     []IncomeStatementItem `json:"expenses"`
	TotalRevenue decimal.Decimal       `json:"total_revenue"`
	TotalExpense decimal.Decimal       `json:"total_expense"`
	NetIncome    decimal.Decimal       `json:"net_income"`
}

// GetIncomeStatement generates income statement
func (uc *ReportUsecase) GetIncomeStatement(ctx context.Context, companyID uuid.UUID, start, end time.Time) (*IncomeStatement, error) {
	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, start, end)
	if err != nil {
		return nil, err
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Build account map
	accountMap := make(map[uuid.UUID]*entity.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	// Calculate account balances from posted journals
	balances := make(map[uuid.UUID]decimal.Decimal)
	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			acc := accountMap[line.AccountID]
			if acc == nil {
				continue
			}
			// Revenue: credit - debit, Expense: debit - credit
			if acc.Type == entity.AccountTypeRevenue {
				balances[line.AccountID] = balances[line.AccountID].Add(line.CreditAmount).Sub(line.DebitAmount)
			} else if acc.Type == entity.AccountTypeExpense {
				balances[line.AccountID] = balances[line.AccountID].Add(line.DebitAmount).Sub(line.CreditAmount)
			}
		}
	}

	var revenue, expenses []IncomeStatementItem
	totalRevenue := decimal.Zero
	totalExpense := decimal.Zero

	for id, amount := range balances {
		acc := accountMap[id]
		if acc == nil {
			continue
		}
		item := IncomeStatementItem{
			AccountCode: acc.Code,
			AccountName: acc.Name,
			Amount:      amount,
		}
		if acc.Type == entity.AccountTypeRevenue {
			revenue = append(revenue, item)
			totalRevenue = totalRevenue.Add(amount)
		} else if acc.Type == entity.AccountTypeExpense {
			expenses = append(expenses, item)
			totalExpense = totalExpense.Add(amount)
		}
	}

	return &IncomeStatement{
		CompanyID:    companyID,
		StartDate:    start,
		EndDate:      end,
		Revenue:      revenue,
		Expenses:     expenses,
		TotalRevenue: totalRevenue,
		TotalExpense: totalExpense,
		NetIncome:    totalRevenue.Sub(totalExpense),
	}, nil
}
