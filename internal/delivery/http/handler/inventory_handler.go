package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/inventory"
	"github.com/shopspring/decimal"
)

type InventoryHandler struct {
	usecase *inventory.InventoryUsecase
}

func NewInventoryHandler(uc *inventory.InventoryUsecase) *InventoryHandler {
	return &InventoryHandler{usecase: uc}
}

type StockInRequest struct {
	TransactionNo   string `json:"transaction_no" binding:"required"`
	ProductID       string `json:"product_id" binding:"required"`
	WarehouseID     string `json:"warehouse_id" binding:"required"`
	Quantity        string `json:"quantity" binding:"required"`
	UnitCost        string `json:"unit_cost"`
	Reference       string `json:"reference"`
	Notes           string `json:"notes"`
	TransactionDate string `json:"transaction_date"`
}

func (h *InventoryHandler) StockIn(c *gin.Context) {
	var req StockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	productID, _ := uuid.Parse(req.ProductID)
	warehouseID, _ := uuid.Parse(req.WarehouseID)
	qty, _ := decimal.NewFromString(req.Quantity)
	cost, _ := decimal.NewFromString(req.UnitCost)
	txDate := time.Now()
	if req.TransactionDate != "" {
		txDate, _ = time.Parse("2006-01-02", req.TransactionDate)
	}
	result, err := h.usecase.StockIn(c.Request.Context(), inventory.StockInInput{
		CompanyID: companyID, TransactionNo: req.TransactionNo, ProductID: productID, WarehouseID: warehouseID,
		Quantity: qty, UnitCost: cost, Reference: req.Reference, Notes: req.Notes, TransactionDate: txDate,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

type StockOutRequest struct {
	TransactionNo   string `json:"transaction_no" binding:"required"`
	ProductID       string `json:"product_id" binding:"required"`
	WarehouseID     string `json:"warehouse_id" binding:"required"`
	Quantity        string `json:"quantity" binding:"required"`
	Reference       string `json:"reference"`
	Notes           string `json:"notes"`
	TransactionDate string `json:"transaction_date"`
}

func (h *InventoryHandler) StockOut(c *gin.Context) {
	var req StockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	productID, _ := uuid.Parse(req.ProductID)
	warehouseID, _ := uuid.Parse(req.WarehouseID)
	qty, _ := decimal.NewFromString(req.Quantity)
	txDate := time.Now()
	if req.TransactionDate != "" {
		txDate, _ = time.Parse("2006-01-02", req.TransactionDate)
	}
	result, err := h.usecase.StockOut(c.Request.Context(), inventory.StockOutInput{
		CompanyID: companyID, TransactionNo: req.TransactionNo, ProductID: productID, WarehouseID: warehouseID,
		Quantity: qty, Reference: req.Reference, Notes: req.Notes, TransactionDate: txDate,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, inventory.ErrInsufficientStock) {
			status = http.StatusUnprocessableEntity
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *InventoryHandler) GetStock(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	productID, _ := uuid.Parse(c.Query("product_id"))
	warehouseID, _ := uuid.Parse(c.Query("warehouse_id"))
	if warehouseID != uuid.Nil {
		stock, err := h.usecase.GetStock(c.Request.Context(), companyID, productID, warehouseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": stock})
		return
	}
	stocks, err := h.usecase.GetStockByProduct(c.Request.Context(), companyID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stocks})
}

func (h *InventoryHandler) ListTransactions(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	filter := repository.InventoryFilter{Page: 1, PageSize: 20}
	if p := c.Query("page"); p != "" {
		filter.Page, _ = strconv.Atoi(p)
	}
	if ps := c.Query("page_size"); ps != "" {
		filter.PageSize, _ = strconv.Atoi(ps)
	}
	if pid := c.Query("product_id"); pid != "" {
		id, _ := uuid.Parse(pid)
		filter.ProductID = &id
	}
	if wid := c.Query("warehouse_id"); wid != "" {
		id, _ := uuid.Parse(wid)
		filter.WarehouseID = &id
	}
	txs, total, err := h.usecase.ListTransactions(c.Request.Context(), companyID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"items": txs, "total": total, "page": filter.Page, "page_size": filter.PageSize}})
}
