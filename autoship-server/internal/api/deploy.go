package api

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/cloud"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/services"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// maxConcurrentDeploys caps how many builds run at the same time. Docker
// build + temp container probing is heavy; without a cap, a burst of requests
// can OOM the host.
const maxConcurrentDeploys = 8

// deploySlots is a counting semaphore — acquire by sending, release by
// receiving. Excess deploys block on send until a slot frees up.
var deploySlots = make(chan struct{}, maxConcurrentDeploys)

type deployResult struct {
	url           string
	projectType   string
	containerPort int
	hostPort      int
	containerName string
}

// runDeployment executes the build/deploy pipeline for a project that has
// already been persisted in the pending state. Runs in a detached goroutine
// from HandleRepoSubmit so HTTP-request cancellation does NOT abort the
// build mid-way (which would otherwise leave half-running containers).
//
// Status transitions: pending -> deploying -> (succeeded | failed). Failure
// detail lands in Project.DeployError so the client polling /projects/:id
// can surface it without needing log access.
func runDeployment(id primitive.ObjectID, username, repoURL, repoName, envContent, startCommand string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[deploy %s] panic: %v", id.Hex(), r)
			_ = db.UpdateProjectByID(id, bson.M{
				"status":       models.StatusFailed,
				"deploy_error": fmt.Sprintf("panic: %v", r),
			})
		}
	}()

	deploySlots <- struct{}{}
	defer func() { <-deploySlots }()

	if err := db.UpdateProjectByID(id, bson.M{"status": models.StatusDeploying}); err != nil {
		log.Printf("[deploy %s] failed to mark deploying: %v", id.Hex(), err)
		// keep going — a transient Mongo blip shouldn't abort the build
	}

	result, err := executeDeploy(username, repoURL, repoName, envContent, startCommand)
	if err != nil {
		log.Printf("[deploy %s] failed: %v", id.Hex(), err)
		_ = db.UpdateProjectByID(id, bson.M{
			"status":       models.StatusFailed,
			"deploy_error": err.Error(),
		})
		return
	}

	if err := db.UpdateProjectByID(id, bson.M{
		"status":         models.StatusSucceeded,
		"project_type":   result.projectType,
		"hosted_url":     result.url,
		"container_port": result.containerPort,
		"host_port":      result.hostPort,
		"container_name": result.containerName,
	}); err != nil {
		log.Printf("[deploy %s] succeeded but failed to persist result: %v", id.Hex(), err)
		return
	}
	log.Printf("[deploy %s] succeeded (%s)", id.Hex(), result.url)
}

func executeDeploy(username, repoURL, repoName, envContent, startCommand string) (deployResult, error) {
	path, err := services.CloneRepository(repoURL, username, repoName)
	if err != nil {
		return deployResult{}, fmt.Errorf("clone failed: %w", err)
	}
	defer os.RemoveAll(path)

	projectType := services.DetectProjectType(path)
	if projectType == "unknown" {
		return deployResult{}, fmt.Errorf("unknown project type; no recognised entry file in repo")
	}

	if projectType == "static" {
		keyPrefix := fmt.Sprintf("%s/%s", username, repoName)
		url, err := cloud.Get().UploadStaticSite(path, keyPrefix)
		if err != nil {
			return deployResult{}, fmt.Errorf("static upload: %w", err)
		}
		return deployResult{url: url, projectType: projectType}, nil
	}

	// Dynamic: build + run container locally, then hand off to autoship-scripts
	// (via the file-based queue) for DNS/nginx/SSL.
	containerPort, hostPort, containerName, err := services.FullPipeline(username, path, envContent, startCommand)
	if err != nil {
		return deployResult{}, fmt.Errorf("container pipeline: %w", err)
	}

	subdomain := utils.GenerateSubdomain(repoName, os.Getenv("DOMAIN"))
	reqID := utils.GenerateRandomID()
	deployReq := map[string]interface{}{
		"id":          reqID,
		"subdomain":   subdomain,
		"projectType": projectType,
		"port":        hostPort,
		"status":      "pending",
	}
	if err := utils.AppendJSONToFile("/var/lib/autoship/deploy/deploy-requests.json", deployReq); err != nil {
		return deployResult{}, fmt.Errorf("queue deploy request: %w", err)
	}

	resp, err := utils.WaitForResponse("/var/lib/autoship/deploy/deploy-responses.json", reqID, 60*time.Second)
	if err != nil {
		return deployResult{}, fmt.Errorf("wait for autoship-scripts: %w", err)
	}
	if status, _ := resp["status"].(string); status != "success" {
		return deployResult{}, fmt.Errorf("autoship-scripts reported failure: %v", resp["error"])
	}

	url, _ := resp["url"].(string)
	if url == "" {
		url = fmt.Sprintf("https://%s", subdomain)
	}
	return deployResult{
		url:           url,
		projectType:   projectType,
		containerPort: containerPort,
		hostPort:      hostPort,
		containerName: containerName,
	}, nil
}
