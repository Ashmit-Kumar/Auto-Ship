package api

import (
	"strings"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/services"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
	"github.com/gofiber/fiber/v2"
)

type RepoRequest struct {
	URL      string `json:"url"`
	Username string `json:"username"`
}

func HandleRepoSubmit(c *fiber.Ctx) error {
	var req RepoRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}

	// Extract repo name from URL
	parts := strings.Split(strings.TrimSuffix(req.URL, ".git"), "/")
	repoName := parts[len(parts)-1]

	path, err := services.CloneRepository(req.URL, req.Username, repoName)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	projectType := services.DetectProjectType(path)
	url := "/static/" + req.Username + "/" + repoName

	project := &models.Project{
		Username:    req.Username,
		RepoURL:     req.URL,
		RepoName:    repoName,
		ProjectType: projectType,
		HostedURL:   url,
	}

	if err := db.SaveProject(project); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save project")
	}

	return c.JSON(fiber.Map{
		"message":     "Repository cloned and hosted",
		"projectType": projectType,
		"url":         url,
	})
}
