package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/usecase/unit"
)

// UnitHandler handles unit of measure endpoints
type UnitHandler struct {
	usecase *unit.UnitUsecase
}

// NewUnitHandler creates a new UnitHandler
func NewUnitHandler(uc *unit.UnitUsecase) *UnitHandler {
	return &UnitHandler{usecase: uc}
}

// CreateUnitRequest represents create unit request
type CreateUnitRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// Create handles POST /units
func (h *UnitHandler) Create(c *gin.Context) {
	var req CreateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))

	input := unit.CreateUnitInput{
		CompanyID: companyID,
		Code:      req.Code,
		Name:      req.Name,
	}

	result, err := h.usecase.CreateUnit(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, unit.ErrUnitCodeExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// List handles GET /units
func (h *UnitHandler) List(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	units, err := h.usecase.List(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": units})
}

// GetByID handles GET /units/:id
func (h *UnitHandler) GetByID(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete handles DELETE /units/:id
func (h *UnitHandler) Delete(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), companyID, id); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, unit.ErrUnitNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unit deleted"})
}
