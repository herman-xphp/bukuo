package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/usecase/period"
)

// PeriodHandler handles period endpoints
type PeriodHandler struct {
	usecase *period.PeriodUsecase
}

// NewPeriodHandler creates a new PeriodHandler
func NewPeriodHandler(uc *period.PeriodUsecase) *PeriodHandler {
	return &PeriodHandler{usecase: uc}
}

// CreatePeriodRequest represents create period request
type CreatePeriodRequest struct {
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

// Create handles POST /periods
func (h *PeriodHandler) Create(c *gin.Context) {
	var req CreatePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	input := period.CreatePeriodInput{
		CompanyID: helper.GetCompanyID(c),
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := h.usecase.CreatePeriod(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// GetAll handles GET /periods
func (h *PeriodHandler) GetAll(c *gin.Context) {
	companyID := helper.GetCompanyID(c)

	// Check if pagination/search
	if c.Query("limit") != "" || c.Query("q") != "" || c.Query("offset") != "" || c.Query("page") != "" {
		p := helper.ParsePagination(c)
		search := c.Query("q")

		periods, total, err := h.usecase.List(c.Request.Context(), companyID, p.Limit, p.Offset, search)
		if err != nil {
			helper.InternalError(c, err)
			return
		}

		helper.PaginatedItems(c, periods, int64(total), p)
		return
	}

	// Default: Return Limitless (Backwards compatible)
	periods, err := h.usecase.GetByCompany(c.Request.Context(), companyID)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, periods)
}

// GetByID handles GET /periods/:id
func (h *PeriodHandler) GetByID(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "period")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		helper.NotFound(c, "period")
		return
	}

	helper.Success(c, result)
}

// Update handles PUT /periods/:id
func (h *PeriodHandler) Update(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "period")
		return
	}

	var req CreatePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	input := period.CreatePeriodInput{
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := h.usecase.UpdatePeriod(c.Request.Context(), id, input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// Delete handles DELETE /periods/:id
func (h *PeriodHandler) Delete(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "period")
		return
	}

	if err := h.usecase.DeletePeriod(c.Request.Context(), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Deleted(c, "period")
}

// Close handles POST /periods/:id/close
func (h *PeriodHandler) Close(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "period")
		return
	}

	userID := helper.GetUserID(c)

	result, err := h.usecase.ClosePeriod(c.Request.Context(), id, userID)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{
		"message": "period closed",
		"data":    result,
	})
}
