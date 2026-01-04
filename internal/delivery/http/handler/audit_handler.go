package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	audituc "github.com/herman-xphp/bukuo/internal/usecase/audit"
)

// AuditHandler handles audit trail HTTP endpoints
type AuditHandler struct {
	usecase *audituc.AuditUsecase
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(uc *audituc.AuditUsecase) *AuditHandler {
	return &AuditHandler{usecase: uc}
}

// List handles GET /audit/logs
func (h *AuditHandler) List(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	logs, err := h.usecase.GetCompanyLogs(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{
		"data":      logs,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetByEntity handles GET /audit/entity/:type/:id
func (h *AuditHandler) GetByEntity(c *gin.Context) {
	entityType := c.Param("type")
	entityIDStr := c.Param("id")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		helper.InvalidID(c, "entity")
		return
	}

	logs, err := h.usecase.GetEntityLogs(c.Request.Context(), entityType, entityID)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, logs)
}

// GetByUser handles GET /audit/user/:id
func (h *AuditHandler) GetByUser(c *gin.Context) {
	userID, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "user")
		return
	}

	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	logs, err := h.usecase.GetUserLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{
		"data":      logs,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetByDateRange handles GET /audit/range
func (h *AuditHandler) GetByDateRange(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		helper.BadRequestMessage(c, "invalid start date, use YYYY-MM-DD format")
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		helper.BadRequestMessage(c, "invalid end date, use YYYY-MM-DD format")
		return
	}

	// Include full end day
	end = end.Add(24*time.Hour - time.Second)

	logs, err := h.usecase.GetLogsByDateRange(c.Request.Context(), helper.GetCompanyID(c), start, end)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, logs)
}

// Summary handles GET /audit/summary
func (h *AuditHandler) Summary(c *gin.Context) {
	summary, err := h.usecase.GetAuditSummary(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, summary)
}
