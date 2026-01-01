package closing

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/shopspring/decimal"
)

// ClosingUsecase handles period closing business logic
type ClosingUsecase struct {
	journalRepo repository.JournalRepository
	accountRepo repository.AccountRepository
	periodRepo  repository.PeriodRepository
}

// NewClosingUsecase creates a new ClosingUsecase
func NewClosingUsecase(
	jr repository.JournalRepository,
	ar repository.AccountRepository,
	pr repository.PeriodRepository,
) *ClosingUsecase {
	return &ClosingUsecase{
		journalRepo: jr,
		accountRepo: ar,
		periodRepo:  pr,
	}
}

// ClosePeriodInput represents input for closing a period
type ClosePeriodInput struct {
	PeriodID           uuid.UUID
	CompanyID          uuid.UUID
	RetainedEarningsID uuid.UUID // Account ID for Retained Earnings
	UserID             uuid.UUID
}

// ClosePeriodOutput represents output of closing a period
type ClosePeriodOutput struct {
	Period         *entity.AccountingPeriod `json:"period"`
	ClosingJournal *entity.JournalEntry     `json:"closing_journal"`
	NetIncome      decimal.Decimal          `json:"net_income"`
}

// ClosePeriod creates closing entries and closes the period
func (uc *ClosingUsecase) ClosePeriod(ctx context.Context, input ClosePeriodInput) (*ClosePeriodOutput, error) {
	// 1. Get period and validate
	period, err := uc.periodRepo.GetByID(ctx, input.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("period not found: %w", err)
	}

	if period.Status != entity.PeriodStatusOpen {
		return nil, fmt.Errorf("period is not open, status: %s", period.Status)
	}

	// 2. Get all accounts for the company
	accounts, err := uc.accountRepo.GetByCompany(ctx, input.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	// 3. Get all posted journals in this period
	journals, err := uc.journalRepo.GetByPeriod(ctx, input.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get journals: %w", err)
	}

	// 4. Calculate balances by account type
	balances := make(map[uuid.UUID]decimal.Decimal)
	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			current := balances[line.AccountID]
			balances[line.AccountID] = current.Add(line.DebitAmount).Sub(line.CreditAmount)
		}
	}

	// 5. Calculate net income (Revenue - Expense)
	var totalRevenue, totalExpense decimal.Decimal
	revenueAccounts := make([]uuid.UUID, 0)
	expenseAccounts := make([]uuid.UUID, 0)

	for _, acc := range accounts {
		balance, ok := balances[acc.ID]
		if !ok {
			continue
		}

		switch acc.Type {
		case entity.AccountTypeRevenue:
			// Revenue has credit normal balance, so negate
			totalRevenue = totalRevenue.Add(balance.Neg())
			if !balance.IsZero() {
				revenueAccounts = append(revenueAccounts, acc.ID)
			}
		case entity.AccountTypeExpense:
			// Expense has debit normal balance
			totalExpense = totalExpense.Add(balance)
			if !balance.IsZero() {
				expenseAccounts = append(expenseAccounts, acc.ID)
			}
		}
	}

	netIncome := totalRevenue.Sub(totalExpense)

	// 6. Create closing journal entry
	closingJournal := entity.NewJournalEntry(
		input.CompanyID,
		input.PeriodID,
		input.UserID,
		period.EndDate,
		fmt.Sprintf("Closing Entry for %s", period.Name),
	)
	closingJournal.SourceType = "CLOSING"

	// 7. Close revenue accounts (debit to zero out credit balances)
	for _, accID := range revenueAccounts {
		balance := balances[accID].Neg() // Revenue balance is stored as negative
		if balance.IsPositive() {
			closingJournal.AddLine(accID, "Close revenue", balance, decimal.Zero)
		}
	}

	// 8. Close expense accounts (credit to zero out debit balances)
	for _, accID := range expenseAccounts {
		balance := balances[accID]
		if balance.IsPositive() {
			closingJournal.AddLine(accID, "Close expense", decimal.Zero, balance)
		}
	}

	// 9. Transfer net income to retained earnings
	if netIncome.IsPositive() {
		// Profit: Credit Retained Earnings
		closingJournal.AddLine(input.RetainedEarningsID, "Net income to retained earnings", decimal.Zero, netIncome)
	} else if netIncome.IsNegative() {
		// Loss: Debit Retained Earnings
		closingJournal.AddLine(input.RetainedEarningsID, "Net loss to retained earnings", netIncome.Abs(), decimal.Zero)
	}

	// 10. Generate entry number and post
	count, _ := uc.journalRepo.CountByYear(ctx, input.CompanyID, period.EndDate.Year())
	closingJournal.EntryNumber = fmt.Sprintf("CL-%d-%04d", period.EndDate.Year(), count+1)

	if err := closingJournal.Post(input.UserID); err != nil {
		return nil, fmt.Errorf("failed to post closing journal: %w", err)
	}

	// 11. Save closing journal
	if err := uc.journalRepo.Create(ctx, closingJournal); err != nil {
		return nil, fmt.Errorf("failed to save closing journal: %w", err)
	}

	// 12. Close the period
	if err := period.Close(input.UserID); err != nil {
		return nil, fmt.Errorf("failed to close period: %w", err)
	}

	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, fmt.Errorf("failed to update period: %w", err)
	}

	return &ClosePeriodOutput{
		Period:         period,
		ClosingJournal: closingJournal,
		NetIncome:      netIncome,
	}, nil
}

