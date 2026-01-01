package main

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/herman-xphp/bukuo/internal/config"
	"github.com/herman-xphp/bukuo/internal/database"
	httpDelivery "github.com/herman-xphp/bukuo/internal/delivery/http"
	"github.com/herman-xphp/bukuo/internal/delivery/http/handler"
	"github.com/herman-xphp/bukuo/internal/delivery/http/middleware"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres"
	accountUC "github.com/herman-xphp/bukuo/internal/usecase/account"
	authUC "github.com/herman-xphp/bukuo/internal/usecase/auth"
	closingUC "github.com/herman-xphp/bukuo/internal/usecase/closing"
	journalUC "github.com/herman-xphp/bukuo/internal/usecase/journal"
	periodUC "github.com/herman-xphp/bukuo/internal/usecase/period"
	reportUC "github.com/herman-xphp/bukuo/internal/usecase/report"

	_ "github.com/herman-xphp/bukuo/docs" // Swagger docs
)

// @title Bukuo API
// @version 1.0
// @description Financial Accounting Platform API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@bukuo.id

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter: Bearer {token}

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
	auditLogRepo := postgres.NewAuditLogRepository(db)

	// JWT Service
	jwtService := authUC.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Usecase Layer - Business Logic
	authUsecase := authUC.NewAuthUsecase(userRepo, companyRepo, jwtService, auditLogRepo, cfg.Security)
	accountUsecase := accountUC.NewAccountUsecase(accountRepo)
	periodUsecase := periodUC.NewPeriodUsecase(periodRepo)
	journalUsecase := journalUC.NewJournalUsecase(journalRepo, accountRepo, periodRepo)
	reportUsecase := reportUC.NewReportUsecase(journalRepo, accountRepo, periodRepo)
	closingUsecase := closingUC.NewClosingUsecase(journalRepo, accountRepo, periodRepo)

	// Delivery Layer - HTTP Handlers
	handlers := &httpDelivery.Handlers{
		Health:  handler.NewHealthHandler(),
		Auth:    handler.NewAuthHandler(authUsecase),
		Account: handler.NewAccountHandler(accountUsecase),
		Period:  handler.NewPeriodHandler(periodUsecase),
		Journal: handler.NewJournalHandler(journalUsecase),
		Report:  handler.NewReportHandler(reportUsecase),
		Closing: handler.NewClosingHandler(closingUsecase),
	}

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	rateLimiter := middleware.NewRateLimiter(cfg.Security.RateLimitPerMin, time.Minute)

	// ============================================
	// ROUTER SETUP
	// ============================================

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(middleware.RecoveryHandler())
	r.Use(corsMiddleware(cfg.Security.AllowedOrigins, cfg.Server.Env))
	r.Use(rateLimiter.RateLimitMiddleware())

	// 404 handler
	r.NoRoute(middleware.NotFoundHandler())

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup routes with injected dependencies
	httpDelivery.SetupRouter(r, handlers, authMiddleware)

	// Start server
	log.Printf("🚀 Bukuo running on port %s (%s)", cfg.Server.Port, cfg.Server.Env)
	log.Printf("📚 Swagger: http://localhost:%s/swagger/index.html", cfg.Server.Port)
	log.Printf("🔒 CORS: %v", cfg.Security.AllowedOrigins)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}

// corsMiddleware handles CORS with whitelist
func corsMiddleware(allowedOrigins []string, env string) gin.HandlerFunc {
	// Build origin map for O(1) lookup
	originMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originMap[strings.TrimSpace(origin)] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// In development, allow all. In production, check whitelist.
		if env != "production" || originMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
