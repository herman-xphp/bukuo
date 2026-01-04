package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	inventoryuc "github.com/herman-xphp/bukuo/internal/usecase/inventory"
	"github.com/shopspring/decimal"
)

// TransferHandler handles stock transfer HTTP endpoints
type TransferHandler struct {
	usecase *inventoryuc.TransferUsecase
}

// NewTransferHandler creates a new TransferHandler
func NewTransferHandler(uc *inventoryuc.TransferUsecase) *TransferHandler {
	return &TransferHandler{usecase: uc}
}

type createTransferRequest struct {
	FromWarehouseID string                `json:"from_warehouse_id" binding:"required"`
	ToWarehouseID   string                `json:"to_warehouse_id" binding:"required"`
	TransferDate    string                `json:"transfer_date" binding:"required"`
	Notes           string                `json:"notes"`
	Lines           []transferLineRequest `json:"lines" binding:"required,min=1"`
}

type transferLineRequest struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}

// Create handles POST /inventory/transfers
func (h *TransferHandler) Create(c *gin.Context) {
	var req createTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	fromWarehouseID, err := uuid.Parse(req.FromWarehouseID)
	if err != nil {
		helper.BadRequestMessage(c, "invalid from_warehouse_id")
		return
	}

	toWarehouseID, err := uuid.Parse(req.ToWarehouseID)
	if err != nil {
		helper.BadRequestMessage(c, "invalid to_warehouse_id")
		return
	}

	transferDate, _ := time.Parse("2006-01-02", req.TransferDate)

	lines := make([]inventoryuc.TransferLineInput, len(req.Lines))
	for i, l := range req.Lines {
		productID, _ := uuid.Parse(l.ProductID)
		lines[i] = inventoryuc.TransferLineInput{
			ProductID: productID,
			Quantity:  decimal.NewFromFloat(l.Quantity),
		}
	}

	input := inventoryuc.CreateTransferInput{
		FromWarehouseID: fromWarehouseID,
		ToWarehouseID:   toWarehouseID,
		TransferDate:    transferDate,
		Notes:           req.Notes,
		Lines:           lines,
	}

	result, err := h.usecase.CreateTransfer(c.Request.Context(), helper.GetCompanyID(c), helper.GetUserID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// Get handles GET /inventory/transfers/:id
func (h *TransferHandler) Get(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "transfer")
		return
	}

	result, err := h.usecase.GetTransfer(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "transfer")
		return
	}

	helper.Success(c, result)
}

// List handles GET /inventory/transfers
func (h *TransferHandler) List(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	result, total, err := h.usecase.ListTransfers(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, page, pageSize)
}
