package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	result, err := h.usecase.CreateWarehouse(c.Request.Context(), warehouse.CreateWarehouseInput{
		CompanyID: companyID, Code: req.Code, Name: req.Name, Address: req.Address,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, warehouse.ErrWarehouseCodeExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *WarehouseHandler) List(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	warehouses, err := h.usecase.List(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": warehouses})
}

func (h *WarehouseHandler) GetByID(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, _ := uuid.Parse(c.Param("id"))
	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "warehouse not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *WarehouseHandler) SetDefault(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.usecase.SetDefault(c.Request.Context(), companyID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "default warehouse updated"})
}

type UpdateWarehouseRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, _ := uuid.Parse(c.Param("id"))
	var req UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.usecase.Update(c.Request.Context(), companyID, id, req.Name, req.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *WarehouseHandler) Delete(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.usecase.Delete(c.Request.Context(), companyID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "warehouse deleted"})
}
