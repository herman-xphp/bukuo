package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/product"
	"github.com/shopspring/decimal"
)

// ProductHandler handles product endpoints
type ProductHandler struct {
	usecase *product.ProductUsecase
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(uc *product.ProductUsecase) *ProductHandler {
	return &ProductHandler{usecase: uc}
}

// CreateProductRequest represents create product request
type CreateProductRequest struct {
	Code               string  `json:"code" binding:"required"`
	Name               string  `json:"name" binding:"required"`
	Type               string  `json:"type" binding:"required"`
	CategoryID         *string `json:"category_id"`
	UnitID             string  `json:"unit_id" binding:"required"`
	Description        string  `json:"description"`
	SalesPrice         string  `json:"sales_price"`
	PurchasePrice      string  `json:"purchase_price"`
	SalesAccountID     string  `json:"sales_account_id" binding:"required"`
	PurchaseAccountID  string  `json:"purchase_account_id" binding:"required"`
	InventoryAccountID *string `json:"inventory_account_id"`
	MinStock           string  `json:"min_stock"`
}

// Create handles POST /products
func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))
	unitID, _ := uuid.Parse(req.UnitID)
	salesAccountID, _ := uuid.Parse(req.SalesAccountID)
	purchaseAccountID, _ := uuid.Parse(req.PurchaseAccountID)

	var categoryID *uuid.UUID
	if req.CategoryID != nil {
		id, _ := uuid.Parse(*req.CategoryID)
		categoryID = &id
	}

	var inventoryAccountID *uuid.UUID
	if req.InventoryAccountID != nil {
		id, _ := uuid.Parse(*req.InventoryAccountID)
		inventoryAccountID = &id
	}

	salesPrice := decimal.Zero
	if req.SalesPrice != "" {
		sp, _ := decimal.NewFromString(req.SalesPrice)
		salesPrice = sp
	}

	purchasePrice := decimal.Zero
	if req.PurchasePrice != "" {
		pp, _ := decimal.NewFromString(req.PurchasePrice)
		purchasePrice = pp
	}

	minStock := decimal.Zero
	if req.MinStock != "" {
		ms, _ := decimal.NewFromString(req.MinStock)
		minStock = ms
	}

	input := product.CreateProductInput{
		CompanyID:          companyID,
		Code:               req.Code,
		Name:               req.Name,
		Type:               entity.ProductType(req.Type),
		CategoryID:         categoryID,
		UnitID:             unitID,
		Description:        req.Description,
		SalesPrice:         salesPrice,
		PurchasePrice:      purchasePrice,
		SalesAccountID:     salesAccountID,
		PurchaseAccountID:  purchaseAccountID,
		InventoryAccountID: inventoryAccountID,
		MinStock:           minStock,
	}

	result, err := h.usecase.CreateProduct(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, product.ErrProductCodeExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// List handles GET /products
func (h *ProductHandler) List(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil {
			pageSize = parsed
		}
	}

	input := product.ListInput{
		CompanyID: companyID,
		Search:    c.Query("q"),
		Page:      page,
		PageSize:  pageSize,
	}

	// Filter by product type
	if t := c.Query("type"); t != "" {
		productType := entity.ProductType(t)
		input.ProductType = &productType
	}

	// Filter by category
	if cat := c.Query("category_id"); cat != "" {
		catID, _ := uuid.Parse(cat)
		input.CategoryID = &catID
	}

	// Filter by active status
	if a := c.Query("active"); a != "" {
		active := a == "true"
		input.IsActive = &active
	}

	result, err := h.usecase.List(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"items":       result.Products,
			"total":       result.Total,
			"page":        result.Page,
			"page_size":   result.PageSize,
			"total_pages": result.TotalPages,
		},
	})
}

// GetByID handles GET /products/:id
func (h *ProductHandler) GetByID(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateProductRequest represents update product request
type UpdateProductRequest struct {
	Name               string  `json:"name" binding:"required"`
	Type               string  `json:"type" binding:"required"`
	CategoryID         *string `json:"category_id"`
	UnitID             string  `json:"unit_id" binding:"required"`
	Description        string  `json:"description"`
	SalesPrice         string  `json:"sales_price"`
	PurchasePrice      string  `json:"purchase_price"`
	SalesAccountID     string  `json:"sales_account_id" binding:"required"`
	PurchaseAccountID  string  `json:"purchase_account_id" binding:"required"`
	InventoryAccountID *string `json:"inventory_account_id"`
	MinStock           string  `json:"min_stock"`
	IsActive           bool    `json:"is_active"`
}

// Update handles PUT /products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unitID, _ := uuid.Parse(req.UnitID)
	salesAccountID, _ := uuid.Parse(req.SalesAccountID)
	purchaseAccountID, _ := uuid.Parse(req.PurchaseAccountID)

	var categoryID *uuid.UUID
	if req.CategoryID != nil {
		catID, _ := uuid.Parse(*req.CategoryID)
		categoryID = &catID
	}

	var inventoryAccountID *uuid.UUID
	if req.InventoryAccountID != nil {
		invID, _ := uuid.Parse(*req.InventoryAccountID)
		inventoryAccountID = &invID
	}

	salesPrice := decimal.Zero
	if req.SalesPrice != "" {
		sp, _ := decimal.NewFromString(req.SalesPrice)
		salesPrice = sp
	}

	purchasePrice := decimal.Zero
	if req.PurchasePrice != "" {
		pp, _ := decimal.NewFromString(req.PurchasePrice)
		purchasePrice = pp
	}

	minStock := decimal.Zero
	if req.MinStock != "" {
		ms, _ := decimal.NewFromString(req.MinStock)
		minStock = ms
	}

	input := product.UpdateProductInput{
		CompanyID:          companyID,
		ID:                 id,
		Name:               req.Name,
		Type:               entity.ProductType(req.Type),
		CategoryID:         categoryID,
		UnitID:             unitID,
		Description:        req.Description,
		SalesPrice:         salesPrice,
		PurchasePrice:      purchasePrice,
		SalesAccountID:     salesAccountID,
		PurchaseAccountID:  purchaseAccountID,
		InventoryAccountID: inventoryAccountID,
		MinStock:           minStock,
		IsActive:           req.IsActive,
	}

	result, err := h.usecase.UpdateProduct(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, product.ErrProductNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete handles DELETE /products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.usecase.DeleteProduct(c.Request.Context(), companyID, id); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, product.ErrProductNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
