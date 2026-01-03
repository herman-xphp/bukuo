package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
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
		helper.BadRequest(c, err)
		return
	}

	input := unit.CreateUnitInput{
		CompanyID: helper.GetCompanyID(c),
		Code:      req.Code,
		Name:      req.Name,
	}

	result, err := h.usecase.CreateUnit(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, unit.ErrUnitCodeExists) {
			helper.Conflict(c, err)
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// List handles GET /units
func (h *UnitHandler) List(c *gin.Context) {
	units, err := h.usecase.List(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, units)
}

// GetByID handles GET /units/:id
func (h *UnitHandler) GetByID(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "unit")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		helper.NotFound(c, "unit")
		return
	}

	helper.Success(c, result)
}

// Delete handles DELETE /units/:id
func (h *UnitHandler) Delete(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "unit")
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), companyID, id); err != nil {
		if errors.Is(err, unit.ErrUnitNotFound) {
			helper.NotFound(c, "unit")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Deleted(c, "unit")
}
