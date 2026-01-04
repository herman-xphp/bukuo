package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/sales"
	"github.com/shopspring/decimal"
)

// QuotationHandler handles quotation HTTP endpoints
type QuotationHandler struct {
	usecase *sales.QuotationUsecase
}

// NewQuotationHandler creates a new QuotationHandler
func NewQuotationHandler(uc *sales.QuotationUsecase) *QuotationHandler {
	return &QuotationHandler{usecase: uc}
}

type createQuotationRequest struct {
	CustomerID    string                 `json:"customer_id" binding:"required"`
	QuotationDate string                 `json:"quotation_date" binding:"required"`
	ValidUntil    string                 `json:"valid_until" binding:"required"`
	Notes         string                 `json:"notes"`
	Lines         []quotationLineRequest `json:"lines" binding:"required,min=1"`
}

type quotationLineRequest struct {
	ProductID   string  `json:"product_id" binding:"required"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" binding:"required,gte=0"`
	DiscountPct float64 `json:"discount_pct"`
	TaxPct      float64 `json:"tax_pct"`
}

// Create handles POST /quotations
func (h *QuotationHandler) Create(c *gin.Context) {
	var req createQuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		helper.BadRequestMessage(c, "invalid customer_id")
		return
	}

	quotationDate, _ := time.Parse("2006-01-02", req.QuotationDate)
	validUntil, _ := time.Parse("2006-01-02", req.ValidUntil)

	lines := make([]sales.QuotationLineInput, len(req.Lines))
	for i, l := range req.Lines {
		productID, _ := uuid.Parse(l.ProductID)
		lines[i] = sales.QuotationLineInput{
			ProductID:   productID,
			Description: l.Description,
			Quantity:    decimal.NewFromFloat(l.Quantity),
			UnitPrice:   decimal.NewFromFloat(l.UnitPrice),
			DiscountPct: decimal.NewFromFloat(l.DiscountPct),
			TaxPct:      decimal.NewFromFloat(l.TaxPct),
		}
	}

	input := sales.CreateQuotationInput{
		CustomerID:    customerID,
		QuotationDate: quotationDate,
		ValidUntil:    validUntil,
		Notes:         req.Notes,
		Lines:         lines,
	}

	result, err := h.usecase.CreateQuotation(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// Get handles GET /quotations/:id
func (h *QuotationHandler) Get(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "quotation")
		return
	}

	result, err := h.usecase.GetQuotation(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "quotation")
		return
	}

	helper.Success(c, result)
}

// List handles GET /quotations
func (h *QuotationHandler) List(c *gin.Context) {
	filter := repository.SalesFilter{
		Page:     helper.GetPage(c),
		PageSize: helper.GetPageSize(c),
	}

	if customerID := c.Query("customer_id"); customerID != "" {
		id := helper.ParseUUIDString(customerID)
		filter.CustomerID = &id
	}

	result, total, err := h.usecase.ListQuotations(c.Request.Context(), helper.GetCompanyID(c), filter)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, filter.Page, filter.PageSize)
}

// Send handles POST /quotations/:id/send
func (h *QuotationHandler) Send(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "quotation")
		return
	}

	if err := h.usecase.SendQuotation(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "quotation sent"})
}

// Accept handles POST /quotations/:id/accept
func (h *QuotationHandler) Accept(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "quotation")
		return
	}

	order, err := h.usecase.AcceptQuotation(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "quotation accepted", "order": order})
}

// Delete handles DELETE /quotations/:id
func (h *QuotationHandler) Delete(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "quotation")
		return
	}

	if err := h.usecase.DeleteQuotation(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.NoContent(c)
}
