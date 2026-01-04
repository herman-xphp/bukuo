package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WidgetType represents the type of dashboard widget
type WidgetType string

const (
	WidgetTypeChart     WidgetType = "CHART"     // Bar, Line, Pie
	WidgetTypeMetric    WidgetType = "METRIC"    // Big Number (e.g. Total Sales)
	WidgetTypeList      WidgetType = "LIST"      // Recent Transactions
	WidgetTypeQuickLink WidgetType = "QUICKLINK" // Shortcuts
)

// WidgetDatasource represents the source of data for the widget
type WidgetDatasource string

const (
	DatasourceSales       WidgetDatasource = "SALES"
	DatasourceExpense     WidgetDatasource = "EXPENSE"
	DatasourceCashFlow    WidgetDatasource = "CASHFLOW"
	DatasourceBank        WidgetDatasource = "BANK"
	DatasourcePayables    WidgetDatasource = "PAYABLES"
	DatasourceReceivables WidgetDatasource = "RECEIVABLES"
)

// DashboardWidget represents a configured widget on a dashboard
type DashboardWidget struct {
	ID         uuid.UUID        `json:"id"`
	CompanyID  uuid.UUID        `json:"company_id"`
	UserID     uuid.UUID        `json:"user_id"` // Personalized dashboard
	Title      string           `json:"title"`
	Type       WidgetType       `json:"type"`
	Datasource WidgetDatasource `json:"datasource"`
	Config     string           `json:"config"` // JSON config (period, filters, etc)
	PositionX  int              `json:"position_x"`
	PositionY  int              `json:"position_y"`
	Width      int              `json:"width"`
	Height     int              `json:"height"`
	IsVisible  bool             `json:"is_visible"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// UserDashboardPreference represents dashboard preferences
type UserDashboardPreference struct {
	UserID          uuid.UUID `json:"user_id"`
	CompanyID       uuid.UUID `json:"company_id"`
	Theme           string    `json:"theme"`            // LIGHT, DARK
	LayoutType      string    `json:"layout_type"`      // DEFAULT, COMPACT
	RefreshInterval int       `json:"refresh_interval"` // Seconds
	UpdatedAt       time.Time `json:"updated_at"`
}

// SalesTrendItem represents a single data point for sales chart
type SalesTrendItem struct {
	Date   string          `json:"date"` // YYYY-MM-DD
	Amount decimal.Decimal `json:"amount"`
}

// ExpenseBreakdownItem represents a slice of expense pie chart
type ExpenseBreakdownItem struct {
	Category   string          `json:"category"`
	Amount     decimal.Decimal `json:"amount"`
	Percentage float64         `json:"percentage"`
}

// CashFlowTrendItem represents a single data point for cash flow chart
type CashFlowTrendItem struct {
	Date      string          `json:"date"` // YYYY-MM-DD
	Incoming  decimal.Decimal `json:"incoming"`
	Outgoing  decimal.Decimal `json:"outgoing"`
	NetChange decimal.Decimal `json:"net_change"`
}

// DashboardStats used in ReportUsecase
type DashboardStats struct {
	TotalRevenue        decimal.Decimal        `json:"total_revenue"`
	TotalExpenses       decimal.Decimal        `json:"total_expenses"`
	NetIncome           decimal.Decimal        `json:"net_income"`
	ActiveAccounts      int                    `json:"active_accounts"`
	RecentJournals      []JournalEntry         `json:"recent_journals"`
	RevenueGrowth       float64                `json:"revenue_growth"`
	ActiveAccountGrowth int                    `json:"active_account_growth"`
	SalesTrend          []SalesTrendItem       `json:"sales_trend"`
	ExpenseBreakdown    []ExpenseBreakdownItem `json:"expense_breakdown"`
	CashFlowTrend       []CashFlowTrendItem    `json:"cash_flow_trend"`
}
