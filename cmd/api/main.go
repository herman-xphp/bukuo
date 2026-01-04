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
	"github.com/herman-xphp/bukuo/internal/infrastructure/startup"
	accountUC "github.com/herman-xphp/bukuo/internal/usecase/account"
	arapUC "github.com/herman-xphp/bukuo/internal/usecase/arap"
	authUC "github.com/herman-xphp/bukuo/internal/usecase/auth"
	bankingUC "github.com/herman-xphp/bukuo/internal/usecase/banking"
	categoryUC "github.com/herman-xphp/bukuo/internal/usecase/category"
	closingUC "github.com/herman-xphp/bukuo/internal/usecase/closing"
	contactUC "github.com/herman-xphp/bukuo/internal/usecase/contact"
	currencyUC "github.com/herman-xphp/bukuo/internal/usecase/currency"
	exchangerateUC "github.com/herman-xphp/bukuo/internal/usecase/exchangerate"
	fixedassetUC "github.com/herman-xphp/bukuo/internal/usecase/fixedasset"
	inventoryUC "github.com/herman-xphp/bukuo/internal/usecase/inventory"
	journalUC "github.com/herman-xphp/bukuo/internal/usecase/journal"
	openingUC "github.com/herman-xphp/bukuo/internal/usecase/opening"
	periodUC "github.com/herman-xphp/bukuo/internal/usecase/period"
	productUC "github.com/herman-xphp/bukuo/internal/usecase/product"
	purchasingUC "github.com/herman-xphp/bukuo/internal/usecase/purchasing"
	reportUC "github.com/herman-xphp/bukuo/internal/usecase/report"
	salesUC "github.com/herman-xphp/bukuo/internal/usecase/sales"
	taxUC "github.com/herman-xphp/bukuo/internal/usecase/tax"
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
	// Load config (validation happens inside)
	cfg := config.Load()

	// Set Gin to release mode for clean logs
	gin.SetMode(gin.ReleaseMode)

	// Setup structured logger (silent initially, for internal use)
	logFormat := "text"
	if cfg.Server.Env == "production" {
		logFormat = "json"
	}
	logger := middleware.SetupLogger(middleware.LoggerConfig{
		Format:      logFormat,
		Level:       "info",
		Environment: cfg.Server.Env,
	})

	// Connect database
	db, err := database.Connect(cfg.Database)
	dbConnected := err == nil
	if err != nil {
		// Print banner first even on error
		startup.PrintBanner(cfg, false, cfg.Validation.Errors, cfg.Validation.Warnings)
		logger.Error("Failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Print clean startup banner
	startup.PrintBanner(cfg, dbConnected, cfg.Validation.Errors, cfg.Validation.Warnings)

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
	salesInvoiceRepo := postgres.NewSalesInvoiceRepository(db)
	quotationRepo := postgres.NewSalesQuotationRepository(db)
	orderRepo := postgres.NewSalesOrderRepository(db)
	deliveryRepo := postgres.NewDeliveryOrderRepository(db)

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
	productUsecase := productUC.NewProductUsecase(productRepo, inventoryRepo)
	currencyUsecase := currencyUC.NewCurrencyUsecase(currencyRepo)
	exchangerateUsecase := exchangerateUC.NewExchangeRateUsecase(exchangeRateRepo, currencyRepo)
	warehouseUsecase := warehouseUC.NewWarehouseUsecase(warehouseRepo)
	inventoryUsecase := inventoryUC.NewInventoryUsecase(inventoryRepo, warehouseRepo, productRepo, txManager)
	salesUsecase := salesUC.NewSalesUsecase(salesInvoiceRepo, inventoryUsecase, warehouseRepo, txManager)
	quotationUsecase := salesUC.NewQuotationUsecase(quotationRepo, orderRepo, productRepo)
	orderUsecase := salesUC.NewOrderUsecase(orderRepo, deliveryRepo, txManager)
	deliveryUsecase := salesUC.NewDeliveryUsecase(deliveryRepo, orderRepo, inventoryUsecase, txManager)

	// Advanced Inventory
	opnameRepo := postgres.NewStockOpnameRepository(db)
	transferRepo := postgres.NewStockTransferRepository(db)
	opnameUsecase := inventoryUC.NewOpnameUsecase(opnameRepo, inventoryUsecase, inventoryRepo, txManager)
	transferUsecase := inventoryUC.NewTransferUsecase(transferRepo, inventoryUsecase, inventoryRepo, txManager)

	// Fixed Assets
	fixedAssetRepo := postgres.NewFixedAssetRepository(db)
	assetCategoryRepo := postgres.NewAssetCategoryRepository(db)
	fixedAssetUsecase := fixedassetUC.NewFixedAssetUsecase(fixedAssetRepo, assetCategoryRepo, journalRepo, periodRepo, txManager)

	// Tax
	taxRateRepo := postgres.NewTaxRateRepository(db)
	taxReturnRepo := postgres.NewTaxReturnRepository(db)
	taxUsecase := taxUC.NewTaxUsecase(taxRateRepo, taxReturnRepo, journalRepo, txManager)

	// Banking
	bankAccountRepo := postgres.NewBankAccountRepository(db)
	bankTxnRepo := postgres.NewBankTransactionRepository(db)
	bankingUsecase := bankingUC.NewBankingUsecase(bankAccountRepo, bankTxnRepo, journalRepo, txManager)

	// Purchasing
	purchaseOrderRepo := postgres.NewPurchaseOrderRepository(db)
	purchaseInvoiceRepo := postgres.NewPurchaseInvoiceRepository(db)
	purchasingUsecase := purchasingUC.NewPurchasingUsecase(purchaseOrderRepo, purchaseInvoiceRepo, journalRepo, txManager)

	// AR/AP Reporting
	arapRepo := postgres.NewARAPRepository(db)
	arapUsecase := arapUC.NewARAPUsecase(arapRepo)

	// Delivery Layer - HTTP Handlers
	handlers := &httpDelivery.Handlers{
		Health:       handler.NewHealthHandler(),
		Auth:         handler.NewAuthHandler(authUsecase, jwtService),
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
		Upload:       handler.NewUploadHandler("./uploads", "http://localhost:"+cfg.Server.Port),
		Quotation:    handler.NewQuotationHandler(quotationUsecase),
		Order:        handler.NewOrderHandler(orderUsecase),
		Delivery:     handler.NewDeliveryHandler(deliveryUsecase),
		Opname:       handler.NewOpnameHandler(opnameUsecase),
		Transfer:     handler.NewTransferHandler(transferUsecase),
		FixedAsset:   handler.NewFixedAssetHandler(fixedAssetUsecase),
		Tax:          handler.NewTaxHandler(taxUsecase),
		Banking:      handler.NewBankingHandler(bankingUsecase),
		Purchasing:   handler.NewPurchasingHandler(purchasingUsecase),
		ARAP:         handler.NewARAPHandler(arapUsecase),
	}

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.GlobalPerMin, time.Duration(cfg.RateLimit.GlobalWindowSec)*time.Second)

	// Start periodic cleanup goroutine to prevent memory leak
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rateLimiter.Cleanup()
		}
	}()

	// ============================================
	// ROUTER SETUP
	// ============================================

	// Gin already set to ReleaseMode at startup
	r := gin.New()

	// Global middleware
	r.Use(middleware.RequestLogger(logger)) // Structured logging
	r.Use(middleware.RecoveryHandler(cfg.Server.Env))
	r.Use(middleware.SecurityHeaders())                 // Security headers
	r.Use(middleware.StrictTransportSecurity(31536000)) // HSTS
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.Security.AllowedOrigins,
		Environment:    cfg.Server.Env,
	}
	r.Use(middleware.NewCORSMiddleware(corsConfig))
	if cfg.RateLimit.Enabled {
		r.Use(rateLimiter.RateLimitMiddleware())
	}

	// 404 handler
	r.NoRoute(middleware.NotFoundHandler())

	// Swagger (with feature flag)
	if cfg.Features.EnableSwagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Setup routes with injected dependencies
	httpDelivery.SetupRouter(r, handlers, authMiddleware)

	// ============================================
	// GRACEFUL SHUTDOWN
	// ============================================

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	startup.PrintShutdownBanner()

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}

	// Close database connection
	db.Close()

	startup.PrintShutdownComplete()
}
