package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/usecase/opening"
	"github.com/shopspring/decimal"
)

// OpeningHandler handles opening balance endpoints
type OpeningHandler struct {
	usecase *opening.OpeningBalanceUsecase
}

// NewOpeningHandler creates a new OpeningHandler
func NewOpeningHandler(uc *opening.OpeningBalanceUsecase) *OpeningHandler {
	return &OpeningHandler{usecase: uc}
}

// ImportOpeningBalanceRequest represents the import request
type ImportOpeningBalanceRequest struct {
	PeriodID    string                      `json:"period_id" binding:"required"`
	BalanceDate string                      `json:"balance_date" binding:"required"`
	Balances    []OpeningBalanceItemRequest `json:"balances" binding:"required,min=1"`
}

// OpeningBalanceItemRequest represents a single balance item
type OpeningBalanceItemRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Balance   string `json:"balance" binding:"required"`
}

// Import handles POST /opening-balance/import
func (h *OpeningHandler) Import(c *gin.Context) {
	var req ImportOpeningBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	periodID, err := uuid.Parse(req.PeriodID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_id"})
		return
	}

	balanceDate, err := time.Parse("2006-01-02", req.BalanceDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid balance_date, use YYYY-MM-DD"})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))
	userID, _ := uuid.Parse(c.GetString("user_id"))

	// Parse balances
	balances := make([]opening.OpeningBalanceInput, 0, len(req.Balances))
	for _, b := range req.Balances {
		accountID, err := uuid.Parse(b.AccountID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id: " + b.AccountID})
			return
		}

		balance, err := decimal.NewFromString(b.Balance)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid balance value: " + b.Balance})
			return
		}

		balances = append(balances, opening.OpeningBalanceInput{
			AccountID: accountID,
			Balance:   balance,
		})
	}

	input := opening.ImportOpeningBalanceInput{
		CompanyID:   companyID,
		PeriodID:    periodID,
		UserID:      userID,
		BalanceDate: balanceDate,
		Balances:    balances,
	}

	result, err := h.usecase.ImportOpeningBalance(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Opening balance imported successfully",
		"total_debit":   result.TotalDebit,
		"total_credit":  result.TotalCredit,
		"account_count": result.AccountCount,
		"data":          result,
	})
}

// Template handles GET /opening-balance/template
func (h *OpeningHandler) Template(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	result, err := h.usecase.GetOpeningBalanceTemplate(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"count": len(result),
	})
}
