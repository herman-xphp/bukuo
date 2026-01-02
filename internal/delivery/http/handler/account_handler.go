package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/account"
)

// AccountHandler handles account endpoints
type AccountHandler struct {
	usecase *account.AccountUsecase
}

// NewAccountHandler creates a new AccountHandler
func NewAccountHandler(uc *account.AccountUsecase) *AccountHandler {
	return &AccountHandler{usecase: uc}
}

// CreateAccountRequest represents create account request
type CreateAccountRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	ParentID    *string `json:"parent_id"`
	IsPostable  bool    `json:"is_postable"`
	Description string  `json:"description"`
}

// Create handles POST /accounts
func (h *AccountHandler) Create(c *gin.Context) {
	var req CreateAccountRequest
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

	input := account.CreateAccountInput{
		CompanyID:   companyID,
		Code:        req.Code,
		Name:        req.Name,
		Type:        entity.AccountType(req.Type),
		ParentID:    parentID,
		IsPostable:  req.IsPostable,
		Description: req.Description,
	}

	result, err := h.usecase.CreateAccount(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// GetAll handles GET /accounts
func (h *AccountHandler) GetAll(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	// Check if pagination/search is requested
	if c.Query("limit") != "" || c.Query("q") != "" || c.Query("offset") != "" || c.Query("page") != "" {
		limit := 50
		page := 1
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}
		if p := c.Query("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil {
				page = parsed
			}
		}
		offset := (page - 1) * limit
		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}
		search := c.Query("q")

		result, total, err := h.usecase.List(c.Request.Context(), companyID, limit, offset, search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"items":  result,
				"total":  total,
				"limit":  limit,
				"page":   page,
				"offset": offset,
			},
		})
		return
	}

	// Default: Return Limitless (Existing behavior)
	accounts, err := h.usecase.GetAccountsWithBalances(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": accounts})
}

// GetByID handles GET /accounts/:id
func (h *AccountHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateAccountRequest represents update account request
type UpdateAccountRequest struct {
	Name        string  `json:"name" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	ParentID    *string `json:"parent_id"`
	IsPostable  bool    `json:"is_postable"`
	IsActive    bool    `json:"is_active"`
	Description string  `json:"description"`
}

// Update handles PUT /accounts/:id
func (h *AccountHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, _ := uuid.Parse(*req.ParentID)
		parentID = &pid
	}

	input := account.UpdateAccountInput{
		ID:          id,
		Name:        req.Name,
		Type:        entity.AccountType(req.Type),
		ParentID:    parentID,
		IsPostable:  req.IsPostable,
		IsActive:    req.IsActive,
		Description: req.Description,
	}

	result, err := h.usecase.UpdateAccount(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete handles DELETE /accounts/:id
func (h *AccountHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	if err := h.usecase.DeleteAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
