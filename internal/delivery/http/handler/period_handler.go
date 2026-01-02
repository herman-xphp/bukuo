package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	input := period.CreatePeriodInput{
		CompanyID: companyID,
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := h.usecase.CreatePeriod(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// GetAll handles GET /periods
func (h *PeriodHandler) GetAll(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	// Check if pagination/search
	if c.Query("limit") != "" || c.Query("q") != "" || c.Query("offset") != "" || c.Query("page") != "" {
		limit := 50
		page := 1
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}
		if p := c.Query("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil {
				page = parsed
			}
		}
		offset := (page - 1) * limit
		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}
		search := c.Query("q")

		periods, total, err := h.usecase.List(c.Request.Context(), companyID, limit, offset, search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"items":  periods,
				"total":  total,
				"limit":  limit,
				"page":   page,
				"offset": offset,
			},
		})
		return
	}

	// Default: Return Limitless (Backwards compatible)
	periods, err := h.usecase.GetByCompany(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": periods})
}

// GetByID handles GET /periods/:id
func (h *PeriodHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "period not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Update handles PUT /periods/:id
func (h *PeriodHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period id"})
		return
	}

	var req CreatePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete handles DELETE /periods/:id
func (h *PeriodHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period id"})
		return
	}

	if err := h.usecase.DeletePeriod(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "period deleted"})
}

// Close handles POST /periods/:id/close
func (h *PeriodHandler) Close(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period id"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))

	result, err := h.usecase.ClosePeriod(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "period closed",
		"data":    result,
	})
}
