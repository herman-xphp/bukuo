package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/usecase/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	usecase *auth.AuthUsecase
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(uc *auth.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: uc}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	CompanyName string `json:"company_name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Name        string `json:"name" binding:"required"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := auth.RegisterInput{
		CompanyName: req.CompanyName,
		Email:       req.Email,
		Password:    req.Password,
		Name:        req.Name,
	}

	result, err := h.usecase.Register(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"data":    result,
	})
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.usecase.Login(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    result,
	})
}

// Me handles GET /auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"company_id": c.GetString("company_id"),
		"email":      c.GetString("email"),
		"role":       c.GetString("role"),
	})
}
