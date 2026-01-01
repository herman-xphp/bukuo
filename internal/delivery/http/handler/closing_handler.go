package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/usecase/closing"
)

// ClosingHandler handles closing endpoints
type ClosingHandler struct {
	usecase *closing.ClosingUsecase
}

// NewClosingHandler creates a new ClosingHandler
func NewClosingHandler(uc *closing.ClosingUsecase) *ClosingHandler {
	return &ClosingHandler{usecase: uc}
}

// ClosePeriodRequest represents the request to close a period
type ClosePeriodRequest struct {
	PeriodID           string `json:"period_id" binding:"required"`
	RetainedEarningsID string `json:"retained_earnings_id" binding:"required"`
}

// ClosePeriod handles POST /closing/period
func (h *ClosingHandler) ClosePeriod(c *gin.Context) {
	var req ClosePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	periodID, err := uuid.Parse(req.PeriodID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_id"})
		return
	}

	retainedEarningsID, err := uuid.Parse(req.RetainedEarningsID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid retained_earnings_id"})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))
	userID, _ := uuid.Parse(c.GetString("user_id"))

	input := closing.ClosePeriodInput{
		PeriodID:           periodID,
		CompanyID:          companyID,
		RetainedEarningsID: retainedEarningsID,
		UserID:             userID,
	}

	result, err := h.usecase.ClosePeriod(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Period closed successfully",
		"net_income": result.NetIncome,
		"data":       result,
	})
}

// PreviewClosing handles GET /closing/preview/:period_id
func (h *ClosingHandler) PreviewClosing(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("period_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_id"})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))

	result, err := h.usecase.PreviewClosing(c.Request.Context(), periodID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}
