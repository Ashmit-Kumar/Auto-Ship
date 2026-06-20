package main

import (
	"log"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/api"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/cloud"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/config"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := cloud.Init(cfg.CloudProvider); err != nil {
		log.Fatalf("Failed to initialize cloud provider: %v", err)
	}
	if err := utils.LoadEnv(); err != nil {
		log.Fatalf("Error loading JWT environment variables: %v", err)
	}

	db.SetMongoURI(cfg.MongoURI)
	db.Connect()
	defer db.Disconnect()
	if err := db.EnsurePortsIndex(cfg.MongoCollection); err != nil {
		log.Fatalf("Failed to ensure ports unique index: %v", err)
	}

	app := fiber.New()
	app.Use(cors.New())
	api.RegisterRoutes(app)

	log.Printf("🚀 Server running on http://localhost:%s", cfg.ServerPort)
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}