// PreviewClosing calculates what the closing would look like without executing
func (uc *ClosingUsecase) PreviewClosing(ctx context.Context, periodID, companyID uuid.UUID) (*ClosingPreview, error) {
	// Get all accounts and journals
	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	journals, err := uc.journalRepo.GetByPeriod(ctx, periodID)
	if err != nil {
		return nil, err
	}

	// Calculate balances
	balances := make(map[uuid.UUID]decimal.Decimal)
	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			current := balances[line.AccountID]
			balances[line.AccountID] = current.Add(line.DebitAmount).Sub(line.CreditAmount)
		}
	}

	// Calculate totals
	var totalRevenue, totalExpense decimal.Decimal
	revenueDetails := make([]AccountBalance, 0)
	expenseDetails := make([]AccountBalance, 0)

	for _, acc := range accounts {
		balance, ok := balances[acc.ID]
		if !ok || balance.IsZero() {
			continue
		}

		switch acc.Type {
		case entity.AccountTypeRevenue:
			amount := balance.Neg()
			totalRevenue = totalRevenue.Add(amount)
			revenueDetails = append(revenueDetails, AccountBalance{
				AccountID:   acc.ID,
				AccountCode: acc.Code,
				AccountName: acc.Name,
				Balance:     amount,
			})
		case entity.AccountTypeExpense:
			totalExpense = totalExpense.Add(balance)
			expenseDetails = append(expenseDetails, AccountBalance{
				AccountID:   acc.ID,
				AccountCode: acc.Code,
				AccountName: acc.Name,
				Balance:     balance,
			})
		}
	}

	return &ClosingPreview{
		TotalRevenue:   totalRevenue,
		TotalExpense:   totalExpense,
		NetIncome:      totalRevenue.Sub(totalExpense),
		RevenueDetails: revenueDetails,
		ExpenseDetails: expenseDetails,
	}, nil
}

// ClosingPreview represents a preview of closing
type ClosingPreview struct {
	TotalRevenue   decimal.Decimal  `json:"total_revenue"`
	TotalExpense   decimal.Decimal  `json:"total_expense"`
	NetIncome      decimal.Decimal  `json:"net_income"`
	RevenueDetails []AccountBalance `json:"revenue_details"`
	ExpenseDetails []AccountBalance `json:"expense_details"`
}

// AccountBalance represents account with balance
type AccountBalance struct {
	AccountID   uuid.UUID       `json:"account_id"`
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Balance     decimal.Decimal `json:"balance"`
}
