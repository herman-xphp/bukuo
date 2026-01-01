package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/config"
	"github.com/herman-xphp/bukuo/internal/database"
	"github.com/herman-xphp/bukuo/internal/routes"
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

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Setup routes
	routes.Setup(r)

	// Start server
	log.Printf("🚀 Bukuo running on port %s", cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}
