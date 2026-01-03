package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/product"
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
		helper.BadRequest(c, err)
		return
	}

	input := product.CreateProductInput{
		CompanyID:          helper.GetCompanyID(c),
		Code:               req.Code,
		Name:               req.Name,
		Type:               entity.ProductType(req.Type),
		CategoryID:         helper.ParseOptionalUUID(req.CategoryID),
		UnitID:             helper.ParseUUIDString(req.UnitID),
		Description:        req.Description,
		SalesPrice:         helper.ParseDecimal(req.SalesPrice),
		PurchasePrice:      helper.ParseDecimal(req.PurchasePrice),
		SalesAccountID:     helper.ParseUUIDString(req.SalesAccountID),
		PurchaseAccountID:  helper.ParseUUIDString(req.PurchaseAccountID),
		InventoryAccountID: helper.ParseOptionalUUID(req.InventoryAccountID),
		MinStock:           helper.ParseDecimal(req.MinStock),
	}

	result, err := h.usecase.CreateProduct(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, product.ErrProductCodeExists) {
			helper.Conflict(c, err)
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// List handles GET /products
func (h *ProductHandler) List(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	p := helper.ParsePagination(c)

	input := product.ListInput{
		CompanyID: companyID,
		Search:    c.Query("q"),
		Page:      p.Page,
		PageSize:  p.PageSize,
	}

	// Filter by product type
	if t := c.Query("type"); t != "" {
		productType := entity.ProductType(t)
		input.ProductType = &productType
	}

	// Filter by category
	if cat := c.Query("category_id"); cat != "" {
		catID := helper.ParseUUIDString(cat)
		input.CategoryID = &catID
	}

	// Filter by active status
	input.IsActive = helper.ParseQueryBool(c, "active")

	result, err := h.usecase.List(c.Request.Context(), input)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.PaginatedItems(c, result.Products, int64(result.Total), p)
}

// GetByID handles GET /products/:id
func (h *ProductHandler) GetByID(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "product")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		helper.NotFound(c, "product")
		return
	}

	helper.Success(c, result)
}

// UpdateProductRequest represents update product request
type UpdateProductRequest struct {
	Name               string  `json:"name" binding:"required"`
	Type               string  `json:"type" binding:"required"`
	CategoryID         *string `json:"category_id"`
	UnitID             string  `json:"unit_id" binding:"required"`
	Description        string  `json:"description"`
	ImageURL           *string `json:"image_url"`
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
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "product")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	input := product.UpdateProductInput{
		CompanyID:          companyID,
		ID:                 id,
		Name:               req.Name,
		Type:               entity.ProductType(req.Type),
		CategoryID:         helper.ParseOptionalUUID(req.CategoryID),
		UnitID:             helper.ParseUUIDString(req.UnitID),
		Description:        req.Description,
		ImageURL:           req.ImageURL,
		SalesPrice:         helper.ParseDecimal(req.SalesPrice),
		PurchasePrice:      helper.ParseDecimal(req.PurchasePrice),
		SalesAccountID:     helper.ParseUUIDString(req.SalesAccountID),
		PurchaseAccountID:  helper.ParseUUIDString(req.PurchaseAccountID),
		InventoryAccountID: helper.ParseOptionalUUID(req.InventoryAccountID),
		MinStock:           helper.ParseDecimal(req.MinStock),
		IsActive:           req.IsActive,
	}

	result, err := h.usecase.UpdateProduct(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			helper.NotFound(c, "product")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// Delete handles DELETE /products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "product")
		return
	}

	if err := h.usecase.DeleteProduct(c.Request.Context(), companyID, id); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			helper.NotFound(c, "product")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Deleted(c, "product")
}
