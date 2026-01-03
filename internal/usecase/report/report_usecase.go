package report

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/shopspring/decimal"
)

// ReportUsecase handles accounting reports
type ReportUsecase struct {
	journalRepo repository.JournalRepository
	accountRepo repository.AccountRepository
	periodRepo  repository.PeriodRepository
}

// NewReportUsecase creates a new ReportUsecase
func NewReportUsecase(jr repository.JournalRepository, ar repository.AccountRepository, pr repository.PeriodRepository) *ReportUsecase {
	return &ReportUsecase{journalRepo: jr, accountRepo: ar, periodRepo: pr}
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
		return nil, common.WrapErr("get journals", err)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	accountMap := make(map[uuid.UUID]*entity.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	balances := make(map[uuid.UUID]struct{ debit, credit decimal.Decimal })
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

	var items []TrialBalanceItem
	totalDebit := decimal.Zero
	totalCredit := decimal.Zero

	for _, acc := range accounts {
		bal, ok := balances[acc.ID]
		if !ok {
			continue
		}
		items = append(items, TrialBalanceItem{
			AccountID: acc.ID, AccountCode: acc.Code, AccountName: acc.Name,
			AccountType: string(acc.Type), Debit: bal.debit, Credit: bal.credit,
		})
		totalDebit = totalDebit.Add(bal.debit)
		totalCredit = totalCredit.Add(bal.credit)
	}

	return &TrialBalanceReport{
		CompanyID: companyID, PeriodID: periodID, StartDate: period.StartDate, EndDate: period.EndDate,
		Items: items, TotalDebit: totalDebit, TotalCredit: totalCredit, IsBalanced: totalDebit.Equal(totalCredit),
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
		return nil, common.WrapErr("get journals", err)
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
				Date: journal.EntryDate, EntryNumber: journal.EntryNumber, Description: line.Description,
				Debit: line.DebitAmount, Credit: line.CreditAmount, Balance: balance,
			})
		}
	}

	return &GeneralLedger{
		AccountID: accountID, AccountCode: account.Code, AccountName: account.Name,
		StartDate: start, EndDate: end, OpeningBalance: decimal.Zero, Entries: entries, ClosingBalance: balance,
	}, nil
}

// IncomeStatementItem and IncomeStatement
type IncomeStatementItem struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Amount      decimal.Decimal `json:"amount"`
}

type IncomeStatement struct {
	CompanyID     uuid.UUID             `json:"company_id"`
	StartDate     time.Time             `json:"start_date"`
	EndDate       time.Time             `json:"end_date"`
	Revenue       []IncomeStatementItem `json:"revenue"`
	Expenses      []IncomeStatementItem `json:"expenses"`
	TotalRevenue  decimal.Decimal       `json:"total_revenue"`
	TotalExpenses decimal.Decimal       `json:"total_expenses"`
	NetIncome     decimal.Decimal       `json:"net_income"`
}

func (uc *ReportUsecase) GetIncomeStatement(ctx context.Context, companyID uuid.UUID, start, end time.Time) (*IncomeStatement, error) {
	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, start, end)
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	accountMap := make(map[uuid.UUID]*entity.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

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
			if acc.Type == entity.AccountTypeRevenue {
				balances[line.AccountID] = balances[line.AccountID].Add(line.CreditAmount).Sub(line.DebitAmount)
			} else if acc.Type == entity.AccountTypeExpense {
				balances[line.AccountID] = balances[line.AccountID].Add(line.DebitAmount).Sub(line.CreditAmount)
			}
		}
	}

	var revenue, expenses []IncomeStatementItem
	totalRevenue, totalExpense := decimal.Zero, decimal.Zero

	for id, amount := range balances {
		acc := accountMap[id]
		if acc == nil {
			continue
		}
		item := IncomeStatementItem{AccountCode: acc.Code, AccountName: acc.Name, Amount: amount}
		if acc.Type == entity.AccountTypeRevenue {
			revenue = append(revenue, item)
			totalRevenue = totalRevenue.Add(amount)
		} else if acc.Type == entity.AccountTypeExpense {
			expenses = append(expenses, item)
			totalExpense = totalExpense.Add(amount)
		}
	}

	return &IncomeStatement{
		CompanyID: companyID, StartDate: start, EndDate: end, Revenue: revenue, Expenses: expenses,
		TotalRevenue: totalRevenue, TotalExpenses: totalExpense, NetIncome: totalRevenue.Sub(totalExpense),
	}, nil
}

