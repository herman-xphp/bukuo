package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))

	// Check if current user has permission
	role := entity.UserRole(c.GetString("role"))
	if role != entity.UserRoleOwner && role != entity.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owners and admins can create users"})
		return
	}

	input := user.CreateUserInput{
		CompanyID: companyID,
		Email:     req.Email,
		Password:  req.Password,
		Name:      req.Name,
		Role:      req.Role,
	}

	res, err := h.usecase.CreateUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"data":    res,
	})
}

func (h *UserHandler) List(c *gin.Context) {
	companyID, _ := uuid.Parse(c.GetString("company_id"))

	// Check if pagination/search
	if c.Query("limit") != "" || c.Query("q") != "" || c.Query("offset") != "" || c.Query("page") != "" {
		limit := 50
		page := 1
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}
		if p := c.Query("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil {
				page = parsed
			}
		}
		offset := (page - 1) * limit
		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}
		search := c.Query("q")

		users, total, err := h.usecase.List(c.Request.Context(), companyID, limit, offset, search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"items":  users,
				"total":  total,
				"limit":  limit,
				"page":   page,
				"offset": offset,
			},
		})
		return
	}

	// Default: Return Limitless (Backwards compatible)
	users, err := h.usecase.ListByCompany(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	companyID, _ := uuid.Parse(c.GetString("company_id"))

	// Check if current user has permission (Owner only for deletion usually, or Admin for non-admins)
	role := entity.UserRole(c.GetString("role"))
	if role != entity.UserRoleOwner && role != entity.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owners and admins can delete users"})
		return
	}

	if err := h.usecase.DeleteUser(c.Request.Context(), id, companyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
