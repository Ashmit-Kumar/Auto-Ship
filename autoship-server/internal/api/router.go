package api

import (
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes wires every HTTP route exposed by the server.
//
// FUTURE: API versioning. When the first breaking change to an endpoint is
// actually needed, introduce per-version handler packages (internal/api/v1/,
// internal/api/v2/) each exposing their own RegisterRoutes(r fiber.Router),
// and mount them here: app.Group("/api/v1") -> v1.RegisterRoutes(...), etc.
// Prerequisite: thin out handlers first so orchestration lives in services
// both versions can share — otherwise v2 will just be a 100-line copy of v1.
// URL-prefix-only versioning (every route prefixed /api/v1 while handlers
// stay shared) is not versioning, it's renaming — don't bother until the
// handler split is real.
func RegisterRoutes(app *fiber.App) {
	app.Get("/health", healthCheck)

	app.Static("/static", "./static", fiber.Static{
		Browse:   true,
		Index:    "index.html",
		Compress: true,
	})
	app.Get("/autoship-server/static/:username/:repo/*", redirectLegacyStatic)

	registerAuthRoutes(app)
	registerProjectRoutes(app)
}

func registerAuthRoutes(app *fiber.App) {
	app.Post("/signup", Signup)
	app.Post("/login", Login)
	app.Get("/auth/github", GitHubLogin)
	app.Get("/github/callback", GitHubCallback)
}

func registerProjectRoutes(app *fiber.App) {
	app.Post("/projects/submit", middleware.IsAuthenticated, HandleRepoSubmit)
	app.Get("/projects", middleware.IsAuthenticated, GetUserProjects)
	app.Delete("/projects/:containerName", middleware.IsAuthenticated, DeleteDeployment)
}

func healthCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func redirectLegacyStatic(c *fiber.Ctx) error {
	newPath := "/static/" + c.Params("username") + "/" + c.Params("repo") + "/" + c.Params("*")
	return c.Redirect(newPath, fiber.StatusTemporaryRedirect)
}
