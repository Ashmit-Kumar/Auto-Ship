package services

import (
	"fmt"
	// "io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/utils"
)

// Environment represents the type of backend environment detected.
type Environment string

const (
	EnvNode    Environment = "node"
	EnvPython  Environment = "python"
	EnvGo      Environment = "go"
	EnvUnknown Environment = "unknown"
)

// detectEnvironment inspects the repo to determine the runtime environment.
func detectEnvironment(repoPath string) Environment {
	var foundNode, foundPython, foundGo bool

	err := filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // ignore errors
		}

		// Normalize file name (in case-insensitive FS)
		switch strings.ToLower(d.Name()) {
		case "package.json":
			foundNode = true
		case "requirements.txt", "app.py":
			foundPython = true
		case "main.go":
			foundGo = true
		}

		// Stop early if we've found something
		if foundNode || foundPython || foundGo {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		fmt.Println("error walking repo:", err)
		return EnvUnknown
	}

	switch {
	case foundNode:
		return EnvNode
	case foundPython:
		return EnvPython
	case foundGo:
		return EnvGo
	default:
		return EnvUnknown
	}
}


// writeDockerfile generates a Dockerfile based on the detected environment.
// GenerateDockerfile creates a Dockerfile dynamically using detected environment and user-provided startCommand.
func GenerateDockerfile(env Environment, repoPath, startCommand string) error {
	var baseImage string
	var installCmd string

	switch env {
	case EnvNode:
		baseImage = "node:18"
		installCmd = "RUN npm install"
	case EnvPython:
		baseImage = "python:3.10"
		installCmd = "RUN pip install -r requirements.txt"
	case EnvGo:
		baseImage = "golang:1.20"
		installCmd = "RUN go build -o app ."
	default:
		return fmt.Errorf("unsupported environment: %s", env)
	}

	// Sanitize and convert the startCommand to CMD array format for Dockerfile
	cmdParts := strings.Fields(startCommand)
	if len(cmdParts) == 0 {
		return fmt.Errorf("start command is empty")
	}

	var quotedParts []string
	for _, part := range cmdParts {
		quotedParts = append(quotedParts, fmt.Sprintf("\"%s\"", part))
	}
	cmdLine := fmt.Sprintf("CMD [%s]", strings.Join(quotedParts, ", "))

	// Assemble the Dockerfile content
	dockerfile := fmt.Sprintf(`FROM %s
	WORKDIR /app
	COPY . .
	%s
	%s
	`, baseImage, installCmd, cmdLine)

	// Write to Dockerfile
	dockerfilePath := filepath.Join(repoPath, "Dockerfile")
	return os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
}


// buildAndRunContainer builds the Docker image and runs it on a specified port.
func buildAndRunContainerHybrid(repoPath, containerName string) (int, error) {
	imageTag := containerName + ":latest"

	// Step 1: Build Docker image
	buildCmd := exec.Command("docker", "build", "-t", imageTag, ".")
	buildCmd.Dir = repoPath
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return 0, fmt.Errorf("docker build failed: %w", err)
	}

	// Step 2: Run container temporarily (no port binding)
	tmpContainerName := containerName + "-tmp"
	runCmd := exec.Command(
		"docker", "run", "-d", "--name", tmpContainerName, imageTag,
	)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		return 0, fmt.Errorf("docker run (tmp) failed: %w", err)
	}

	// Step 3: Detect the exposed port inside container
	port, err := utils.DetectExposedPort(tmpContainerName)
	if err != nil {
		// Clean up the temporary container
		_ = exec.Command("docker", "rm", "-f", tmpContainerName).Run()
		return 0, fmt.Errorf("failed to detect port: %w", err)
	}

	// Step 4: Reserve a host port (DB + EC2)
	hostPort := port
	if !utils.IsPortAvailable(port) {
		var err error
		hostPort, err = utils.FindFreeHostPort()
		if err != nil {
			_ = exec.Command("docker", "rm", "-f", tmpContainerName).Run()
			return 0, fmt.Errorf("failed to find free host port: %w", err)
		}
	}
	if err := utils.AuthorizeEC2Port(hostPort); err != nil {
		_ = exec.Command("docker", "rm", "-f", tmpContainerName).Run()
		return 0, fmt.Errorf("EC2 SG error: %w", err)
	}

	// Step 5: Commit the container state as image (optional if no changes made)
	_ = exec.Command("docker", "commit", tmpContainerName, imageTag).Run()

	// Step 6: Remove temp container
	_ = exec.Command("docker", "rm", "-f", tmpContainerName).Run()

	// Step 7: Run final container with proper port binding
	finalRunCmd := exec.Command(
		"docker", "run", "-d",
		"-p", fmt.Sprintf("%d:%d", hostPort, port),
		"--name", containerName,
		imageTag,
	)
	finalRunCmd.Stdout = os.Stdout
	finalRunCmd.Stderr = os.Stderr
	if err := finalRunCmd.Run(); err != nil {
		return 0, fmt.Errorf("docker final run failed: %w", err)
	}

	return port, nil
}


// FullPipeline executes the full flow: detects env, generates Dockerfile, builds, and runs container.
func FullPipeline(repoPath, envContent, startCommand string) (int, error) {
	// Step 1: Save .env if provided
	if envContent != "" {
		if err := utils.SaveEnvFile(repoPath, envContent); err != nil {
			return 0, fmt.Errorf("failed to save .env: %w", err)
		}
	}

	// Step 2: Detect environment
	envType := detectEnvironment(repoPath)
	if envType == EnvUnknown {
		return 0, fmt.Errorf("unsupported environment")
	}

	// Step 3: Generate Dockerfile dynamically using startCommand
	if err := GenerateDockerfile(envType, repoPath, startCommand); err != nil {
		return 0, fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Step 4: Derive container name from repo path
	repoName := filepath.Base(repoPath)
	containerName := fmt.Sprintf("autoship-%s", strings.ToLower(repoName))

	// Step 5: Build and run container
	port, err := buildAndRunContainerHybrid(repoPath, containerName)
	if err != nil {
		return 0, fmt.Errorf("container error: %w", err)
	}

	log.Printf("Container %s started on port %d", containerName, port)
	return port, nil
}

