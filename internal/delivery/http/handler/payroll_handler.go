package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	payrolluc "github.com/herman-xphp/bukuo/internal/usecase/payroll"
	"github.com/shopspring/decimal"
)

type PayrollHandler struct {
	usecase *payrolluc.PayrollUsecase
}

func NewPayrollHandler(uc *payrolluc.PayrollUsecase) *PayrollHandler {
	return &PayrollHandler{usecase: uc}
}

// -- Employee --

type createEmployeeRequest struct {
	FirstName   string                 `json:"first_name" binding:"required"`
	LastName    string                 `json:"last_name" binding:"required"`
	Email       string                 `json:"email" binding:"required,email"`
	Phone       string                 `json:"phone"`
	JoinDate    string                 `json:"join_date" binding:"required"`
	Status      string                 `json:"status" binding:"required"`
	JobTitle    string                 `json:"job_title" binding:"required"`
	Department  string                 `json:"department"`
	BasicSalary float64                `json:"basic_salary" binding:"required,gt=0"`
	BankName    string                 `json:"bank_name"`
	BankAccount string                 `json:"bank_account"`
	TaxID       string                 `json:"tax_id"`
	Components  []componentLinkRequest `json:"components"`
}

type componentLinkRequest struct {
	ComponentID string  `json:"component_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
}

func (h *PayrollHandler) CreateEmployee(c *gin.Context) {
	var req createEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	joinDate, _ := time.Parse("2006-01-02", req.JoinDate)

	comps := make([]entity.EmployeeSalaryComponent, len(req.Components))
	for i, cr := range req.Components {
		cID, _ := uuid.Parse(cr.ComponentID)
		comps[i] = entity.EmployeeSalaryComponent{
			ID:          uuid.New(),
			ComponentID: cID,
			Amount:      decimal.NewFromFloat(cr.Amount),
		}
	}

	emp := entity.Employee{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		JoinDate:    joinDate,
		Status:      entity.EmploymentStatus(req.Status),
		JobTitle:    req.JobTitle,
		Department:  req.Department,
		BasicSalary: decimal.NewFromFloat(req.BasicSalary),
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		TaxID:       req.TaxID,
		Components:  comps,
	}

	if err := h.usecase.CreateEmployee(c.Request.Context(), helper.GetCompanyID(c), emp); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, gin.H{"message": "employee created"})
}

func (h *PayrollHandler) ListEmployees(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	emps, total, err := h.usecase.ListEmployees(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithPagination(c, emps, total, page, pageSize)
}

func (h *PayrollHandler) GetEmployee(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "employee")
		return
	}

	emp, err := h.usecase.GetEmployee(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c, emp)
}

// -- Pay Run --

type generatePayRunRequest struct {
	PeriodID    string `json:"period_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
	PaymentDate string `json:"payment_date" binding:"required"`
}

func (h *PayrollHandler) GeneratePayRun(c *gin.Context) {
	var req generatePayRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	pID, _ := uuid.Parse(req.PeriodID)
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	payDate, _ := time.Parse("2006-01-02", req.PaymentDate)

	input := payrolluc.CreatePayRunInput{
		PeriodID:    pID,
		StartDate:   start,
		EndDate:     end,
		PaymentDate: payDate,
	}

	pr, err := h.usecase.GeneratePayRun(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Created(c, pr)
}

func (h *PayrollHandler) GetPayRun(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "pay_run")
		return
	}

	pr, err := h.usecase.GetPayRun(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c, pr)
}

func (h *PayrollHandler) ApprovePayRun(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "pay_run")
		return
	}

	// For now, no bank account input, handled in usecase or simplified
	if err := h.usecase.ApprovePayRun(c.Request.Context(), helper.GetCompanyID(c), id, uuid.Nil); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c, gin.H{"message": "pay run approved"})
}