// BalanceSheetItem and BalanceSheet
type BalanceSheetItem struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Balance     decimal.Decimal `json:"balance"`
}

type BalanceSheet struct {
	CompanyID        uuid.UUID          `json:"company_id"`
	AsOfDate         time.Time          `json:"as_of_date"`
	Assets           []BalanceSheetItem `json:"assets"`
	Liabilities      []BalanceSheetItem `json:"liabilities"`
	Equity           []BalanceSheetItem `json:"equity"`
	TotalAssets      decimal.Decimal    `json:"total_assets"`
	TotalLiabilities decimal.Decimal    `json:"total_liabilities"`
	TotalEquity      decimal.Decimal    `json:"total_equity"`
	IsBalanced       bool               `json:"is_balanced"`
}

func (uc *ReportUsecase) GetBalanceSheet(ctx context.Context, companyID uuid.UUID, asOfDate time.Time) (*BalanceSheet, error) {
	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, time.Time{}, asOfDate)
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	accountMap := make(map[uuid.UUID]*entity.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

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
			if acc.Type == entity.AccountTypeAsset || acc.Type == entity.AccountTypeExpense {
				balances[line.AccountID] = balances[line.AccountID].Add(line.DebitAmount).Sub(line.CreditAmount)
			} else {
				balances[line.AccountID] = balances[line.AccountID].Add(line.CreditAmount).Sub(line.DebitAmount)
			}
		}
	}

	var assets, liabilities, equity []BalanceSheetItem
	totalAssets, totalLiabilities, totalEquity := decimal.Zero, decimal.Zero, decimal.Zero

	for id, balance := range balances {
		acc := accountMap[id]
		if acc == nil {
			continue
		}
		item := BalanceSheetItem{AccountCode: acc.Code, AccountName: acc.Name, Balance: balance}
		switch acc.Type {
		case entity.AccountTypeAsset:
			assets = append(assets, item)
			totalAssets = totalAssets.Add(balance)
		case entity.AccountTypeLiability:
			liabilities = append(liabilities, item)
			totalLiabilities = totalLiabilities.Add(balance)
		case entity.AccountTypeEquity:
			equity = append(equity, item)
			totalEquity = totalEquity.Add(balance)
		}
	}

	return &BalanceSheet{
		CompanyID: companyID, AsOfDate: asOfDate, Assets: assets, Liabilities: liabilities, Equity: equity,
		TotalAssets: totalAssets, TotalLiabilities: totalLiabilities, TotalEquity: totalEquity,
		IsBalanced: totalAssets.Equal(totalLiabilities.Add(totalEquity)),
	}, nil
}

// CashFlowItem and CashFlow
type CashFlowItem struct {
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
}

type CashFlow struct {
	CompanyID      uuid.UUID       `json:"company_id"`
	StartDate      time.Time       `json:"start_date"`
	EndDate        time.Time       `json:"end_date"`
	Operating      []CashFlowItem  `json:"operating"`
	TotalOperating decimal.Decimal `json:"total_operating"`
	Investing      []CashFlowItem  `json:"investing"`
	TotalInvesting decimal.Decimal `json:"total_investing"`
	Financing      []CashFlowItem  `json:"financing"`
	TotalFinancing decimal.Decimal `json:"total_financing"`
	NetCashChange  decimal.Decimal `json:"net_cash_change"`
}

