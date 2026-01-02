package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	categoryUC "github.com/herman-xphp/bukuo/internal/usecase/category"
	closingUC "github.com/herman-xphp/bukuo/internal/usecase/closing"
	contactUC "github.com/herman-xphp/bukuo/internal/usecase/contact"
	currencyUC "github.com/herman-xphp/bukuo/internal/usecase/currency"
	exchangerateUC "github.com/herman-xphp/bukuo/internal/usecase/exchangerate"
	inventoryUC "github.com/herman-xphp/bukuo/internal/usecase/inventory"
	journalUC "github.com/herman-xphp/bukuo/internal/usecase/journal"
	openingUC "github.com/herman-xphp/bukuo/internal/usecase/opening"
	periodUC "github.com/herman-xphp/bukuo/internal/usecase/period"
	productUC "github.com/herman-xphp/bukuo/internal/usecase/product"
	reportUC "github.com/herman-xphp/bukuo/internal/usecase/report"
	salesUC "github.com/herman-xphp/bukuo/internal/usecase/sales"
	unitUC "github.com/herman-xphp/bukuo/internal/usecase/unit"
	userUC "github.com/herman-xphp/bukuo/internal/usecase/user"
	warehouseUC "github.com/herman-xphp/bukuo/internal/usecase/warehouse"

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

	// Setup structured logger
	logFormat := "text"
	if cfg.Server.Env == "production" {
		logFormat = "json"
	}
	logger := middleware.SetupLogger(middleware.LoggerConfig{
		Format:      logFormat,
		Level:       "info",
		Environment: cfg.Server.Env,
	})

	logger.Info("Starting Bukuo API",
		slog.String("env", cfg.Server.Env),
		slog.String("port", cfg.Server.Port),
	)

	// Connect database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.Error("Failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Database connected successfully")

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
	contactRepo := postgres.NewContactRepository(db)
	unitRepo := postgres.NewUnitRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	productRepo := postgres.NewProductRepository(db)
	currencyRepo := postgres.NewCurrencyRepository(db)
	exchangeRateRepo := postgres.NewExchangeRateRepository(db)
	warehouseRepo := postgres.NewWarehouseRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)

	// Transaction Manager
	txManager := postgres.NewTransactionManager(db)

	// JWT Service
	jwtService := authUC.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Usecase Layer - Business Logic
	authUsecase := authUC.NewAuthUsecase(userRepo, companyRepo, jwtService, auditLogRepo, txManager, cfg.Security)
	accountUsecase := accountUC.NewAccountUsecase(accountRepo, journalRepo)
	periodUsecase := periodUC.NewPeriodUsecase(periodRepo)
	journalUsecase := journalUC.NewJournalUsecase(journalRepo, accountRepo, periodRepo, auditLogRepo)
	reportUsecase := reportUC.NewReportUsecase(journalRepo, accountRepo, periodRepo)
	closingUsecase := closingUC.NewClosingUsecase(journalRepo, accountRepo, periodRepo, auditLogRepo)
	openingUsecase := openingUC.NewOpeningBalanceUsecase(journalRepo, accountRepo, periodRepo)
	userUsecase := userUC.NewUserUsecase(userRepo)
	contactUsecase := contactUC.NewContactUsecase(contactRepo)
	unitUsecase := unitUC.NewUnitUsecase(unitRepo)
	categoryUsecase := categoryUC.NewCategoryUsecase(categoryRepo)
	productUsecase := productUC.NewProductUsecase(productRepo)
	currencyUsecase := currencyUC.NewCurrencyUsecase(currencyRepo)
	exchangerateUsecase := exchangerateUC.NewExchangeRateUsecase(exchangeRateRepo, currencyRepo)
	warehouseUsecase := warehouseUC.NewWarehouseUsecase(warehouseRepo)
	inventoryUsecase := inventoryUC.NewInventoryUsecase(inventoryRepo, warehouseRepo, productRepo)
	salesUsecase := salesUC.NewSalesUsecase()

	// Delivery Layer - HTTP Handlers
	handlers := &httpDelivery.Handlers{
		Health:       handler.NewHealthHandler(),
		Auth:         handler.NewAuthHandler(authUsecase),
		Account:      handler.NewAccountHandler(accountUsecase),
		Period:       handler.NewPeriodHandler(periodUsecase),
		Journal:      handler.NewJournalHandler(journalUsecase),
		Report:       handler.NewReportHandler(reportUsecase),
		Closing:      handler.NewClosingHandler(closingUsecase),
		Opening:      handler.NewOpeningHandler(openingUsecase),
		User:         handler.NewUserHandler(userUsecase),
		Contact:      handler.NewContactHandler(contactUsecase),
		Unit:         handler.NewUnitHandler(unitUsecase),
		Category:     handler.NewCategoryHandler(categoryUsecase),
		Product:      handler.NewProductHandler(productUsecase),
		Currency:     handler.NewCurrencyHandler(currencyUsecase),
		ExchangeRate: handler.NewExchangeRateHandler(exchangerateUsecase),
		Warehouse:    handler.NewWarehouseHandler(warehouseUsecase),
		Inventory:    handler.NewInventoryHandler(inventoryUsecase),
		Sales:        handler.NewSalesHandler(salesUsecase),
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
	r.Use(middleware.RequestLogger(logger)) // Structured logging
	r.Use(middleware.RecoveryHandler(cfg.Server.Env))
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.Security.AllowedOrigins,
		Environment:    cfg.Server.Env,
	}
	r.Use(middleware.NewCORSMiddleware(corsConfig))
	r.Use(rateLimiter.RateLimitMiddleware())

	// 404 handler
	r.NoRoute(middleware.NotFoundHandler())

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup routes with injected dependencies
	httpDelivery.SetupRouter(r, handlers, authMiddleware)

	// ============================================
	// GRACEFUL SHUTDOWN
	// ============================================

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("🚀 Bukuo API started",
			slog.String("port", cfg.Server.Port),
			slog.String("swagger", "http://localhost:"+cfg.Server.Port+"/swagger/index.html"),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("⏳ Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}

	// Close database connection
	db.Close()
	logger.Info("✅ Database connection closed")

	logger.Info("👋 Server exited gracefully")
}
