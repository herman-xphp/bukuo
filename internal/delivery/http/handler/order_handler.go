package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/sales"
	"github.com/shopspring/decimal"
)

// OrderHandler handles sales order HTTP endpoints
type OrderHandler struct {
	usecase *sales.OrderUsecase
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(uc *sales.OrderUsecase) *OrderHandler {
	return &OrderHandler{usecase: uc}
}

type createOrderRequest struct {
	CustomerID   string             `json:"customer_id" binding:"required"`
	QuotationID  string             `json:"quotation_id"`
	OrderDate    string             `json:"order_date" binding:"required"`
	DeliveryDate string             `json:"delivery_date"`
	Notes        string             `json:"notes"`
	Lines        []orderLineRequest `json:"lines" binding:"required,min=1"`
}

type orderLineRequest struct {
	ProductID   string  `json:"product_id" binding:"required"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" binding:"required,gte=0"`
	DiscountPct float64 `json:"discount_pct"`
	TaxPct      float64 `json:"tax_pct"`
}

// Create handles POST /orders
func (h *OrderHandler) Create(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		helper.BadRequestMessage(c, "invalid customer_id")
		return
	}

	orderDate, _ := time.Parse("2006-01-02", req.OrderDate)

	var quotationID *uuid.UUID
	if req.QuotationID != "" {
		id, err := uuid.Parse(req.QuotationID)
		if err == nil {
			quotationID = &id
		}
	}

	var deliveryDate *time.Time
	if req.DeliveryDate != "" {
		d, _ := time.Parse("2006-01-02", req.DeliveryDate)
		deliveryDate = &d
	}

	lines := make([]sales.OrderLineInput, len(req.Lines))
	for i, l := range req.Lines {
		productID, _ := uuid.Parse(l.ProductID)
		lines[i] = sales.OrderLineInput{
			ProductID:   productID,
			Description: l.Description,
			Quantity:    decimal.NewFromFloat(l.Quantity),
			UnitPrice:   decimal.NewFromFloat(l.UnitPrice),
			DiscountPct: decimal.NewFromFloat(l.DiscountPct),
			TaxPct:      decimal.NewFromFloat(l.TaxPct),
		}
	}

	input := sales.CreateOrderInput{
		CustomerID:   customerID,
		QuotationID:  quotationID,
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

// Get handles GET /orders/:id
func (h *OrderHandler) Get(c *gin.Context) {
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

// List handles GET /orders
func (h *OrderHandler) List(c *gin.Context) {
	filter := repository.SalesFilter{
		Page:     helper.GetPage(c),
		PageSize: helper.GetPageSize(c),
	}

	if customerID := c.Query("customer_id"); customerID != "" {
		id := helper.ParseUUIDString(customerID)
		filter.CustomerID = &id
	}

	result, total, err := h.usecase.ListOrders(c.Request.Context(), helper.GetCompanyID(c), filter)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, filter.Page, filter.PageSize)
}

// Confirm handles POST /orders/:id/confirm
func (h *OrderHandler) Confirm(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "order")
		return
	}

	if err := h.usecase.ConfirmOrder(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "order confirmed"})
}

// Cancel handles POST /orders/:id/cancel
func (h *OrderHandler) Cancel(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "order")
		return
	}

	if err := h.usecase.CancelOrder(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "order cancelled"})
}
