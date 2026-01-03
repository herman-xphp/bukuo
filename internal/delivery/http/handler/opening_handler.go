package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/usecase/opening"
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
		helper.BadRequest(c, err)
		return
	}

	periodID := helper.ParseUUIDString(req.PeriodID)
	if periodID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.BadRequestMessage(c, "invalid period_id")
		return
	}

	balanceDate, err := time.Parse("2006-01-02", req.BalanceDate)
	if err != nil {
		helper.BadRequestMessage(c, "invalid balance_date, use YYYY-MM-DD")
		return
	}

	// Parse balances
	balances := make([]opening.OpeningBalanceInput, 0, len(req.Balances))
	for _, b := range req.Balances {
		accountID := helper.ParseUUIDString(b.AccountID)
		if accountID.String() == "00000000-0000-0000-0000-000000000000" {
			helper.BadRequestMessage(c, "invalid account_id: "+b.AccountID)
			return
		}

		balance := helper.ParseDecimal(b.Balance)

		balances = append(balances, opening.OpeningBalanceInput{
			AccountID: accountID,
			Balance:   balance,
		})
	}

	input := opening.ImportOpeningBalanceInput{
		CompanyID:   helper.GetCompanyID(c),
		PeriodID:    periodID,
		UserID:      helper.GetUserID(c),
		BalanceDate: balanceDate,
		Balances:    balances,
	}

	result, err := h.usecase.ImportOpeningBalance(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, gin.H{
		"message":       "Opening balance imported successfully",
		"total_debit":   result.TotalDebit,
		"total_credit":  result.TotalCredit,
		"account_count": result.AccountCount,
		"data":          result,
	})
}

// Template handles GET /opening-balance/template
func (h *OpeningHandler) Template(c *gin.Context) {
	result, err := h.usecase.GetOpeningBalanceTemplate(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, gin.H{
		"data":  result,
		"count": len(result),
	})
}
