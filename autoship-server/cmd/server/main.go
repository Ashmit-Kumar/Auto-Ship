// main.go
package main

import (
	"log"
	"os"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/api"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/middleware"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/utils" // Import utils package for JWT
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/services"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	if err := services.InitS3(); err != nil {
		log.Fatalf("Failed to initialize S3: %v", err)
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
	app.Use(cors.New())

	// health check route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	
	// Routes
	app.Post("/signup", api.Signup)
	app.Post("/login", api.Login)
	app.Get("/auth/github", api.GitHubLogin)
	// app.Get("/auth/github/callback", api.GitHubCallback)
	app.Get("/github/callback", api.GitHubCallback)
	app.Post("/projects/submit", api.HandleRepoSubmit)
	// Serve everything under static as public files
	app.Static("/static", "./static", fiber.Static{
		Browse:     true,
		Index:      "index.html",
		Compress:   true,
	})
	// to serve static files from the static directory like scripts, styles, etc.
	app.Get("/autoship-server/static/:username/:repo/*", func(c *fiber.Ctx) error {
		newPath := "/static/" + c.Params("username") + "/" + c.Params("repo") + "/" + c.Params("*")
		return c.Redirect(newPath, fiber.StatusTemporaryRedirect)
	})

	// Protected routes
	app.Get("/protected", middleware.IsAuthenticated, func(c *fiber.Ctx) error {
		user := c.Locals("user")
		return c.JSON(fiber.Map{"message": "Protected content", "user": user})
	})

	// Start the server
	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
