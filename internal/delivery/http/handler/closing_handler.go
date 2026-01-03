package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
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
		helper.BadRequest(c, err)
		return
	}

	periodID := helper.ParseUUIDString(req.PeriodID)
	if periodID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.BadRequestMessage(c, "invalid period_id")
		return
	}

	retainedEarningsID := helper.ParseUUIDString(req.RetainedEarningsID)
	if retainedEarningsID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.BadRequestMessage(c, "invalid retained_earnings_id")
		return
	}

	input := closing.ClosePeriodInput{
		PeriodID:           periodID,
		CompanyID:          helper.GetCompanyID(c),
		RetainedEarningsID: retainedEarningsID,
		UserID:             helper.GetUserID(c),
	}

	result, err := h.usecase.ClosePeriod(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{
		"message":    "Period closed successfully",
		"net_income": result.NetIncome,
		"data":       result,
	})
}

// PreviewClosing handles GET /closing/preview/:period_id
func (h *ClosingHandler) PreviewClosing(c *gin.Context) {
	periodID, err := helper.ParseUUID(c, "period_id")
	if err != nil {
		helper.InvalidID(c, "period")
		return
	}

	result, err := h.usecase.PreviewClosing(c.Request.Context(), periodID, helper.GetCompanyID(c))
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, result)
}
