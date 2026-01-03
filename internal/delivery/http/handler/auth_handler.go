package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	usecase    *auth.AuthUsecase
	jwtService *auth.JWTService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(uc *auth.AuthUsecase, jwt *auth.JWTService) *AuthHandler {
	return &AuthHandler{usecase: uc, jwtService: jwt}
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
		helper.BadRequest(c, err)
		return
	}

	input := auth.RegisterInput{
		CompanyName: req.CompanyName,
		Email:       req.Email,
		Password:    req.Password,
		Name:        req.Name,
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	}

	result, err := h.usecase.Register(c.Request.Context(), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	input := auth.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	result, err := h.usecase.Login(c.Request.Context(), input)
	if err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "account is locked due to too many failed attempts" {
			status = http.StatusLocked
		}
		helper.ErrorMessage(c, status, err.Error())
		return
	}

	helper.Success(c, result)
}

// Me handles GET /api/me
func (h *AuthHandler) Me(c *gin.Context) {
	helper.Success(c, gin.H{
		"user_id":    c.GetString("user_id"),
		"company_id": c.GetString("company_id"),
		"email":      c.GetString("email"),
		"role":       c.GetString("role"),
	})
}

// UpdateProfileRequest represents update profile request
type UpdateProfileRequest struct {
	Name           string  `json:"name"`
	ProfilePicture *string `json:"profile_picture"`
}

// UpdateProfile handles PUT /api/me/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	userID := helper.GetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.BadRequestMessage(c, "invalid user id")
		return
	}

	input := auth.UpdateProfileInput{
		Name:           req.Name,
		ProfilePicture: req.ProfilePicture,
	}

	user, err := h.usecase.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, user)
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword handles PUT /api/me/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	userID := helper.GetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.BadRequestMessage(c, "invalid user id")
		return
	}

	// Validate new password
	if err := entity.ValidatePassword(req.NewPassword); err != nil {
		helper.BadRequest(c, err)
		return
	}

	err := h.usecase.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Message(c, "Password changed successfully")
}

type UnlockRequest struct {
	Pin string `json:"pin" binding:"required"`
}

func (h *AuthHandler) Unlock(c *gin.Context) {
	var req UnlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	userID := helper.GetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		helper.Unauthorized(c, "invalid session")
		return
	}

	if err := h.usecase.VerifyPin(c.Request.Context(), userID, req.Pin); err != nil {
		helper.Unauthorized(c, "Invalid PIN")
		return
	}

	helper.Message(c, "Unlocked")
}

// RefreshToken handles POST /auth/refresh - generates new access token from valid token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 7 {
		helper.Unauthorized(c, "missing authorization header")
		return
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix

	// Validate current token
	claims, err := h.jwtService.ValidateToken(tokenString)
	if err != nil {
		helper.Unauthorized(c, "invalid or expired token")
		return
	}

	// Generate new token
	newToken, err := h.jwtService.RefreshToken(claims)
	if err != nil {
		helper.ErrorMessage(c, 500, "failed to generate new token")
		return
	}

	helper.Success(c, gin.H{
		"token":   newToken,
		"message": "Token refreshed successfully",
	})
}
