package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	arapuc "github.com/herman-xphp/bukuo/internal/usecase/arap"
)

// ARAPHandler handles AR/AP reporting HTTP endpoints
type ARAPHandler struct {
	usecase *arapuc.ARAPUsecase
}

// NewARAPHandler creates a new ARAPHandler
func NewARAPHandler(uc *arapuc.ARAPUsecase) *ARAPHandler {
	return &ARAPHandler{usecase: uc}
}

// ARAgingReport handles GET /arap/receivables/aging
func (h *ARAPHandler) ARAgingReport(c *gin.Context) {
	report, err := h.usecase.GetARAgingReport(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c, report)
}

// APAgingReport handles GET /arap/payables/aging
func (h *ARAPHandler) APAgingReport(c *gin.Context) {
	report, err := h.usecase.GetAPAgingReport(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c, report)
}

// OutstandingReceivables handles GET /arap/receivables/outstanding
func (h *ARAPHandler) OutstandingReceivables(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	invoices, total, totalAmount, err := h.usecase.GetOutstandingReceivables(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{
		"data":         invoices,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_amount": totalAmount,
	})
}

// OutstandingPayables handles GET /arap/payables/outstanding
func (h *ARAPHandler) OutstandingPayables(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	invoices, total, totalAmount, err := h.usecase.GetOutstandingPayables(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{
		"data":         invoices,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_amount": totalAmount,
	})
}

// Summary handles GET /arap/summary
func (h *ARAPHandler) Summary(c *gin.Context) {
	summary, err := h.usecase.GetARAPSummary(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c, summary)
}
