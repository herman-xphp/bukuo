package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents standardized error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// ErrorHandler provides centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
		}
	}
}

// RecoveryHandler handles panics
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v\n%s", r, debug.Stack())

				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Error:   "Internal server error",
					Code:    "INTERNAL_ERROR",
				})
			}
		}()
		c.Next()
	}
}

// NotFoundHandler handles 404
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "Endpoint not found",
			Code:    "NOT_FOUND",
		})
	}
}
