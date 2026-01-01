package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/config"
	"github.com/herman-xphp/bukuo/internal/database"
	httpDelivery "github.com/herman-xphp/bukuo/internal/delivery/http"
	"github.com/herman-xphp/bukuo/internal/delivery/http/handler"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres"
	journalUC "github.com/herman-xphp/bukuo/internal/usecase/journal"
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
	accountRepo := postgres.NewAccountRepository(db)
	periodRepo := postgres.NewPeriodRepository(db)
	journalRepo := postgres.NewJournalRepository(db)

	// Usecase Layer - Business Logic
	journalUsecase := journalUC.NewJournalUsecase(journalRepo, accountRepo, periodRepo)

	// Delivery Layer - HTTP Handlers
	healthHandler := handler.NewHealthHandler()
	journalHandler := handler.NewJournalHandler(journalUsecase)

	// ============================================
	// ROUTER SETUP
	// ============================================

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Setup routes with injected dependencies
	httpDelivery.SetupRouter(r, healthHandler, journalHandler)

	// Start server
	log.Printf("🚀 Bukuo running on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
