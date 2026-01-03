package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/usecase/journal"
)

// JournalHandler handles HTTP requests for journals
type JournalHandler struct {
	usecase *journal.JournalUsecase
}

// NewJournalHandler creates a new JournalHandler
func NewJournalHandler(uc *journal.JournalUsecase) *JournalHandler {
	return &JournalHandler{usecase: uc}
}

// CreateJournalRequest represents the request body for creating a journal
type CreateJournalRequest struct {
	EntryDate   string               `json:"entry_date" binding:"required"`
	Description string               `json:"description" binding:"required"`
	Lines       []JournalLineRequest `json:"lines" binding:"required,min=2"`
}

// JournalLineRequest represents a journal line in the request
type JournalLineRequest struct {
	AccountID    string `json:"account_id" binding:"required"`
	Description  string `json:"description"`
	DebitAmount  string `json:"debit_amount"`
	CreditAmount string `json:"credit_amount"`
}

// Create handles POST /journals
func (h *JournalHandler) Create(c *gin.Context) {
	var req CreateJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		helper.BadRequestMessage(c, "invalid date format, use YYYY-MM-DD")
		return
	}

	// Build input
	input := journal.CreateJournalInput{
		CompanyID:   helper.GetCompanyID(c),
		EntryDate:   entryDate,
		Description: req.Description,
		CreatedBy:   helper.GetUserID(c),
	}

	// Parse lines
	for _, l := range req.Lines {
		accountID := helper.ParseUUIDString(l.AccountID)
		if accountID.String() == "00000000-0000-0000-0000-000000000000" {
			helper.BadRequestMessage(c, "invalid account_id")
			return
		}

		input.Lines = append(input.Lines, journal.JournalLineInput{
			AccountID:    accountID,
			Description:  l.Description,
			DebitAmount:  helper.ParseDecimal(l.DebitAmount),
			CreditAmount: helper.ParseDecimal(l.CreditAmount),
		})
	}

	result, err := h.usecase.CreateJournal(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// GetByID handles GET /journals/:id
func (h *JournalHandler) GetByID(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "journal")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		helper.NotFound(c, "journal")
		return
	}

	helper.Success(c, result)
}

// Post handles POST /journals/:id/post
func (h *JournalHandler) Post(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "journal")
		return
	}

	result, err := h.usecase.PostJournal(c.Request.Context(), id, helper.GetUserID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "Journal posted successfully", "data": result})
}

// Reverse handles POST /journals/:id/reverse
func (h *JournalHandler) Reverse(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "journal")
		return
	}

	var req struct {
		ReversalDate string `json:"reversal_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	reversalDate, _ := time.Parse("2006-01-02", req.ReversalDate)

	result, err := h.usecase.ReverseJournal(c.Request.Context(), id, helper.GetUserID(c), reversalDate)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// SubmitForApproval handles POST /journals/:id/submit-approval
func (h *JournalHandler) SubmitForApproval(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "journal")
		return
	}

	result, err := h.usecase.SubmitForApproval(c.Request.Context(), id, helper.GetUserID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "Journal submitted for approval", "data": result})
}

// Approve handles POST /journals/:id/approve
func (h *JournalHandler) Approve(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "journal")
		return
	}

	result, err := h.usecase.ApproveJournal(c.Request.Context(), id, helper.GetUserID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "Journal approved", "data": result})
}

// Reject handles POST /journals/:id/reject
func (h *JournalHandler) Reject(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "journal")
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	result, err := h.usecase.RejectJournal(c.Request.Context(), id, helper.GetUserID(c), req.Reason)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "Journal rejected", "data": result})
}

// PendingApprovals handles GET /journals/pending
func (h *JournalHandler) PendingApprovals(c *gin.Context) {
	result, err := h.usecase.GetPendingApprovals(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, gin.H{"data": result, "count": len(result)})
}

// List handles GET /journals
func (h *JournalHandler) List(c *gin.Context) {
	p := helper.ParsePagination(c)
	search := c.Query("q")

	result, total, err := h.usecase.List(c.Request.Context(), helper.GetCompanyID(c), p.Limit, p.Offset, search)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.PaginatedItems(c, result, int64(total), p)
}

// Update handles PUT /journals/:id
func (h *JournalHandler) Update(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "journal")
		return
	}

	var req CreateJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		helper.BadRequestMessage(c, "invalid date format, use YYYY-MM-DD")
		return
	}

	input := journal.CreateJournalInput{
		CompanyID:   helper.GetCompanyID(c),
		EntryDate:   entryDate,
		Description: req.Description,
		CreatedBy:   helper.GetUserID(c),
	}

	for _, l := range req.Lines {
		accountID := helper.ParseUUIDString(l.AccountID)
		if accountID.String() == "00000000-0000-0000-0000-000000000000" {
			helper.BadRequestMessage(c, "invalid account_id")
			return
		}

		input.Lines = append(input.Lines, journal.JournalLineInput{
			AccountID:    accountID,
			Description:  l.Description,
			DebitAmount:  helper.ParseDecimal(l.DebitAmount),
			CreditAmount: helper.ParseDecimal(l.CreditAmount),
		})
	}

	result, err := h.usecase.UpdateJournal(c.Request.Context(), id, input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}
