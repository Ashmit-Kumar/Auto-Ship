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
// I will work on This Today
func writeDockerfile(env Environment, repoPath string) error {
	templateDir := filepath.Join("internal", "services", "docker_templates")

	var templateFile string
	switch env {
	case EnvNode:
		templateFile = filepath.Join(templateDir, "node.Dockerfile")
	case EnvPython:
		templateFile = filepath.Join(templateDir, "python.Dockerfile")
	case EnvGo:
		templateFile = filepath.Join(templateDir, "go.Dockerfile")
	default:
		return fmt.Errorf("unknown environment")
	}

	content, err := os.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile template: %w", err)
	}

	return os.WriteFile(filepath.Join(repoPath, "Dockerfile"), content, 0644)
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
func FullPipeline(repoPath, envContent string) (int, error) {
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

	// Step 3: Generate Dockerfile
	if err := writeDockerfile(envType, repoPath); err != nil {
		return 0, fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Step 4: Derive container name from repo path
	repoName := filepath.Base(repoPath)
	containerName := fmt.Sprintf("autoship-%s", strings.ToLower(repoName))


	port, err := buildAndRunContainerHybrid(repoPath, containerName)
	if err != nil {
		return 0, fmt.Errorf("container error: %w", err)
	}

	// You can log or store the port if needed
	log.Printf("Container %s started on port %d", containerName, port)

	return port, nil
}
