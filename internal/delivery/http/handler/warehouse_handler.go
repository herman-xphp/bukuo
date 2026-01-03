package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/usecase/warehouse"
)

type WarehouseHandler struct {
	usecase *warehouse.WarehouseUsecase
}

func NewWarehouseHandler(uc *warehouse.WarehouseUsecase) *WarehouseHandler {
	return &WarehouseHandler{usecase: uc}
}

type CreateWarehouseRequest struct {
	Code    string `json:"code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	var req CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	result, err := h.usecase.CreateWarehouse(c.Request.Context(), warehouse.CreateWarehouseInput{
		CompanyID: helper.GetCompanyID(c),
		Code:      req.Code,
		Name:      req.Name,
		Address:   req.Address,
	})
	if err != nil {
		if errors.Is(err, warehouse.ErrWarehouseCodeExists) {
			helper.Conflict(c, err)
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

func (h *WarehouseHandler) List(c *gin.Context) {
	warehouses, err := h.usecase.List(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, warehouses)
}

func (h *WarehouseHandler) GetByID(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "warehouse")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		helper.NotFound(c, "warehouse")
		return
	}

	helper.Success(c, result)
}

func (h *WarehouseHandler) SetDefault(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "warehouse")
		return
	}

	if err := h.usecase.SetDefault(c.Request.Context(), companyID, id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Message(c, "default warehouse updated")
}

type UpdateWarehouseRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "warehouse")
		return
	}

	var req UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	result, err := h.usecase.Update(c.Request.Context(), companyID, id, req.Name, req.Address)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

func (h *WarehouseHandler) Delete(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "warehouse")
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), companyID, id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Deleted(c, "warehouse")
}
