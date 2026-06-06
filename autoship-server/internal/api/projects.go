// internal/api/projects.go
package api

import (
	"log"
	"strings"
	"time"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/config"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/services"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RepoRequest is the body shape of POST /projects/submit.
type RepoRequest struct {
	RepoURL      string `json:"repoURL"`
	EnvContent   string `json:"envContent,omitempty"`
	StartCommand string `json:"startCommand"`
}

// HandleRepoSubmit accepts a deploy request, persists a pending Project, and
// kicks off the build in a detached goroutine. Returns 202 with the deployment
// id so the client can poll GET /projects/:id for status. Running the build
// inline would exceed typical browser/proxy timeouts on real repos (docker
// build + Azure NSG poll can run several minutes).
func HandleRepoSubmit(c *fiber.Ctx) error {
	var req RepoRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}
	if req.RepoURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "repoURL is required"})
	}

	username, err := utils.ExtractUsernameFromRepoURL(req.RepoURL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	parts := strings.Split(strings.TrimSuffix(req.RepoURL, ".git"), "/")
	repoName := parts[len(parts)-1]

	now := time.Now()
	project := &models.Project{
		ID:           primitive.NewObjectID(),
		Username:     username,
		RepoURL:      req.RepoURL,
		RepoName:     repoName,
		Status:       models.StatusPending,
		StartCommand: req.StartCommand,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.SaveProject(project); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record deployment")
	}

	go runDeployment(project.ID, username, req.RepoURL, repoName, req.EnvContent, req.StartCommand)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"deploymentID": project.ID.Hex(),
		"status":       project.Status,
	})
}

// GetProjectStatus returns the current state of a deployment by its id. The
// client polls this while Status is pending or deploying. Scoped to the
// authenticated user — ObjectIDs aren't unguessable enough to skip the check.
func GetProjectStatus(c *fiber.Ctx) error {
	objID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid deployment id")
	}

	project, err := db.GetProjectByID(objID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "deployment not found")
	}

	claims := c.Locals("user").(*utils.Claims)
	if project.Username != claims.Email {
		return fiber.NewError(fiber.StatusNotFound, "deployment not found")
	}

	return c.JSON(project)
}

// GetUserProjects fetches all projects belonging to the authenticated user.
func GetUserProjects(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	username := claims.Email

	collection := db.GetCollection("projects")
	cursor, err := collection.Find(c.Context(), bson.M{"username": username})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch projects"})
	}
	defer cursor.Close(c.Context())

	var projects []bson.M
	if err := cursor.All(c.Context(), &projects); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode projects"})
	}

	return c.JSON(projects)
}

// DeleteDeployment deletes a project deployment by container name.
func DeleteDeployment(c *fiber.Ctx) error {
	containerName := c.Params("containerName")
	if containerName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "containerName is required")
	}

	if err := services.DeleteProject(containerName); err != nil {
		log.Printf("Failed to delete project deployment for container %s: %v", containerName, err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete deployment")
	}

	if err := db.DeleteProjectByContainerName(containerName); err != nil {
		log.Printf("Failed to delete project from DB for container %s: %v", containerName, err)
	}

	// Free the host port for reuse. Non-fatal: the container is gone and the
	// OS port is unbound regardless; the worst case is a stale "used" doc.
	if err := db.ReleasePort(config.Get().MongoCollection, containerName); err != nil {
		log.Printf("Failed to release port for container %s: %v", containerName, err)
	}

	return c.JSON(fiber.Map{"message": "Deployment deleted successfully"})
}
