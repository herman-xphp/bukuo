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

// DeliveryHandler handles delivery order HTTP endpoints
type DeliveryHandler struct {
	usecase *sales.DeliveryUsecase
}

// NewDeliveryHandler creates a new DeliveryHandler
func NewDeliveryHandler(uc *sales.DeliveryUsecase) *DeliveryHandler {
	return &DeliveryHandler{usecase: uc}
}

type createDeliveryRequest struct {
	OrderID      string                `json:"order_id" binding:"required"`
	WarehouseID  string                `json:"warehouse_id" binding:"required"`
	DeliveryDate string                `json:"delivery_date" binding:"required"`
	Notes        string                `json:"notes"`
	Lines        []deliveryLineRequest `json:"lines" binding:"required,min=1"`
}

type deliveryLineRequest struct {
	OrderLineID string  `json:"order_line_id" binding:"required"`
	ProductID   string  `json:"product_id" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
}

// Create handles POST /deliveries
func (h *DeliveryHandler) Create(c *gin.Context) {
	var req createDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		helper.BadRequestMessage(c, "invalid order_id")
		return
	}

	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		helper.BadRequestMessage(c, "invalid warehouse_id")
		return
	}

	deliveryDate, _ := time.Parse("2006-01-02", req.DeliveryDate)

	lines := make([]sales.DeliveryLineInput, len(req.Lines))
	for i, l := range req.Lines {
		orderLineID, _ := uuid.Parse(l.OrderLineID)
		productID, _ := uuid.Parse(l.ProductID)
		lines[i] = sales.DeliveryLineInput{
			OrderLineID: orderLineID,
			ProductID:   productID,
			Quantity:    decimal.NewFromFloat(l.Quantity),
		}
	}

	input := sales.CreateDeliveryInput{
		OrderID:      orderID,
		WarehouseID:  warehouseID,
		DeliveryDate: deliveryDate,
		Notes:        req.Notes,
		Lines:        lines,
	}

	result, err := h.usecase.CreateDelivery(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// Get handles GET /deliveries/:id
func (h *DeliveryHandler) Get(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "delivery")
		return
	}

	result, err := h.usecase.GetDelivery(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "delivery")
		return
	}

	helper.Success(c, result)
}

// List handles GET /deliveries
func (h *DeliveryHandler) List(c *gin.Context) {
	filter := repository.SalesFilter{
		Page:     helper.GetPage(c),
		PageSize: helper.GetPageSize(c),
	}

	if customerID := c.Query("customer_id"); customerID != "" {
		id := helper.ParseUUIDString(customerID)
		filter.CustomerID = &id
	}

	result, total, err := h.usecase.ListDeliveries(c.Request.Context(), helper.GetCompanyID(c), filter)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, filter.Page, filter.PageSize)
}
