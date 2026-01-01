package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/usecase/journal"
	"github.com/shopspring/decimal"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get context values (from middleware)
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	userID, _ := uuid.Parse(c.GetString("user_id"))

	// Parse date
	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	// Build input
	input := journal.CreateJournalInput{
		CompanyID:   companyID,
		EntryDate:   entryDate,
		Description: req.Description,
		CreatedBy:   userID,
	}

	// Parse lines
	for _, l := range req.Lines {
		accountID, err := uuid.Parse(l.AccountID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
			return
		}

		debit, _ := decimal.NewFromString(l.DebitAmount)
		credit, _ := decimal.NewFromString(l.CreditAmount)

		input.Lines = append(input.Lines, journal.JournalLineInput{
			AccountID:    accountID,
			Description:  l.Description,
			DebitAmount:  debit,
			CreditAmount: credit,
		})
	}

	// Create journal
	result, err := h.usecase.CreateJournal(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Journal created successfully",
		"data":    result,
	})
}

// GetByID handles GET /journals/:id
func (h *JournalHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "journal not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Post handles POST /journals/:id/post
func (h *JournalHandler) Post(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal id"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))

	result, err := h.usecase.PostJournal(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Journal posted successfully",
		"data":    result,
	})
}

// Reverse handles POST /journals/:id/reverse
func (h *JournalHandler) Reverse(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal id"})
		return
	}

	var req struct {
		ReversalDate string `json:"reversal_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reversalDate, _ := time.Parse("2006-01-02", req.ReversalDate)
	userID, _ := uuid.Parse(c.GetString("user_id"))

	result, err := h.usecase.ReverseJournal(c.Request.Context(), id, userID, reversalDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Journal reversed successfully",
		"data":    result,
	})
}

// SubmitForApproval handles POST /journals/:id/submit-approval
func (h *JournalHandler) SubmitForApproval(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal id"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))

	result, err := h.usecase.SubmitForApproval(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Journal submitted for approval",
		"data":    result,
	})
}

// Approve handles POST /journals/:id/approve
func (h *JournalHandler) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal id"})
		return
	}

	approverID, _ := uuid.Parse(c.GetString("user_id"))

	result, err := h.usecase.ApproveJournal(c.Request.Context(), id, approverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Journal approved",
		"data":    result,
	})
}

// Reject handles POST /journals/:id/reject
func (h *JournalHandler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal id"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rejectorID, _ := uuid.Parse(c.GetString("user_id"))

	result, err := h.usecase.RejectJournal(c.Request.Context(), id, rejectorID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Journal rejected",
		"data":    result,
	})
}

// PendingApprovals handles GET /journals/pending
func (h *JournalHandler) PendingApprovals(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	result, err := h.usecase.GetPendingApprovals(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"count": len(result),
	})
}
