package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	purchasinguc "github.com/herman-xphp/bukuo/internal/usecase/purchasing"
	"github.com/shopspring/decimal"
)

// PurchasingHandler handles purchasing HTTP endpoints
type PurchasingHandler struct {
	usecase *purchasinguc.PurchasingUsecase
}

// NewPurchasingHandler creates a new PurchasingHandler
func NewPurchasingHandler(uc *purchasinguc.PurchasingUsecase) *PurchasingHandler {
	return &PurchasingHandler{usecase: uc}
}

type createPurchaseOrderRequest struct {
	SupplierID   string                `json:"supplier_id" binding:"required"`
	OrderDate    string                `json:"order_date" binding:"required"`
	DeliveryDate *string               `json:"delivery_date"`
	Notes        string                `json:"notes"`
	Lines        []purchaseLineRequest `json:"lines" binding:"required,min=1"`
}

type purchaseLineRequest struct {
	ProductID   string  `json:"product_id" binding:"required"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" binding:"required,gte=0"`
	DiscountPct float64 `json:"discount_pct"`
	TaxPct      float64 `json:"tax_pct"`
}

// CreateOrder handles POST /purchasing/orders
func (h *PurchasingHandler) CreateOrder(c *gin.Context) {
	var req createPurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	supplierID, _ := uuid.Parse(req.SupplierID)
	orderDate, _ := time.Parse("2006-01-02", req.OrderDate)
	var deliveryDate *time.Time
	if req.DeliveryDate != nil {
		d, _ := time.Parse("2006-01-02", *req.DeliveryDate)
		deliveryDate = &d
	}

	lines := make([]purchasinguc.OrderLineInput, len(req.Lines))
	for i, l := range req.Lines {
		productID, _ := uuid.Parse(l.ProductID)
		lines[i] = purchasinguc.OrderLineInput{
			ProductID:   productID,
			Description: l.Description,
			Quantity:    decimal.NewFromFloat(l.Quantity),
			UnitPrice:   decimal.NewFromFloat(l.UnitPrice),
			DiscountPct: decimal.NewFromFloat(l.DiscountPct),
			TaxPct:      decimal.NewFromFloat(l.TaxPct),
		}
	}

	input := purchasinguc.CreateOrderInput{
		SupplierID:   supplierID,
		OrderDate:    orderDate,
		DeliveryDate: deliveryDate,
		Notes:        req.Notes,
		Lines:        lines,
	}

	result, err := h.usecase.CreateOrder(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// ListOrders handles GET /purchasing/orders
func (h *PurchasingHandler) ListOrders(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	result, total, err := h.usecase.ListOrders(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, page, pageSize)
}

// GetOrder handles GET /purchasing/orders/:id
func (h *PurchasingHandler) GetOrder(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "order")
		return
	}

	result, err := h.usecase.GetOrder(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "order")
		return
	}

	helper.Success(c, result)
}

// ApproveOrder handles POST /purchasing/orders/:id/approve
func (h *PurchasingHandler) ApproveOrder(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "order")
		return
	}

	if err := h.usecase.ApproveOrder(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "order approved"})
}

type createPurchaseInvoiceRequest struct {
	SupplierID  string                `json:"supplier_id" binding:"required"`
	OrderID     *string               `json:"order_id"`
	InvoiceDate string                `json:"invoice_date" binding:"required"`
	DueDate     string                `json:"due_date" binding:"required"`
	Notes       string                `json:"notes"`
	Lines       []purchaseLineRequest `json:"lines" binding:"required,min=1"`
}

// CreateInvoice handles POST /purchasing/invoices
func (h *PurchasingHandler) CreateInvoice(c *gin.Context) {
	var req createPurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	supplierID, _ := uuid.Parse(req.SupplierID)
	invoiceDate, _ := time.Parse("2006-01-02", req.InvoiceDate)
	dueDate, _ := time.Parse("2006-01-02", req.DueDate)
	var orderID *uuid.UUID
	if req.OrderID != nil {
		id, _ := uuid.Parse(*req.OrderID)
		orderID = &id
	}

	lines := make([]purchasinguc.InvoiceLineInput, len(req.Lines))
	for i, l := range req.Lines {
		productID, _ := uuid.Parse(l.ProductID)
		lines[i] = purchasinguc.InvoiceLineInput{
			ProductID:   productID,
			Description: l.Description,
			Quantity:    decimal.NewFromFloat(l.Quantity),
			UnitPrice:   decimal.NewFromFloat(l.UnitPrice),
			DiscountPct: decimal.NewFromFloat(l.DiscountPct),
			TaxPct:      decimal.NewFromFloat(l.TaxPct),
		}
	}

	input := purchasinguc.CreateInvoiceInput{
		SupplierID:  supplierID,
		OrderID:     orderID,
		InvoiceDate: invoiceDate,
		DueDate:     dueDate,
		Notes:       req.Notes,
		Lines:       lines,
	}

	result, err := h.usecase.CreateInvoice(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// ListInvoices handles GET /purchasing/invoices
func (h *PurchasingHandler) ListInvoices(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	result, total, err := h.usecase.ListInvoices(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, page, pageSize)
}

// GetInvoice handles GET /purchasing/invoices/:id
func (h *PurchasingHandler) GetInvoice(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "invoice")
		return
	}

	result, err := h.usecase.GetInvoice(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "invoice")
		return
	}

	helper.Success(c, result)
}
