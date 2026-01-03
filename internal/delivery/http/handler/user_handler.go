package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/user"
)

type UserHandler struct {
	usecase *user.UserUsecase
}

func NewUserHandler(uc *user.UserUsecase) *UserHandler {
	return &UserHandler{usecase: uc}
}

type CreateUserRequest struct {
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8"`
	Name     string          `json:"name" binding:"required"`
	Role     entity.UserRole `json:"role" binding:"required"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	// Check if current user has permission
	role := entity.UserRole(c.GetString("role"))
	if role != entity.UserRoleOwner && role != entity.UserRoleAdmin {
		helper.Forbidden(c, "only owners and admins can create users")
		return
	}

	input := user.CreateUserInput{
		CompanyID: helper.GetCompanyID(c),
		Email:     req.Email,
		Password:  req.Password,
		Name:      req.Name,
		Role:      req.Role,
	}

	res, err := h.usecase.CreateUser(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, res)
}

func (h *UserHandler) List(c *gin.Context) {
	companyID := helper.GetCompanyID(c)

	// Check if pagination/search
	if c.Query("limit") != "" || c.Query("q") != "" || c.Query("offset") != "" || c.Query("page") != "" {
		p := helper.ParsePagination(c)
		search := c.Query("q")

		users, total, err := h.usecase.List(c.Request.Context(), companyID, p.Limit, p.Offset, search)
		if err != nil {
			helper.InternalError(c, err)
			return
		}

		helper.PaginatedItems(c, users, int64(total), p)
		return
	}

	// Default: Return Limitless (Backwards compatible)
	users, err := h.usecase.ListByCompany(c.Request.Context(), companyID)
	if err != nil {
		helper.InternalError(c, err)
		return
	}

	helper.Success(c, users)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "user")
		return
	}

	companyID := helper.GetCompanyID(c)

	// Check if current user has permission
	role := entity.UserRole(c.GetString("role"))
	if role != entity.UserRoleOwner && role != entity.UserRoleAdmin {
		helper.Forbidden(c, "only owners and admins can delete users")
		return
	}

	if err := h.usecase.DeleteUser(c.Request.Context(), id, companyID); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Message(c, "User deleted successfully")
}

type UpdateUserRequest struct {
	Name           string          `json:"name"`
	Role           entity.UserRole `json:"role"`
	IsActive       *bool           `json:"is_active"`
	ProfilePicture *string         `json:"profile_picture"`
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "user")
		return
	}

	u, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		helper.NotFound(c, "user")
		return
	}

	// Verify company access
	companyID := helper.GetCompanyID(c)
	if u.CompanyID != companyID {
		helper.NotFound(c, "user")
		return
	}

	helper.Success(c, u)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "user")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	companyID := helper.GetCompanyID(c)

	// Check permission
	role := entity.UserRole(c.GetString("role"))
	if role != entity.UserRoleOwner && role != entity.UserRoleAdmin {
		helper.Forbidden(c, "only owners and admins can update users")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	input := user.UpdateUserInput{
		Name:           req.Name,
		Role:           req.Role,
		IsActive:       isActive,
		ProfilePicture: req.ProfilePicture,
	}

	updated, err := h.usecase.UpdateUser(c.Request.Context(), id, companyID, input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, updated)
}

type ResetPasswordRequest struct{}

const DefaultPassword = "P@ssw0rd123"

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "user")
		return
	}

	companyID := helper.GetCompanyID(c)

	// Check permission
	role := entity.UserRole(c.GetString("role"))
	if role != entity.UserRoleOwner && role != entity.UserRoleAdmin {
		helper.Forbidden(c, "only owners and admins can reset passwords")
		return
	}

	if err := h.usecase.ResetPassword(c.Request.Context(), id, companyID, DefaultPassword); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{
		"message":          "Password reset successfully",
		"default_password": DefaultPassword,
	})
}

type SetPinRequest struct {
	Pin string `json:"pin" binding:"required,min=4,max=6"`
}

func (h *UserHandler) SetPin(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "user")
		return
	}

	companyID := helper.GetCompanyID(c)

	// Check permission (Owner/Admin)
	role := entity.UserRole(c.GetString("role"))
	if role != entity.UserRoleOwner && role != entity.UserRoleAdmin {
		helper.Forbidden(c, "only owners and admins can set pins")
		return
	}

	var req SetPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	if err := h.usecase.SetUserPin(c.Request.Context(), id, companyID, req.Pin); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Message(c, "User PIN updated successfully")
}
