package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/usecase/currency"
)

// CurrencyHandler handles currency endpoints
type CurrencyHandler struct {
	usecase *currency.CurrencyUsecase
}

// NewCurrencyHandler creates a new CurrencyHandler
func NewCurrencyHandler(uc *currency.CurrencyUsecase) *CurrencyHandler {
	return &CurrencyHandler{usecase: uc}
}

// CreateCurrencyRequest represents create currency request
type CreateCurrencyRequest struct {
	Code          string `json:"code" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Symbol        string `json:"symbol" binding:"required"`
	DecimalPlaces int    `json:"decimal_places"`
}

// Create handles POST /currencies
func (h *CurrencyHandler) Create(c *gin.Context) {
	var req CreateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	input := currency.CreateCurrencyInput{
		CompanyID:     helper.GetCompanyID(c),
		Code:          req.Code,
		Name:          req.Name,
		Symbol:        req.Symbol,
		DecimalPlaces: req.DecimalPlaces,
	}

	result, err := h.usecase.CreateCurrency(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, currency.ErrCurrencyCodeExists) {
			helper.Conflict(c, err)
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// CreateFromPresetRequest represents create from preset request
type CreateFromPresetRequest struct {
	Code string `json:"code" binding:"required"`
}

// CreateFromPreset handles POST /currencies/preset
func (h *CurrencyHandler) CreateFromPreset(c *gin.Context) {
	var req CreateFromPresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	result, err := h.usecase.CreateFromPreset(c.Request.Context(), helper.GetCompanyID(c), req.Code)
	if err != nil {
		if errors.Is(err, currency.ErrCurrencyCodeExists) {
			helper.Conflict(c, err)
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// List handles GET /currencies
func (h *CurrencyHandler) List(c *gin.Context) {
	currencies, err := h.usecase.List(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, currencies)
}

// GetByID handles GET /currencies/:id
func (h *CurrencyHandler) GetByID(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "currency")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		helper.NotFound(c, "currency")
		return
	}

	helper.Success(c, result)
}

// GetBase handles GET /currencies/base
func (h *CurrencyHandler) GetBase(c *gin.Context) {
	result, err := h.usecase.GetBaseCurrency(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.NotFound(c, "base currency")
		return
	}

	helper.Success(c, result)
}

// SetBase handles POST /currencies/:id/set-base
func (h *CurrencyHandler) SetBase(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "currency")
		return
	}

	if err := h.usecase.SetBaseCurrency(c.Request.Context(), companyID, id); err != nil {
		if errors.Is(err, currency.ErrCurrencyNotFound) {
			helper.NotFound(c, "currency")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Message(c, "base currency updated")
}

// UpdateCurrencyRequest represents update currency request
type UpdateCurrencyRequest struct {
	Name          string `json:"name" binding:"required"`
	Symbol        string `json:"symbol" binding:"required"`
	DecimalPlaces int    `json:"decimal_places"`
}

// Update handles PUT /currencies/:id
func (h *CurrencyHandler) Update(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "currency")
		return
	}

	var req UpdateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	input := currency.UpdateCurrencyInput{
		CompanyID:     companyID,
		ID:            id,
		Name:          req.Name,
		Symbol:        req.Symbol,
		DecimalPlaces: req.DecimalPlaces,
	}

	result, err := h.usecase.UpdateCurrency(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, currency.ErrCurrencyNotFound) {
			helper.NotFound(c, "currency")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// Delete handles DELETE /currencies/:id
func (h *CurrencyHandler) Delete(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "currency")
		return
	}

	if err := h.usecase.DeleteCurrency(c.Request.Context(), companyID, id); err != nil {
		if errors.Is(err, currency.ErrCurrencyNotFound) {
			helper.NotFound(c, "currency")
			return
		}
		if errors.Is(err, currency.ErrCannotDeleteBase) {
			helper.Conflict(c, err)
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Deleted(c, "currency")
}
