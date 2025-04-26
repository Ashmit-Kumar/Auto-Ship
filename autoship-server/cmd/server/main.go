package main

import (
	"log"
	"os"

	"github.com/ashmit123/Auto-Ship/autoship-server/internal/api"
	"github.com/ashmit123/Auto-Ship/autoship-server/internal/db"
	"github.com/ashmit123/Auto-Ship/autoship-server/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to MongoDB
	db.Connect()
	defer db.Disconnect()

	app := fiber.New()

	// Routes
	app.Post("/signup", api.Signup)
	app.Post("/login", api.Login)

	// Protected routes
	app.Get("/protected", middleware.IsAuthenticated, func(c *fiber.Ctx) error {
		user := c.Locals("user")
		return c.JSON(fiber.Map{"message": "Protected content", "user": user})
	})

	log.Println("🚀 Auto-Ship server running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
