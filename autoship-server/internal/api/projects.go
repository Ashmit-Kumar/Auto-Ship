// internal/api/projects.go
package api

import (
	"os"
	"fmt"
	"strings"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/services"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/utils"
	"github.com/gofiber/fiber/v2"
	"log"
	"go.mongodb.org/mongo-driver/bson"
)

// RepoRequest struct defines the structure of the request for submitting a repo
type RepoRequest struct {
	RepoURL string `json:"repoURL"`
	EnvContent string `json:"envContent,omitempty"` // Optional field for .env content
	// IN future, add fields for branch, commit 

}

// HandleRepoSubmit handles the submission of a GitHub repository URL,
// clones the repository, detects the project type, and handles the hosting.
func HandleRepoSubmit(c *fiber.Ctx) error {
	var req RepoRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}
	if req.RepoURL == "" {
		log.Println("repoURL not provided in request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "repoURL is required"})
	}
	
	// fmt.Println("Received repo URL:", req.RepoURL)


	// Extract username from the GitHub repo URL
	username, err := utils.ExtractUsernameFromRepoURL(req.RepoURL)
	if err != nil {
		log.Printf("Error extracting username: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Extract repo name from the URL
	parts := strings.Split(strings.TrimSuffix(req.RepoURL, ".git"), "/")
	repoName := parts[len(parts)-1]

	// Clone the repository
	path, err := services.CloneRepository(req.RepoURL, username, repoName)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Detect the type of the project (static or dynamic)
	projectType := services.DetectProjectType(path)
	var hostedURL string

	// If the project is static, upload to S3 and generate a hosted URL
	if projectType == "static" {
		keyPrefix := fmt.Sprintf("%s/%s", username, repoName)
		url, err := services.UploadStaticSite(path, keyPrefix)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to upload to S3: "+err.Error())
		}
		_ = os.RemoveAll(path)

		//when returning the s3 url shorten it and encrypt it and in future all links we be routed through a proxy server
		hostedURL = url
	} else {
        // Run FullPipeline to detect environment, write Dockerfile, build & run
        port, err := services.FullPipeline(path, req.EnvContent) // returns hostPort
        if err != nil {
            _ = os.RemoveAll(path)
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to deploy dynamic project: "+err.Error())
        }

        ec2Host := os.Getenv("EC2_PUBLIC_IP") // Set this in your .env or config
        if ec2Host == "" {
            ec2Host = "localhost"
        }
        hostedURL = fmt.Sprintf("http://%s:%d", ec2Host, port)
    }
		
	// encrypt the hosted URL for security
	// Create a new project model
	project := &models.Project{
		Username:    username,
		RepoURL:     req.RepoURL,
		RepoName:    repoName,
		ProjectType: projectType,
		HostedURL:   hostedURL,
		// add columns for createdAt, updatedAt, etc.
		// ports will be added in future
		// start command: "",
	}

	// Save the project details to the database
	if err := db.SaveProject(project); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save project")
	}

	// Return a success response with project details
	return c.JSON(fiber.Map{
		"message":     "Repository cloned and hosted successfully",
		"projectType": projectType,
		"url":         hostedURL,
	})
}



// GetUserProjects fetches all projects belonging to the authenticated user
func GetUserProjects(c *fiber.Ctx) error {
	claims := c.Locals("user").(map[string]interface{})
	username := claims["username"].(string)

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
