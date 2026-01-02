package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))

	input := currency.CreateCurrencyInput{
		CompanyID:     companyID,
		Code:          req.Code,
		Name:          req.Name,
		Symbol:        req.Symbol,
		DecimalPlaces: req.DecimalPlaces,
	}

	result, err := h.usecase.CreateCurrency(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, currency.ErrCurrencyCodeExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// CreateFromPresetRequest represents create from preset request
type CreateFromPresetRequest struct {
	Code string `json:"code" binding:"required"`
}

// CreateFromPreset handles POST /currencies/preset
func (h *CurrencyHandler) CreateFromPreset(c *gin.Context) {
	var req CreateFromPresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))

	result, err := h.usecase.CreateFromPreset(c.Request.Context(), companyID, req.Code)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, currency.ErrCurrencyCodeExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// List handles GET /currencies
func (h *CurrencyHandler) List(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	currencies, err := h.usecase.List(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": currencies})
}

// GetByID handles GET /currencies/:id
func (h *CurrencyHandler) GetByID(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid currency id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "currency not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetBase handles GET /currencies/base
func (h *CurrencyHandler) GetBase(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	result, err := h.usecase.GetBaseCurrency(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no base currency set"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// SetBase handles POST /currencies/:id/set-base
func (h *CurrencyHandler) SetBase(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid currency id"})
		return
	}

	if err := h.usecase.SetBaseCurrency(c.Request.Context(), companyID, id); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, currency.ErrCurrencyNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "base currency updated"})
}

// UpdateCurrencyRequest represents update currency request
type UpdateCurrencyRequest struct {
	Name          string `json:"name" binding:"required"`
	Symbol        string `json:"symbol" binding:"required"`
	DecimalPlaces int    `json:"decimal_places"`
}

// Update handles PUT /currencies/:id
func (h *CurrencyHandler) Update(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid currency id"})
		return
	}

	var req UpdateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		status := http.StatusBadRequest
		if errors.Is(err, currency.ErrCurrencyNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete handles DELETE /currencies/:id
func (h *CurrencyHandler) Delete(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid currency id"})
		return
	}

	if err := h.usecase.DeleteCurrency(c.Request.Context(), companyID, id); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, currency.ErrCurrencyNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, currency.ErrCannotDeleteBase) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "currency deleted"})
}
