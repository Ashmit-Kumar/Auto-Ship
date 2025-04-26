package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/api"
)

func main() {
	app := fiber.New()

	// Routes
	app.Get("/ping", api.Ping)

	log.Println("🚀 Server running at http://localhost:5000")
	log.Fatal(app.Listen(":5000"))
}
