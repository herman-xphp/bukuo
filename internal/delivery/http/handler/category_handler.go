package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))

	var parentID *uuid.UUID
	if req.ParentID != nil {
		id, _ := uuid.Parse(*req.ParentID)
		parentID = &id
	}

	input := category.CreateCategoryInput{
		CompanyID: companyID,
		Name:      req.Name,
		ParentID:  parentID,
	}

	result, err := h.usecase.CreateCategory(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// List handles GET /product-categories
func (h *CategoryHandler) List(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	categories, err := h.usecase.List(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// GetByID handles GET /product-categories/:id
func (h *CategoryHandler) GetByID(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateCategoryRequest represents update category request
type UpdateCategoryRequest struct {
	Name     string  `json:"name" binding:"required"`
	ParentID *string `json:"parent_id"`
}

// Update handles PUT /product-categories/:id
func (h *CategoryHandler) Update(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, _ := uuid.Parse(*req.ParentID)
		parentID = &pid
	}

	input := category.UpdateCategoryInput{
		CompanyID: companyID,
		ID:        id,
		Name:      req.Name,
		ParentID:  parentID,
	}

	result, err := h.usecase.UpdateCategory(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, category.ErrCategoryNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete handles DELETE /product-categories/:id
func (h *CategoryHandler) Delete(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), companyID, id); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, category.ErrCategoryNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}
