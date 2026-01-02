package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/contact"
	"github.com/shopspring/decimal"
)

// ContactHandler handles contact endpoints
type ContactHandler struct {
	usecase *contact.ContactUsecase
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(uc *contact.ContactUsecase) *ContactHandler {
	return &ContactHandler{usecase: uc}
}

// CreateContactRequest represents create contact request
type CreateContactRequest struct {
	Code            string `json:"code" binding:"required"`
	Name            string `json:"name" binding:"required"`
	ContactType     string `json:"contact_type" binding:"required"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	City            string `json:"city"`
	TaxID           string `json:"tax_id"`
	CreditLimit     string `json:"credit_limit"`
	PaymentTermDays int    `json:"payment_term_days"`
}

// Create handles POST /contacts
// @Summary Create a new contact
// @Description Create a new customer or supplier contact
// @Tags contacts
// @Accept json
// @Produce json
// @Param request body CreateContactRequest true "Contact data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/contacts [post]
func (h *ContactHandler) Create(c *gin.Context) {
	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))

	creditLimit := decimal.Zero
	if req.CreditLimit != "" {
		cl, err := decimal.NewFromString(req.CreditLimit)
		if err == nil {
			creditLimit = cl
		}
	}

	input := contact.CreateContactInput{
		CompanyID:       companyID,
		Code:            req.Code,
		Name:            req.Name,
		ContactType:     entity.ContactType(req.ContactType),
		Email:           req.Email,
		Phone:           req.Phone,
		Address:         req.Address,
		City:            req.City,
		TaxID:           req.TaxID,
		CreditLimit:     creditLimit,
		PaymentTermDays: req.PaymentTermDays,
	}

	result, err := h.usecase.CreateContact(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, contact.ErrContactCodeExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// List handles GET /contacts
// @Summary List contacts
// @Description Get list of contacts with optional filtering
// @Tags contacts
// @Accept json
// @Produce json
// @Param type query string false "Contact type filter (CUSTOMER, SUPPLIER, BOTH)"
// @Param q query string false "Search by name or code"
// @Param active query boolean false "Filter by active status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/contacts [get]
func (h *ContactHandler) List(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil {
			pageSize = parsed
		}
	}

	input := contact.ListInput{
		CompanyID: companyID,
		Search:    c.Query("q"),
		Page:      page,
		PageSize:  pageSize,
	}

	// Filter by contact type
	if t := c.Query("type"); t != "" {
		contactType := entity.ContactType(t)
		input.ContactType = &contactType
	}

	// Filter by active status
	if a := c.Query("active"); a != "" {
		active := a == "true"
		input.IsActive = &active
	}

	result, err := h.usecase.List(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"items":       result.Contacts,
			"total":       result.Total,
			"page":        result.Page,
			"page_size":   result.PageSize,
			"total_pages": result.TotalPages,
		},
	})
}

// GetByID handles GET /contacts/:id
// @Summary Get contact by ID
// @Description Get a single contact by its ID
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/contacts/{id} [get]
func (h *ContactHandler) GetByID(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), companyID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateContactRequest represents update contact request
type UpdateContactRequest struct {
	Name            string `json:"name" binding:"required"`
	ContactType     string `json:"contact_type" binding:"required"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	City            string `json:"city"`
	TaxID           string `json:"tax_id"`
	CreditLimit     string `json:"credit_limit"`
	PaymentTermDays int    `json:"payment_term_days"`
	IsActive        bool   `json:"is_active"`
}

// Update handles PUT /contacts/:id
// @Summary Update a contact
// @Description Update an existing contact
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Param request body UpdateContactRequest true "Contact data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/contacts/{id} [put]
func (h *ContactHandler) Update(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}

	var req UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creditLimit := decimal.Zero
	if req.CreditLimit != "" {
		cl, err := decimal.NewFromString(req.CreditLimit)
		if err == nil {
			creditLimit = cl
		}
	}

	input := contact.UpdateContactInput{
		CompanyID:       companyID,
		ID:              id,
		Name:            req.Name,
		ContactType:     entity.ContactType(req.ContactType),
		Email:           req.Email,
		Phone:           req.Phone,
		Address:         req.Address,
		City:            req.City,
		TaxID:           req.TaxID,
		CreditLimit:     creditLimit,
		PaymentTermDays: req.PaymentTermDays,
		IsActive:        req.IsActive,
	}

	result, err := h.usecase.UpdateContact(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, contact.ErrContactNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete handles DELETE /contacts/:id
// @Summary Delete a contact
// @Description Soft delete a contact (sets is_active to false)
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/contacts/{id} [delete]
func (h *ContactHandler) Delete(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}

	if err := h.usecase.DeleteContact(c.Request.Context(), companyID, id); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, contact.ErrContactNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contact deleted"})
}
