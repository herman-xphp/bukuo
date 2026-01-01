package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	periodID, err := uuid.Parse(c.Query("period_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_id required"})
		return
	}

	result, err := h.usecase.GetTrialBalance(c.Request.Context(), companyID, periodID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GeneralLedger handles GET /reports/ledger/:account_id
func (h *ReportHandler) GeneralLedger(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("account_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
		return
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	if start.IsZero() || end.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date required"})
		return
	}

	result, err := h.usecase.GetGeneralLedger(c.Request.Context(), accountID, start, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// IncomeStatement handles GET /reports/income-statement
func (h *ReportHandler) IncomeStatement(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	if start.IsZero() || end.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date required"})
		return
	}

	result, err := h.usecase.GetIncomeStatement(c.Request.Context(), companyID, start, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// BalanceSheet handles GET /reports/balance-sheet
func (h *ReportHandler) BalanceSheet(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	dateStr := c.Query("as_of_date")

	asOfDate, _ := time.Parse("2006-01-02", dateStr)
	if asOfDate.IsZero() {
		asOfDate = time.Now()
	}

	result, err := h.usecase.GetBalanceSheet(c.Request.Context(), companyID, asOfDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// CashFlow handles GET /reports/cash-flow
func (h *ReportHandler) CashFlow(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	if start.IsZero() || end.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date required"})
		return
	}

	result, err := h.usecase.GetCashFlow(c.Request.Context(), companyID, start, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}
