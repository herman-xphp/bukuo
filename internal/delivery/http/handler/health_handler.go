package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
)

// HealthHandler handles health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	helper.Success(c, gin.H{
		"status":  "ok",
		"service": "bukuo",
	})
}

// Ping handles GET /ping
func (h *HealthHandler) Ping(c *gin.Context) {
	helper.Message(c, "pong")
}
