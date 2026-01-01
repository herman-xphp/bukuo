package http

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/handler"
	"github.com/herman-xphp/bukuo/internal/delivery/http/middleware"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	Health  *handler.HealthHandler
	Auth    *handler.AuthHandler
	Journal *handler.JournalHandler
	Account *handler.AccountHandler
	Period  *handler.PeriodHandler
}

// SetupRouter configures all routes
func SetupRouter(r *gin.Engine, h *Handlers, authMW *middleware.AuthMiddleware) {
	// Public routes
	r.GET("/health", h.Health.Health)
	r.GET("/ping", h.Health.Ping)

	// Auth routes (public)
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
	}

	// Protected API routes
	api := r.Group("/api")
	api.Use(authMW.Authenticate())
	{
		// Auth
		api.GET("/me", h.Auth.Me)

		// Accounts
		accounts := api.Group("/accounts")
		{
			accounts.POST("", h.Account.Create)
			accounts.GET("", h.Account.GetAll)
			accounts.GET("/:id", h.Account.GetByID)
			accounts.PUT("/:id", h.Account.Update)
			accounts.DELETE("/:id", h.Account.Delete)
		}

		// Periods
		periods := api.Group("/periods")
		{
			periods.POST("", h.Period.Create)
			periods.GET("", h.Period.GetAll)
			periods.GET("/:id", h.Period.GetByID)
			periods.POST("/:id/close", h.Period.Close)
		}

		// Journals
		journals := api.Group("/journals")
		{
			journals.POST("", h.Journal.Create)
			journals.GET("/:id", h.Journal.GetByID)
			journals.POST("/:id/post", h.Journal.Post)
		}
	}
}
