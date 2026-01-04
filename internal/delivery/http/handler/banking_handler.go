package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	bankinguc "github.com/herman-xphp/bukuo/internal/usecase/banking"
	"github.com/shopspring/decimal"
)

// BankingHandler handles banking HTTP endpoints
type BankingHandler struct {
	usecase *bankinguc.BankingUsecase
}

// NewBankingHandler creates a new BankingHandler
func NewBankingHandler(uc *bankinguc.BankingUsecase) *BankingHandler {
	return &BankingHandler{usecase: uc}
}

type createBankAccountRequest struct {
	AccountID      string  `json:"account_id" binding:"required"`
	BankName       string  `json:"bank_name" binding:"required"`
	AccountNumber  string  `json:"account_number" binding:"required"`
	AccountName    string  `json:"account_name" binding:"required"`
	CurrencyID     string  `json:"currency_id" binding:"required"`
	OpeningBalance float64 `json:"opening_balance"`
}

// CreateAccount handles POST /banking/accounts
func (h *BankingHandler) CreateAccount(c *gin.Context) {
	var req createBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	accountID, _ := uuid.Parse(req.AccountID)
	currencyID, _ := uuid.Parse(req.CurrencyID)

	input := bankinguc.CreateBankAccountInput{
		AccountID:      accountID,
		BankName:       req.BankName,
		AccountNumber:  req.AccountNumber,
		AccountName:    req.AccountName,
		CurrencyID:     currencyID,
		OpeningBalance: decimal.NewFromFloat(req.OpeningBalance),
	}

	result, err := h.usecase.CreateBankAccount(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// ListAccounts handles GET /banking/accounts
func (h *BankingHandler) ListAccounts(c *gin.Context) {
	result, err := h.usecase.ListBankAccounts(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

// GetAccount handles GET /banking/accounts/:id
func (h *BankingHandler) GetAccount(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "bank account")
		return
	}

	result, err := h.usecase.GetBankAccount(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "bank account")
		return
	}

	helper.Success(c, result)
}

type createTransactionRequest struct {
	BankAccountID   string  `json:"bank_account_id" binding:"required"`
	TransactionDate string  `json:"transaction_date" binding:"required"`
	Type            string  `json:"type" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Description     string  `json:"description" binding:"required"`
	Reference       string  `json:"reference"`
}

// CreateTransaction handles POST /banking/transactions
func (h *BankingHandler) CreateTransaction(c *gin.Context) {
	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	bankAccID, _ := uuid.Parse(req.BankAccountID)
	txnDate, _ := time.Parse("2006-01-02", req.TransactionDate)

	input := bankinguc.CreateTransactionInput{
		BankAccountID:   bankAccID,
		TransactionDate: txnDate,
		Type:            entity.BankTransactionType(req.Type),
		Amount:          decimal.NewFromFloat(req.Amount),
		Description:     req.Description,
		Reference:       req.Reference,
	}

	result, err := h.usecase.CreateTransaction(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// ListTransactions handles GET /banking/accounts/:id/transactions
func (h *BankingHandler) ListTransactions(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "bank account")
		return
	}

	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	result, total, err := h.usecase.ListTransactions(c.Request.Context(), helper.GetCompanyID(c), id, page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, page, pageSize)
}

// ReconcileTransaction handles POST /banking/transactions/:id/reconcile
func (h *BankingHandler) ReconcileTransaction(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "transaction")
		return
	}

	if err := h.usecase.ReconcileTransaction(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "transaction reconciled"})
}
