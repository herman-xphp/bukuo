package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/config"
	"github.com/herman-xphp/bukuo/internal/database"
	httpDelivery "github.com/herman-xphp/bukuo/internal/delivery/http"
	"github.com/herman-xphp/bukuo/internal/delivery/http/handler"
	"github.com/herman-xphp/bukuo/internal/delivery/http/middleware"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres"
	accountUC "github.com/herman-xphp/bukuo/internal/usecase/account"
	authUC "github.com/herman-xphp/bukuo/internal/usecase/auth"
	journalUC "github.com/herman-xphp/bukuo/internal/usecase/journal"
	periodUC "github.com/herman-xphp/bukuo/internal/usecase/period"
	reportUC "github.com/herman-xphp/bukuo/internal/usecase/report"
)

func main() {
	// Load config
	cfg := config.Load()

	// Connect database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ============================================
	// DEPENDENCY INJECTION (Clean Architecture)
	// ============================================

	// Infrastructure Layer - Repository Implementations
	companyRepo := postgres.NewCompanyRepository(db)
	userRepo := postgres.NewUserRepository(db)
	accountRepo := postgres.NewAccountRepository(db)
	periodRepo := postgres.NewPeriodRepository(db)
	journalRepo := postgres.NewJournalRepository(db)

	// JWT Service
	jwtService := authUC.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Usecase Layer - Business Logic
	authUsecase := authUC.NewAuthUsecase(userRepo, companyRepo, jwtService)
	accountUsecase := accountUC.NewAccountUsecase(accountRepo)
	periodUsecase := periodUC.NewPeriodUsecase(periodRepo)
	journalUsecase := journalUC.NewJournalUsecase(journalRepo, accountRepo, periodRepo)
	reportUsecase := reportUC.NewReportUsecase(journalRepo, accountRepo, periodRepo)

	// Delivery Layer - HTTP Handlers
	handlers := &httpDelivery.Handlers{
		Health:  handler.NewHealthHandler(),
		Auth:    handler.NewAuthHandler(authUsecase),
		Account: handler.NewAccountHandler(accountUsecase),
		Period:  handler.NewPeriodHandler(periodUsecase),
		Journal: handler.NewJournalHandler(journalUsecase),
		Report:  handler.NewReportHandler(reportUsecase),
	}

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// ============================================
	// ROUTER SETUP
	// ============================================

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// CORS middleware
	r.Use(corsMiddleware())

	// Setup routes with injected dependencies
	httpDelivery.SetupRouter(r, handlers, authMiddleware)

	// Start server
	log.Printf("🚀 Bukuo running on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}

// corsMiddleware handles CORS for frontend
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
