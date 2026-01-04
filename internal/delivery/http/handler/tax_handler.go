package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	taxuc "github.com/herman-xphp/bukuo/internal/usecase/tax"
	"github.com/shopspring/decimal"
)

// TaxHandler handles tax HTTP endpoints
type TaxHandler struct {
	usecase *taxuc.TaxUsecase
}

// NewTaxHandler creates a new TaxHandler
func NewTaxHandler(uc *taxuc.TaxUsecase) *TaxHandler {
	return &TaxHandler{usecase: uc}
}

type createTaxRateRequest struct {
	Name              string  `json:"name" binding:"required"`
	Code              string  `json:"code" binding:"required"`
	Type              string  `json:"type" binding:"required"`
	Rate              float64 `json:"rate" binding:"required"`
	SalesAccountID    string  `json:"sales_account_id" binding:"required"`
	PurchaseAccountID string  `json:"purchase_account_id" binding:"required"`
	Description       string  `json:"description"`
}

// CreateRate handles POST /tax/rates
func (h *TaxHandler) CreateRate(c *gin.Context) {
	var req createTaxRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	salesAccID, _ := uuid.Parse(req.SalesAccountID)
	purchaseAccID, _ := uuid.Parse(req.PurchaseAccountID)

	input := taxuc.CreateTaxRateInput{
		Name:              req.Name,
		Code:              req.Code,
		Type:              entity.TaxType(req.Type),
		Rate:              decimal.NewFromFloat(req.Rate),
		SalesAccountID:    salesAccID,
		PurchaseAccountID: purchaseAccID,
		Description:       req.Description,
	}

	result, err := h.usecase.CreateTaxRate(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// ListRates handles GET /tax/rates
func (h *TaxHandler) ListRates(c *gin.Context) {
	result, err := h.usecase.ListTaxRates(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// GetRate handles GET /tax/rates/:id
func (h *TaxHandler) GetRate(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "tax rate")
		return
	}

	result, err := h.usecase.GetTaxRate(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "tax rate")
		return
	}

	helper.Success(c, result)
}

// DeleteRate handles DELETE /tax/rates/:id
func (h *TaxHandler) DeleteRate(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "tax rate")
		return
	}

	if err := h.usecase.DeleteTaxRate(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.NoContent(c)
}

type createTaxReturnRequest struct {
	PeriodID      string  `json:"period_id" binding:"required"`
	TaxRateID     string  `json:"tax_rate_id" binding:"required"`
	ReturnDate    string  `json:"return_date" binding:"required"`
	TaxableAmount float64 `json:"taxable_amount" binding:"required"`
	TaxAmount     float64 `json:"tax_amount" binding:"required"`
	Credits       float64 `json:"credits"`
	Notes         string  `json:"notes"`
}

// CreateReturn handles POST /tax/returns
func (h *TaxHandler) CreateReturn(c *gin.Context) {
	var req createTaxReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	periodID, _ := uuid.Parse(req.PeriodID)
	taxRateID, _ := uuid.Parse(req.TaxRateID)
	returnDate, _ := time.Parse("2006-01-02", req.ReturnDate)

	input := taxuc.CreateTaxReturnInput{
		PeriodID:      periodID,
		TaxRateID:     taxRateID,
		ReturnDate:    returnDate,
		TaxableAmount: decimal.NewFromFloat(req.TaxableAmount),
		TaxAmount:     decimal.NewFromFloat(req.TaxAmount),
		Credits:       decimal.NewFromFloat(req.Credits),
		Notes:         req.Notes,
	}

	result, err := h.usecase.CreateTaxReturn(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// ListReturns handles GET /tax/returns
func (h *TaxHandler) ListReturns(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	result, total, err := h.usecase.ListTaxReturns(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, page, pageSize)
}

// GetReturn handles GET /tax/returns/:id
func (h *TaxHandler) GetReturn(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "tax return")
		return
	}

	result, err := h.usecase.GetTaxReturn(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "tax return")
		return
	}

	helper.Success(c, result)
}

// FileReturn handles POST /tax/returns/:id/file
func (h *TaxHandler) FileReturn(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "tax return")
		return
	}

	if err := h.usecase.FileTaxReturn(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "tax return filed"})
}

type payTaxReturnRequest struct {
	BankAccountID string `json:"bank_account_id" binding:"required"`
}

// PayReturn handles POST /tax/returns/:id/pay
func (h *TaxHandler) PayReturn(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "tax return")
		return
	}

	var req payTaxReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	bankAccID, _ := uuid.Parse(req.BankAccountID)

	if err := h.usecase.PayTaxReturn(c.Request.Context(), helper.GetCompanyID(c), id, helper.GetUserID(c), bankAccID); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "tax paid and journal created"})
}
