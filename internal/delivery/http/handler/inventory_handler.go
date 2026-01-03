package handler

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/inventory"
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
		helper.BadRequest(c, err)
		return
	}

	txDate := time.Now()
	if req.TransactionDate != "" {
		txDate, _ = time.Parse("2006-01-02", req.TransactionDate)
	}

	result, err := h.usecase.StockIn(c.Request.Context(), inventory.StockInInput{
		CompanyID:       helper.GetCompanyID(c),
		TransactionNo:   req.TransactionNo,
		ProductID:       helper.ParseUUIDString(req.ProductID),
		WarehouseID:     helper.ParseUUIDString(req.WarehouseID),
		Quantity:        helper.ParseDecimal(req.Quantity),
		UnitCost:        helper.ParseDecimal(req.UnitCost),
		Reference:       req.Reference,
		Notes:           req.Notes,
		TransactionDate: txDate,
	})
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
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
		helper.BadRequest(c, err)
		return
	}

	txDate := time.Now()
	if req.TransactionDate != "" {
		txDate, _ = time.Parse("2006-01-02", req.TransactionDate)
	}

	result, err := h.usecase.StockOut(c.Request.Context(), inventory.StockOutInput{
		CompanyID:       helper.GetCompanyID(c),
		TransactionNo:   req.TransactionNo,
		ProductID:       helper.ParseUUIDString(req.ProductID),
		WarehouseID:     helper.ParseUUIDString(req.WarehouseID),
		Quantity:        helper.ParseDecimal(req.Quantity),
		Reference:       req.Reference,
		Notes:           req.Notes,
		TransactionDate: txDate,
	})
	if err != nil {
		if errors.Is(err, inventory.ErrInsufficientStock) {
			helper.ErrorMessage(c, 422, err.Error())
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

func (h *InventoryHandler) GetStock(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	productID := helper.ParseUUIDString(c.Query("product_id"))
	warehouseID := helper.ParseUUIDString(c.Query("warehouse_id"))

	if warehouseID != uuid.Nil {
		if stock, err := h.usecase.GetStock(c.Request.Context(), companyID, productID, warehouseID); err == nil {
			helper.Success(c, stock)
			return
		}
	}

	if productID != uuid.Nil {
		stocks, err := h.usecase.GetStockByProduct(c.Request.Context(), companyID, productID)
		if err != nil {
			helper.InternalError(c, err)
			return
		}
		helper.Success(c, stocks)
		return
	}

	// List All Stocks
	stocks, err := h.usecase.ListStocks(c.Request.Context(), companyID)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, stocks)
}

func (h *InventoryHandler) ListTransactions(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	p := helper.ParsePagination(c)

	filter := repository.InventoryFilter{
		Page:     p.Page,
		PageSize: p.PageSize,
	}

	if pid := c.Query("product_id"); pid != "" {
		id := helper.ParseUUIDString(pid)
		filter.ProductID = &id
	}
	if wid := c.Query("warehouse_id"); wid != "" {
		id := helper.ParseUUIDString(wid)
		filter.WarehouseID = &id
	}

	txs, total, err := h.usecase.ListTransactions(c.Request.Context(), companyID, filter)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.PaginatedItems(c, txs, total, p)
}
