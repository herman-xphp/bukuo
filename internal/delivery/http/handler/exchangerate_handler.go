package handler

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/usecase/exchangerate"
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
		helper.BadRequest(c, err)
		return
	}

	effectiveDate, _ := time.Parse("2006-01-02", req.EffectiveDate)

	input := exchangerate.CreateExchangeRateInput{
		CompanyID:      helper.GetCompanyID(c),
		FromCurrencyID: helper.ParseUUIDString(req.FromCurrencyID),
		ToCurrencyID:   helper.ParseUUIDString(req.ToCurrencyID),
		Rate:           helper.ParseDecimal(req.Rate),
		EffectiveDate:  effectiveDate,
	}

	result, err := h.usecase.CreateExchangeRate(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// List handles GET /exchange-rates
func (h *ExchangeRateHandler) List(c *gin.Context) {
	companyID := helper.GetCompanyID(c)

	var fromCurrencyID, toCurrencyID *helper.MoneyAmount
	// These are UUIDs, not decimals
	fromID := helper.ParseOptionalUUID(ptrString(c.Query("from_currency_id")))
	toID := helper.ParseOptionalUUID(ptrString(c.Query("to_currency_id")))

	rates, err := h.usecase.List(c.Request.Context(), companyID, fromID, toID)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	// Suppress unused variable warning
	_ = fromCurrencyID
	_ = toCurrencyID

	helper.Success(c, rates)
}

// Helper to get string pointer
func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// GetByID handles GET /exchange-rates/:id
func (h *ExchangeRateHandler) GetByID(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "exchange rate")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		helper.NotFound(c, "exchange rate")
		return
	}

	helper.Success(c, result)
}

// Convert handles GET /exchange-rates/convert
func (h *ExchangeRateHandler) Convert(c *gin.Context) {
	companyID := helper.GetCompanyID(c)

	fromCurrencyID := helper.ParseUUIDString(c.Query("from_currency_id"))
	toCurrencyID := helper.ParseUUIDString(c.Query("to_currency_id"))
	amount := helper.ParseDecimal(c.Query("amount"))

	if fromCurrencyID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.BadRequestMessage(c, "from_currency_id is required")
		return
	}
	if toCurrencyID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.BadRequestMessage(c, "to_currency_id is required")
		return
	}

	date := time.Now()
	if dateStr := c.Query("date"); dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
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
		helper.NotFound(c, "exchange rate")
		return
	}

	helper.Success(c, result)
}

// UpdateExchangeRateRequest represents update exchange rate request
type UpdateExchangeRateRequest struct {
	Rate string `json:"rate" binding:"required"`
}

// Update handles PUT /exchange-rates/:id
func (h *ExchangeRateHandler) Update(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "exchange rate")
		return
	}

	var req UpdateExchangeRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	result, err := h.usecase.UpdateExchangeRate(c.Request.Context(), companyID, id, helper.ParseDecimal(req.Rate))
	if err != nil {
		if errors.Is(err, exchangerate.ErrExchangeRateNotFound) {
			helper.NotFound(c, "exchange rate")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// Delete handles DELETE /exchange-rates/:id
func (h *ExchangeRateHandler) Delete(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "exchange rate")
		return
	}

	if err := h.usecase.DeleteExchangeRate(c.Request.Context(), companyID, id); err != nil {
		if errors.Is(err, exchangerate.ErrExchangeRateNotFound) {
			helper.NotFound(c, "exchange rate")
			return
		}
		helper.BadRequest(c, err)
		return
	}

	helper.Deleted(c, "exchange rate")
}
