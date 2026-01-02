package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/usecase/sales"
	"github.com/shopspring/decimal"
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
	InvoiceNo   string               `json:"invoice_no" binding:"required"`
	CustomerID  string               `json:"customer_id" binding:"required"`
	InvoiceDate string               `json:"invoice_date" binding:"required"`
	DueDate     string               `json:"due_date" binding:"required"`
	Lines       []InvoiceLineRequest `json:"lines" binding:"required,min=1"`
	Notes       string               `json:"notes"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))
	customerID, _ := uuid.Parse(req.CustomerID)
	invoiceDate, _ := time.Parse("2006-01-02", req.InvoiceDate)
	dueDate, _ := time.Parse("2006-01-02", req.DueDate)

	var lines []sales.InvoiceLineInput
	for _, l := range req.Lines {
		productID, _ := uuid.Parse(l.ProductID)
		qty, _ := decimal.NewFromString(l.Quantity)
		price, _ := decimal.NewFromString(l.UnitPrice)
		discPct, _ := decimal.NewFromString(l.DiscountPct)
		taxPct, _ := decimal.NewFromString(l.TaxPct)

		lines = append(lines, sales.InvoiceLineInput{
			ProductID:   productID,
			Description: l.Description,
			Quantity:    qty,
			UnitPrice:   price,
			DiscountPct: discPct,
			TaxPct:      taxPct,
		})
	}

	result, err := h.usecase.CreateInvoice(c.Request.Context(), sales.CreateInvoiceInput{
		CompanyID:   companyID,
		InvoiceNo:   req.InvoiceNo,
		CustomerID:  customerID,
		InvoiceDate: invoiceDate,
		DueDate:     dueDate,
		Lines:       lines,
		Notes:       req.Notes,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// ListInvoices handles GET /sales/invoices (placeholder)
func (h *SalesHandler) ListInvoices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "message": "Sales invoices list - DB integration pending"})
}

// ListOrders handles GET /sales/orders (placeholder)
func (h *SalesHandler) ListOrders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "message": "Sales orders list - DB integration pending"})
}

// ListQuotations handles GET /sales/quotations (placeholder)
func (h *SalesHandler) ListQuotations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "message": "Quotations list - DB integration pending"})
}
