package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	budgetuc "github.com/herman-xphp/bukuo/internal/usecase/budget"
	"github.com/shopspring/decimal"
)

// BudgetHandler handles budget HTTP endpoints
type BudgetHandler struct {
	usecase *budgetuc.BudgetUsecase
}

// NewBudgetHandler creates a new BudgetHandler
func NewBudgetHandler(uc *budgetuc.BudgetUsecase) *BudgetHandler {
	return &BudgetHandler{usecase: uc}
}

type createBudgetRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description"`
	PeriodID    string              `json:"period_id" binding:"required"`
	StartDate   string              `json:"start_date" binding:"required"`
	EndDate     string              `json:"end_date" binding:"required"`
	Lines       []budgetLineRequest `json:"lines" binding:"required,min=1"`
}

type budgetLineRequest struct {
	AccountID    string  `json:"account_id" binding:"required"`
	BudgetAmount float64 `json:"budget_amount" binding:"required,gt=0"`
	Notes        string  `json:"notes"`
}

// Create handles POST /budgets
func (h *BudgetHandler) Create(c *gin.Context) {
	var req createBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	periodID, _ := uuid.Parse(req.PeriodID)
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	lines := make([]budgetuc.BudgetLineInput, len(req.Lines))
	for i, l := range req.Lines {
		accountID, _ := uuid.Parse(l.AccountID)
		lines[i] = budgetuc.BudgetLineInput{
			AccountID:    accountID,
			BudgetAmount: decimal.NewFromFloat(l.BudgetAmount),
			Notes:        l.Notes,
		}
	}

	input := budgetuc.CreateBudgetInput{
		Name:        req.Name,
		Description: req.Description,
		PeriodID:    periodID,
		StartDate:   startDate,
		EndDate:     endDate,
		Lines:       lines,
	}

	result, err := h.usecase.CreateBudget(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// List handles GET /budgets
func (h *BudgetHandler) List(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	result, total, err := h.usecase.ListBudgets(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, page, pageSize)
}

// Get handles GET /budgets/:id
func (h *BudgetHandler) Get(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "budget")
		return
	}

	result, err := h.usecase.GetBudget(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "budget")
		return
	}

	helper.Success(c, result)
}

// Approve handles POST /budgets/:id/approve
func (h *BudgetHandler) Approve(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "budget")
		return
	}

	if err := h.usecase.ApproveBudget(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "budget approved"})
}

// Activate handles POST /budgets/:id/activate
func (h *BudgetHandler) Activate(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "budget")
		return
	}

	if err := h.usecase.ActivateBudget(c.Request.Context(), helper.GetCompanyID(c), id); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "budget activated"})
}

// Variance handles GET /budgets/:id/variance
func (h *BudgetHandler) Variance(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "budget")
		return
	}

	result, err := h.usecase.GetBudgetVsActual(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}