func (uc *ReportUsecase) GetCashFlow(ctx context.Context, companyID uuid.UUID, start, end time.Time) (*CashFlow, error) {
	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, start, end)
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	accountMap := make(map[uuid.UUID]*entity.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	operating := []CashFlowItem{}
	totalOperating := decimal.Zero

	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			acc := accountMap[line.AccountID]
			if acc == nil {
				continue
			}
			if acc.Code[:1] == "1" && (acc.Name == "Kas" || acc.Name == "Bank") {
				amount := line.DebitAmount.Sub(line.CreditAmount)
				if !amount.IsZero() {
					operating = append(operating, CashFlowItem{Description: journal.Description, Amount: amount})
					totalOperating = totalOperating.Add(amount)
				}
			}
		}
	}

	return &CashFlow{
		CompanyID: companyID, StartDate: start, EndDate: end, Operating: operating, TotalOperating: totalOperating,
		Investing: []CashFlowItem{}, TotalInvesting: decimal.Zero, Financing: []CashFlowItem{}, TotalFinancing: decimal.Zero,
		NetCashChange: totalOperating,
	}, nil
}

// GetSalesTrend aggregates daily sales for charts
func (uc *ReportUsecase) GetSalesTrend(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.SalesTrendItem, error) {
	// Note: Ideally this should access a SalesRepository.
	// But since Sales usually post journals, we can also query the JournalRepository filtering by Revenue Accounts,
	// OR query SalesInvoiceRepository if available.
	// For now, let's use JournalRepository with AccountTypeRevenue to be generic & consistent with accounting.

	// However, pure sales might be better tracked via invoices for "Sales Trend".
	// Let's stick to Reporting Logic via Journals (GL) as it's the source of truth for "Realized Revenue".

	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, start, end)
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	revenueAccMap := make(map[uuid.UUID]bool)
	for _, acc := range accounts {
		if acc.Type == entity.AccountTypeRevenue {
			revenueAccMap[acc.ID] = true
		}
	}

	dailyMap := make(map[string]decimal.Decimal)

	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}

		dateStr := journal.EntryDate.Format("2006-01-02")

		for _, line := range journal.Lines {
			if revenueAccMap[line.AccountID] {
				// Revenue is Credit balance usually.
				// Credit adds to revenue, Debit reduces it (returns).
				amount := line.CreditAmount.Sub(line.DebitAmount)
				dailyMap[dateStr] = dailyMap[dateStr].Add(amount)
			}
		}
	}

	// Fill ALL dates in range
	var trends []entity.SalesTrendItem
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		amount := dailyMap[dateKey] // Default zero if not found
		trends = append(trends, entity.SalesTrendItem{Date: dateKey, Amount: amount})
	}

	return trends, nil
}

// GetExpenseBreakdown aggregates expenses by category (top 5)
func (uc *ReportUsecase) GetExpenseBreakdown(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.ExpenseBreakdownItem, error) {
	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, start, end)
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	expenseAccMap := make(map[uuid.UUID]string) // ID -> Account Name (or Child Category if we had it)
	for _, acc := range accounts {
		if acc.Type == entity.AccountTypeExpense {
			expenseAccMap[acc.ID] = acc.Name
		}
	}

	categoryTotal := make(map[string]decimal.Decimal)
	totalExpense := decimal.Zero

	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}

		for _, line := range journal.Lines {
			if name, ok := expenseAccMap[line.AccountID]; ok {
				// Expense is Debit balance.
				amount := line.DebitAmount.Sub(line.CreditAmount)
				if amount.IsPositive() {
					categoryTotal[name] = categoryTotal[name].Add(amount)
					totalExpense = totalExpense.Add(amount)
				}
			}
		}
	}

	var breakdown []entity.ExpenseBreakdownItem
	for cat, amount := range categoryTotal {
		percentage, _ := amount.Div(totalExpense).Float64()
		breakdown = append(breakdown, entity.ExpenseBreakdownItem{
			Category:   cat,
			Amount:     amount,
			Percentage: percentage * 100,
		})
	}

	// Sort by amount desc? Logic is simple map iteration now.
	// We can leave sorting to frontend or sort here if needed.
	// For simplicity, returning unsorted.

	return breakdown, nil
}

