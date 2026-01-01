package routes

import "github.com/gin-gonic/gin"

func Setup(r *gin.Engine) {
	// Health check
	r.GET("/health", healthCheck)

	// API routes
	api := r.Group("/api")
	{
		api.GET("/ping", ping)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "bukuo",
	})
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
