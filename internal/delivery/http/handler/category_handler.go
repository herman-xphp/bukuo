package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/usecase/category"
)

// CategoryHandler handles product category endpoints
type CategoryHandler struct {
	usecase *category.CategoryUsecase
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(uc *category.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{usecase: uc}
}

// CreateCategoryRequest represents create category request
type CreateCategoryRequest struct {
	Name     string  `json:"name" binding:"required"`
	ParentID *string `json:"parent_id"`
}

// Create handles POST /product-categories
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	input := category.CreateCategoryInput{
		CompanyID: helper.GetCompanyID(c),
		Name:      req.Name,
		ParentID:  helper.ParseOptionalUUID(req.ParentID),
	}

	result, err := h.usecase.CreateCategory(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// List handles GET /product-categories
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.usecase.List(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, categories)
}

// GetByID handles GET /product-categories/:id
func (h *CategoryHandler) GetByID(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "category")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		helper.NotFound(c, "category")
		return
	}

	helper.Success(c, result)
}

// UpdateCategoryRequest represents update category request
type UpdateCategoryRequest struct {
	Name     string  `json:"name" binding:"required"`
	ParentID *string `json:"parent_id"`
}

// Update handles PUT /product-categories/:id
func (h *CategoryHandler) Update(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "category")
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	input := category.UpdateCategoryInput{
		CompanyID: companyID,
		ID:        id,
		Name:      req.Name,
		ParentID:  helper.ParseOptionalUUID(req.ParentID),
	}

	result, err := h.usecase.UpdateCategory(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			helper.NotFound(c, "category")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// Delete handles DELETE /product-categories/:id
func (h *CategoryHandler) Delete(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "category")
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), companyID, id); err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			helper.NotFound(c, "category")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Deleted(c, "category")
}
