package main

import (
	"log"
	"os"

	"github.com/ashmit123/Auto-Ship/autoship-server/internal/api"
	"github.com/ashmit123/Auto-Ship/autoship-server/internal/db"
	"github.com/ashmit123/Auto-Ship/autoship-server/internal/middleware"
	"github.com/ashmit123/Auto-Ship/autoship-server/internal/utils" // Import utils package for JWT
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load JWT_SECRET and JWT_EXPIRATION values from environment variables
	if err := utils.LoadEnv(); err != nil {
		log.Fatalf("Error loading JWT environment variables: %v", err)
	}

	// Get PORT and MONGO_URI from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default to 3000 if not set in .env
	}

	// Get MongoDB URI from environment variables
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI is not set in .env file")
	}

	// Set MongoDB URI
	db.SetMongoURI(mongoURI)

	// Connect to MongoDB
	db.Connect()
	defer db.Disconnect()

	// Initialize Fiber app
	app := fiber.New()

	// Routes
	app.Post("/signup", api.Signup)
	app.Post("/login", api.Login)

	// Protected routes
	app.Get("/protected", middleware.IsAuthenticated, func(c *fiber.Ctx) error {
		user := c.Locals("user")
		return c.JSON(fiber.Map{"message": "Protected content", "user": user})
	})

	// Start the server
	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
