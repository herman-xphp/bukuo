package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/sales"
)

// SalesHandler handles sales endpoints
type SalesHandler struct {
	usecase *sales.SalesUsecase
}

// NewSalesHandler creates a new SalesHandler
func NewSalesHandler(uc *sales.SalesUsecase) *SalesHandler {
	return &SalesHandler{usecase: uc}
}

type CreateInvoiceRequest struct {
	InvoiceNo     string               `json:"invoice_no" binding:"required"`
	CustomerID    string               `json:"customer_id" binding:"required"`
	InvoiceDate   string               `json:"invoice_date" binding:"required"`
	DueDate       string               `json:"due_date" binding:"required"`
	Lines         []InvoiceLineRequest `json:"lines" binding:"required,min=1"`
	Notes         string               `json:"notes"`
	PaymentAmount string               `json:"payment_amount"`
	Status        string               `json:"status"`
}

type InvoiceLineRequest struct {
	ProductID   string `json:"product_id" binding:"required"`
	Description string `json:"description"`
	Quantity    string `json:"quantity" binding:"required"`
	UnitPrice   string `json:"unit_price" binding:"required"`
	DiscountPct string `json:"discount_pct"`
	TaxPct      string `json:"tax_pct"`
}

// CreateInvoice handles POST /sales/invoices
func (h *SalesHandler) CreateInvoice(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	fmt.Printf("DEBUG CreateInvoice Request: %+v\n", req)

	invoiceDate, _ := time.Parse("2006-01-02", req.InvoiceDate)
	dueDate, _ := time.Parse("2006-01-02", req.DueDate)

	var lines []sales.InvoiceLineInput
	for _, l := range req.Lines {
		lines = append(lines, sales.InvoiceLineInput{
			ProductID:   helper.ParseUUIDString(l.ProductID),
			Description: l.Description,
			Quantity:    helper.ParseDecimal(l.Quantity),
			UnitPrice:   helper.ParseDecimal(l.UnitPrice),
			DiscountPct: helper.ParseDecimal(l.DiscountPct),
			TaxPct:      helper.ParseDecimal(l.TaxPct),
		})
	}

	// Default status
	status := entity.SalesStatusDraft
	if req.Status != "" {
		status = entity.SalesStatus(req.Status)
	}

	result, err := h.usecase.CreateInvoice(c.Request.Context(), sales.CreateInvoiceInput{
		CompanyID:     helper.GetCompanyID(c),
		InvoiceNo:     req.InvoiceNo,
		CustomerID:    helper.ParseUUIDString(req.CustomerID),
		InvoiceDate:   invoiceDate,
		DueDate:       dueDate,
		Lines:         lines,
		Notes:         req.Notes,
		Status:        status,
		PaymentAmount: helper.ParseDecimal(req.PaymentAmount),
	})

	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// ListInvoices handles GET /sales/invoices
func (h *SalesHandler) ListInvoices(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	p := helper.ParsePagination(c)

	invoices, total, err := h.usecase.ListInvoices(c.Request.Context(), companyID, p.Page, p.PageSize)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	// Custom response to maintain compatibility
	c.JSON(http.StatusOK, gin.H{
		"data": invoices,
		"meta": gin.H{
			"total":     total,
			"page":      p.Page,
			"page_size": p.PageSize,
		},
	})
}

// GetInvoice handles GET /sales/invoices/:id
func (h *SalesHandler) GetInvoice(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	invoiceID, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "invoice")
		return
	}

	invoice, err := h.usecase.GetInvoice(c.Request.Context(), companyID, invoiceID)
	if err != nil {
		helper.NotFound(c, "invoice")
		return
	}

	helper.Success(c, invoice)
}

// VoidInvoice handles POST /sales/invoices/:id/void
func (h *SalesHandler) VoidInvoice(c *gin.Context) {
	companyID := helper.GetCompanyID(c)
	invoiceID, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "invoice")
		return
	}

	if err := h.usecase.VoidInvoice(c.Request.Context(), companyID, invoiceID); err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Message(c, "invoice voided successfully")
}

// ListOrders handles GET /sales/orders (placeholder)
func (h *SalesHandler) ListOrders(c *gin.Context) {
	helper.Success(c, []interface{}{})
}

// ListQuotations handles GET /sales/quotations (placeholder)
func (h *SalesHandler) ListQuotations(c *gin.Context) {
	helper.Success(c, []interface{}{})
}
