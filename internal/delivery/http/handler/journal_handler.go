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
