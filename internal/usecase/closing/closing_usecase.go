package closing

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/shopspring/decimal"
)

// ClosingUsecase handles period closing business logic
type ClosingUsecase struct {
	journalRepo  repository.JournalRepository
	accountRepo  repository.AccountRepository
	periodRepo   repository.PeriodRepository
	auditLogRepo repository.AuditLogRepository
	audit        *common.AuditLogger
}

// NewClosingUsecase creates a new ClosingUsecase
func NewClosingUsecase(jr repository.JournalRepository, ar repository.AccountRepository, pr repository.PeriodRepository, alr repository.AuditLogRepository) *ClosingUsecase {
	return &ClosingUsecase{
		journalRepo: jr, accountRepo: ar, periodRepo: pr, auditLogRepo: alr,
		audit: common.NewAuditLogger(alr),
	}
}

// ClosePeriodInput represents input for closing a period
type ClosePeriodInput struct {
	PeriodID           uuid.UUID
	CompanyID          uuid.UUID
	RetainedEarningsID uuid.UUID
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
	period, err := uc.periodRepo.GetByID(ctx, input.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("period not found: %w", err)
	}

	if period.Status != entity.PeriodStatusOpen {
		return nil, fmt.Errorf("period is not open, status: %s", period.Status)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, input.CompanyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	journals, err := uc.journalRepo.GetByPeriod(ctx, input.PeriodID)
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	balances := make(map[uuid.UUID]decimal.Decimal)
	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			balances[line.AccountID] = balances[line.AccountID].Add(line.DebitAmount).Sub(line.CreditAmount)
		}
	}

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
			totalRevenue = totalRevenue.Add(balance.Neg())
			if !balance.IsZero() {
				revenueAccounts = append(revenueAccounts, acc.ID)
			}
		case entity.AccountTypeExpense:
			totalExpense = totalExpense.Add(balance)
			if !balance.IsZero() {
				expenseAccounts = append(expenseAccounts, acc.ID)
			}
		}
	}

	netIncome := totalRevenue.Sub(totalExpense)

	closingJournal := entity.NewJournalEntry(input.CompanyID, input.PeriodID, input.UserID, period.EndDate, fmt.Sprintf("Closing Entry for %s", period.Name))
	closingJournal.SourceType = "CLOSING"

	for _, accID := range revenueAccounts {
		balance := balances[accID].Neg()
		if balance.IsPositive() {
			closingJournal.AddLine(accID, "Close revenue", balance, decimal.Zero)
		}
	}

	for _, accID := range expenseAccounts {
		balance := balances[accID]
		if balance.IsPositive() {
			closingJournal.AddLine(accID, "Close expense", decimal.Zero, balance)
		}
	}

	if netIncome.IsPositive() {
		closingJournal.AddLine(input.RetainedEarningsID, "Net income to retained earnings", decimal.Zero, netIncome)
	} else if netIncome.IsNegative() {
		closingJournal.AddLine(input.RetainedEarningsID, "Net loss to retained earnings", netIncome.Abs(), decimal.Zero)
	}

	count, _ := uc.journalRepo.CountByYear(ctx, input.CompanyID, period.EndDate.Year())
	closingJournal.EntryNumber = fmt.Sprintf("CL-%d-%04d", period.EndDate.Year(), count+1)

	if err := closingJournal.Post(input.UserID); err != nil {
		return nil, common.WrapErr("post closing journal", err)
	}

	if err := period.Close(input.UserID); err != nil {
		return nil, common.WrapErr("close period", err)
	}

	if err := uc.journalRepo.ClosePeriodWithTransaction(ctx, closingJournal, period); err != nil {
		return nil, common.WrapErr("execute atomic closing", err)
	}

	uc.audit.LogUpdate(ctx, input.CompanyID, &input.UserID, "accounting_period", &period.ID, fmt.Sprintf("Closed period %s and generated closing entry %s", period.Name, closingJournal.EntryNumber))

	return &ClosePeriodOutput{Period: period, ClosingJournal: closingJournal, NetIncome: netIncome}, nil
}

// PreviewClosing calculates what the closing would look like without executing
func (uc *ClosingUsecase) PreviewClosing(ctx context.Context, periodID, companyID uuid.UUID) (*ClosingPreview, error) {
	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	journals, err := uc.journalRepo.GetByPeriod(ctx, periodID)
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	balances := make(map[uuid.UUID]decimal.Decimal)
	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			balances[line.AccountID] = balances[line.AccountID].Add(line.DebitAmount).Sub(line.CreditAmount)
		}
	}

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
			revenueDetails = append(revenueDetails, AccountBalance{AccountID: acc.ID, AccountCode: acc.Code, AccountName: acc.Name, Balance: amount})
		case entity.AccountTypeExpense:
			totalExpense = totalExpense.Add(balance)
			expenseDetails = append(expenseDetails, AccountBalance{AccountID: acc.ID, AccountCode: acc.Code, AccountName: acc.Name, Balance: balance})
		}
	}

	return &ClosingPreview{
		TotalRevenue: totalRevenue, TotalExpense: totalExpense, NetIncome: totalRevenue.Sub(totalExpense),
		RevenueDetails: revenueDetails, ExpenseDetails: expenseDetails,
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
