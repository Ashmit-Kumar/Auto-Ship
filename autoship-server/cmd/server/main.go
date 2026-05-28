package main

import (
	"log"
	"os"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/api"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/cloud"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	if err := cloud.Init(os.Getenv("CLOUD_PROVIDER")); err != nil {
		log.Fatalf("Failed to initialize cloud provider: %v", err)
	}
	if err := utils.LoadEnv(); err != nil {
		log.Fatalf("Error loading JWT environment variables: %v", err)
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI is not set")
	}
	db.SetMongoURI(mongoURI)
	db.Connect()
	defer db.Disconnect()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app := fiber.New()
	app.Use(cors.New())
	api.RegisterRoutes(app)

	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
