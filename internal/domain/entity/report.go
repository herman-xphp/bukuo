package entity

import (
	"time"

	"github.com/google/uuid"
)

// ReportType represents the type of report
type ReportType string

const (
	ReportTypeFinancial  ReportType = "FINANCIAL"  // BS, PL, CF
	ReportTypeSales      ReportType = "SALES"      // Sales Summary
	ReportTypePurchasing ReportType = "PURCHASING" // Purchase Summary
	ReportTypeInventory  ReportType = "INVENTORY"  // Stock Card
	ReportTypeTax        ReportType = "TAX"        // Tax Summary
)

// ReportTemplate represents a saved report configuration
type ReportTemplate struct {
	ID          uuid.UUID  `json:"id"`
	CompanyID   uuid.UUID  `json:"company_id"`
	Name        string     `json:"name"`
	Type        ReportType `json:"type"`
	Code        string     `json:"code"` // e.g. "BS-STD" for Standard Balance Sheet
	Description string     `json:"description,omitempty"`
	Config      string     `json:"config"` // JSON string of parameters (date range, filters, etc)
	CreatedBy   uuid.UUID  `json:"created_by"`
	IsPublic    bool       `json:"is_public"` // Visible to all users in company
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ReportSchedule represents a schedule for auto-generating/emailing reports
type ReportSchedule struct {
	ID              uuid.UUID `json:"id"`
	CompanyID       uuid.UUID `json:"company_id"`
	TemplateID      uuid.UUID `json:"template_id"`
	Frequency       string    `json:"frequency"` // DAILY, WEEKLY, MONTHLY
	NextRunAt       time.Time `json:"next_run_at"`
	EmailRecipients string    `json:"email_recipients"` // Comma separated emails
	Format          string    `json:"format"`           // PDF, EXCEL, CSV
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GeneratedReport represents an archived report outcome
type GeneratedReport struct {
	ID           uuid.UUID  `json:"id"`
	CompanyID    uuid.UUID  `json:"company_id"`
	TemplateID   *uuid.UUID `json:"template_id,omitempty"` // If generated from template
	ScheduleID   *uuid.UUID `json:"schedule_id,omitempty"` // If generated from schedule
	GeneratedAt  time.Time  `json:"generated_at"`
	GeneratedBy  *uuid.UUID `json:"generated_by,omitempty"` // Null if system
	FileName     string     `json:"file_name"`
	FileUrl      string     `json:"file_url"`
	ReportParams string     `json:"report_params"` // JSON snapshot of params used
}
