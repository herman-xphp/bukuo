package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/usecase/exchangerate"
	"github.com/shopspring/decimal"
)

// ExchangeRateHandler handles exchange rate endpoints
type ExchangeRateHandler struct {
	usecase *exchangerate.ExchangeRateUsecase
}

// NewExchangeRateHandler creates a new ExchangeRateHandler
func NewExchangeRateHandler(uc *exchangerate.ExchangeRateUsecase) *ExchangeRateHandler {
	return &ExchangeRateHandler{usecase: uc}
}

// CreateExchangeRateRequest represents create exchange rate request
type CreateExchangeRateRequest struct {
	FromCurrencyID string `json:"from_currency_id" binding:"required"`
	ToCurrencyID   string `json:"to_currency_id" binding:"required"`
	Rate           string `json:"rate" binding:"required"`
	EffectiveDate  string `json:"effective_date" binding:"required"` // YYYY-MM-DD
}

// Create handles POST /exchange-rates
func (h *ExchangeRateHandler) Create(c *gin.Context) {
	var req CreateExchangeRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))
	fromCurrencyID, _ := uuid.Parse(req.FromCurrencyID)
	toCurrencyID, _ := uuid.Parse(req.ToCurrencyID)
	rate, _ := decimal.NewFromString(req.Rate)
	effectiveDate, _ := time.Parse("2006-01-02", req.EffectiveDate)

	input := exchangerate.CreateExchangeRateInput{
		CompanyID:      companyID,
		FromCurrencyID: fromCurrencyID,
		ToCurrencyID:   toCurrencyID,
		Rate:           rate,
		EffectiveDate:  effectiveDate,
	}

	result, err := h.usecase.CreateExchangeRate(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, exchangerate.ErrSameCurrency) || errors.Is(err, exchangerate.ErrInvalidRate) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// List handles GET /exchange-rates
func (h *ExchangeRateHandler) List(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	var fromCurrencyID, toCurrencyID *uuid.UUID

	if from := c.Query("from_currency_id"); from != "" {
		id, _ := uuid.Parse(from)
		fromCurrencyID = &id
	}

	if to := c.Query("to_currency_id"); to != "" {
		id, _ := uuid.Parse(to)
		toCurrencyID = &id
	}

	rates, err := h.usecase.List(c.Request.Context(), companyID, fromCurrencyID, toCurrencyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rates})
}

// GetByID handles GET /exchange-rates/:id
func (h *ExchangeRateHandler) GetByID(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exchange rate id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "exchange rate not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Convert handles GET /exchange-rates/convert
func (h *ExchangeRateHandler) Convert(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	fromCurrencyID, err := uuid.Parse(c.Query("from_currency_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_currency_id is required"})
		return
	}

	toCurrencyID, err := uuid.Parse(c.Query("to_currency_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to_currency_id is required"})
		return
	}

	amount, err := decimal.NewFromString(c.Query("amount"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}

	date := time.Now()
	if dateStr := c.Query("date"); dateStr != "" {
		d, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			date = d
		}
	}

	input := exchangerate.ConvertInput{
		CompanyID:      companyID,
		FromCurrencyID: fromCurrencyID,
		ToCurrencyID:   toCurrencyID,
		Amount:         amount,
		Date:           date,
	}

	result, err := h.usecase.Convert(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateExchangeRateRequest represents update exchange rate request
type UpdateExchangeRateRequest struct {
	Rate string `json:"rate" binding:"required"`
}

// Update handles PUT /exchange-rates/:id
func (h *ExchangeRateHandler) Update(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exchange rate id"})
		return
	}

	var req UpdateExchangeRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rate, _ := decimal.NewFromString(req.Rate)

	result, err := h.usecase.UpdateExchangeRate(c.Request.Context(), companyID, id, rate)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, exchangerate.ErrExchangeRateNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete handles DELETE /exchange-rates/:id
func (h *ExchangeRateHandler) Delete(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exchange rate id"})
		return
	}

	if err := h.usecase.DeleteExchangeRate(c.Request.Context(), companyID, id); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, exchangerate.ErrExchangeRateNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "exchange rate deleted"})
}
