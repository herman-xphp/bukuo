package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/usecase/report"
)

// ReportHandler handles report endpoints
type ReportHandler struct {
	usecase *report.ReportUsecase
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(uc *report.ReportUsecase) *ReportHandler {
	return &ReportHandler{usecase: uc}
}

// TrialBalance handles GET /reports/trial-balance
func (h *ReportHandler) TrialBalance(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	periodID := helper.ParseUUIDString(c.Query("period_id"))

	if periodID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.BadRequestMessage(c, "period_id required")
		return
	}

	result, err := h.usecase.GetTrialBalance(c.Request.Context(), companyID, periodID)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// GeneralLedger handles GET /reports/ledger/:account_id
func (h *ReportHandler) GeneralLedger(c *gin.Context) {
	accountID, err := helper.ParseUUID(c, "account_id")
	if err != nil {
		helper.InvalidID(c, "account")
		return
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	if start.IsZero() || end.IsZero() {
		helper.BadRequestMessage(c, "start_date and end_date required")
		return
	}

	result, err := h.usecase.GetGeneralLedger(c.Request.Context(), accountID, start, end)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// IncomeStatement handles GET /reports/income-statement
func (h *ReportHandler) IncomeStatement(c *gin.Context) {
	companyID := helper.GetCompanyID(c)

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	if start.IsZero() || end.IsZero() {
		helper.BadRequestMessage(c, "start_date and end_date required")
		return
	}

	result, err := h.usecase.GetIncomeStatement(c.Request.Context(), companyID, start, end)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// BalanceSheet handles GET /reports/balance-sheet
func (h *ReportHandler) BalanceSheet(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	dateStr := c.Query("as_of_date")

	asOfDate, _ := time.Parse("2006-01-02", dateStr)
	if asOfDate.IsZero() {
		asOfDate = time.Now()
	}

	result, err := h.usecase.GetBalanceSheet(c.Request.Context(), companyID, asOfDate)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// CashFlow handles GET /reports/cash-flow
func (h *ReportHandler) CashFlow(c *gin.Context) {
	companyID := helper.GetCompanyID(c)

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	if start.IsZero() || end.IsZero() {
		helper.BadRequestMessage(c, "start_date and end_date required")
		return
	}

	result, err := h.usecase.GetCashFlow(c.Request.Context(), companyID, start, end)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// Dashboard handles GET /reports/dashboard
func (h *ReportHandler) Dashboard(c *gin.Context) {
	result, err := h.usecase.GetDashboardStats(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}
