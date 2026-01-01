package http

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/handler"
)

// SetupRouter configures all routes
func SetupRouter(
	r *gin.Engine,
	healthHandler *handler.HealthHandler,
	journalHandler *handler.JournalHandler,
) {
	// Health endpoints
	r.GET("/health", healthHandler.Health)
	r.GET("/ping", healthHandler.Ping)

	// API routes
	api := r.Group("/api")
	{
		// Journal routes
		journals := api.Group("/journals")
		{
			journals.POST("", journalHandler.Create)
			journals.GET("/:id", journalHandler.GetByID)
			journals.POST("/:id/post", journalHandler.Post)
		}
	}
}
