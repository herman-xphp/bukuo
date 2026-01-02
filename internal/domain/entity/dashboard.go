package entity

import (
	"time"

	"github.com/google/uuid"
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
