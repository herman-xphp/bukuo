package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
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
		helper.BadRequest(c, err)
		return
	}

	input := account.CreateAccountInput{
		CompanyID:   helper.GetCompanyID(c),
		Code:        req.Code,
		Name:        req.Name,
		Type:        entity.AccountType(req.Type),
		ParentID:    helper.ParseOptionalUUID(req.ParentID),
		IsPostable:  req.IsPostable,
		Description: req.Description,
	}

	result, err := h.usecase.CreateAccount(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// GetAll handles GET /accounts
func (h *AccountHandler) GetAll(c *gin.Context) {
	companyID := helper.GetCompanyID(c)

	// Check if pagination/search is requested
	if c.Query("limit") != "" || c.Query("q") != "" || c.Query("offset") != "" || c.Query("page") != "" {
		p := helper.ParsePagination(c)
		search := c.Query("q")

		result, total, err := h.usecase.List(c.Request.Context(), companyID, p.Limit, p.Offset, search)
		if err != nil {
			helper.InternalError(c, err)
			return
		}

		helper.PaginatedItems(c, result, int64(total), p)
		return
	}

	// Default: Return all (existing behavior)
	accounts, err := h.usecase.GetAccountsWithBalances(c.Request.Context(), companyID)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, accounts)
}

// GetByID handles GET /accounts/:id
func (h *AccountHandler) GetByID(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "account")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		helper.NotFound(c, "account")
		return
	}

	helper.Success(c, result)
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
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "account")
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	input := account.UpdateAccountInput{
		ID:          id,
		Name:        req.Name,
		Type:        entity.AccountType(req.Type),
		ParentID:    helper.ParseOptionalUUID(req.ParentID),
		IsPostable:  req.IsPostable,
		IsActive:    req.IsActive,
		Description: req.Description,
	}

	result, err := h.usecase.UpdateAccount(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// Delete handles DELETE /accounts/:id
func (h *AccountHandler) Delete(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "account")
		return
	}

	if err := h.usecase.DeleteAccount(c.Request.Context(), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Deleted(c, "account")
}
