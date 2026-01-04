package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	inventoryuc "github.com/herman-xphp/bukuo/internal/usecase/inventory"
	"github.com/shopspring/decimal"
)

// OpnameHandler handles stock opname HTTP endpoints
type OpnameHandler struct {
	usecase *inventoryuc.OpnameUsecase
}

// NewOpnameHandler creates a new OpnameHandler
func NewOpnameHandler(uc *inventoryuc.OpnameUsecase) *OpnameHandler {
	return &OpnameHandler{usecase: uc}
}

type createOpnameRequest struct {
	WarehouseID string              `json:"warehouse_id" binding:"required"`
	OpnameDate  string              `json:"opname_date" binding:"required"`
	Notes       string              `json:"notes"`
	Lines       []opnameLineRequest `json:"lines" binding:"required,min=1"`
}

type opnameLineRequest struct {
	ProductID string  `json:"product_id" binding:"required"`
	ActualQty float64 `json:"actual_qty" binding:"required,gte=0"`
	Notes     string  `json:"notes"`
}

// Create handles POST /inventory/opname
func (h *OpnameHandler) Create(c *gin.Context) {
	var req createOpnameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		helper.BadRequestMessage(c, "invalid warehouse_id")
		return
	}

	opnameDate, _ := time.Parse("2006-01-02", req.OpnameDate)

	lines := make([]inventoryuc.OpnameLineInput, len(req.Lines))
	for i, l := range req.Lines {
		productID, _ := uuid.Parse(l.ProductID)
		lines[i] = inventoryuc.OpnameLineInput{
			ProductID: productID,
			ActualQty: decimal.NewFromFloat(l.ActualQty),
			Notes:     l.Notes,
		}
	}

	input := inventoryuc.CreateOpnameInput{
		WarehouseID: warehouseID,
		OpnameDate:  opnameDate,
		Notes:       req.Notes,
		Lines:       lines,
	}

	result, err := h.usecase.CreateOpname(c.Request.Context(), helper.GetCompanyID(c), helper.GetUserID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// Get handles GET /inventory/opname/:id
func (h *OpnameHandler) Get(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "opname")
		return
	}

	result, err := h.usecase.GetOpname(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "opname")
		return
	}

	helper.Success(c, result)
}

// List handles GET /inventory/opname
func (h *OpnameHandler) List(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	result, total, err := h.usecase.ListOpnames(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, page, pageSize)
}

// Approve handles POST /inventory/opname/:id/approve
func (h *OpnameHandler) Approve(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "opname")
		return
	}

	if err := h.usecase.ApproveOpname(c.Request.Context(), helper.GetCompanyID(c), id, helper.GetUserID(c)); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "opname approved and adjustments applied"})
}