// GetCashFlowTrend Aggregates incoming/outgoing from Cash/Bank accounts
func (uc *ReportUsecase) GetCashFlowTrend(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.CashFlowTrendItem, error) {
	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, start, end)
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	cashAccMap := make(map[uuid.UUID]bool)
	for _, acc := range accounts {
		// Identify Cash/Bank accounts. Convention: Code starts with '1' and type is Asset?
		// Better checking Name or SubType if available.
		// Detailed logic: usually "Cas" or "Bank" in name, or specific AccountTypeAssetCash if we had it.
		// Re-using logic from GetCashFlow: Code[:1] == "1" && (Name=="Kas" || Name=="Bank") - simplified
		if len(acc.Code) > 0 && acc.Code[:1] == "1" && (acc.Name == "Kas" || acc.Name == "Bank") {
			cashAccMap[acc.ID] = true
		}
	}

	trendMap := make(map[string]*entity.CashFlowTrendItem)

	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		dateKey := journal.EntryDate.Format("2006-01-02")

		if _, ok := trendMap[dateKey]; !ok {
			trendMap[dateKey] = &entity.CashFlowTrendItem{Date: dateKey}
		}
		item := trendMap[dateKey]

		for _, line := range journal.Lines {
			if cashAccMap[line.AccountID] {
				// Debit = Incoming (Asset increase), Credit = Outgoing
				if line.DebitAmount.IsPositive() {
					item.Incoming = item.Incoming.Add(line.DebitAmount)
				}
				if line.CreditAmount.IsPositive() {
					item.Outgoing = item.Outgoing.Add(line.CreditAmount)
				}
			}
		}
	}

	var trends []entity.CashFlowTrendItem
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		if item, ok := trendMap[dateKey]; ok {
			item.NetChange = item.Incoming.Sub(item.Outgoing)
			trends = append(trends, *item)
		} else {
			trends = append(trends, entity.CashFlowTrendItem{Date: dateKey})
		}
	}

	return trends, nil
}

func (uc *ReportUsecase) GetDashboardStats(ctx context.Context, companyID uuid.UUID) (*entity.DashboardStats, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	// Correct end of month: first day of next month - 1 day
	endOfMonth = startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	incomeStmt, err := uc.GetIncomeStatement(ctx, companyID, startOfMonth, endOfMonth)
	if err != nil {
		return nil, common.WrapErr("get income statement", err)
	}

	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	activeAccounts := 0
	for _, acc := range accounts {
		if acc.IsActive {
			activeAccounts++
		}
	}

	journals, err := uc.journalRepo.GetByCompany(ctx, companyID, 5, 0, "")
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	// --- New Visual Data ---
	salesTrend, _ := uc.GetSalesTrend(ctx, companyID, startOfMonth, endOfMonth)
	expenseBreakdown, _ := uc.GetExpenseBreakdown(ctx, companyID, startOfMonth, endOfMonth)
	cashFlowTrend, _ := uc.GetCashFlowTrend(ctx, companyID, startOfMonth, endOfMonth)

	// Calculate growth (simple dummy logic for now as we don't fetch prev month yet)
	revenueGrowth := 0.0

	return &entity.DashboardStats{
		TotalRevenue:        incomeStmt.TotalRevenue,
		TotalExpenses:       incomeStmt.TotalExpenses,
		NetIncome:           incomeStmt.NetIncome,
		ActiveAccounts:      activeAccounts,
		RecentJournals:      journals,
		RevenueGrowth:       revenueGrowth,
		ActiveAccountGrowth: 0,
		SalesTrend:          salesTrend,
		ExpenseBreakdown:    expenseBreakdown,
		CashFlowTrend:       cashFlowTrend,
	}, nil
}
