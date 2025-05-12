package services

import (
	"fmt"
	"io/ioutil"
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
	_, err := ioutil.ReadDir(repoPath)
	if err != nil {
		fmt.Println("read dir error:", err)
		return EnvUnknown
	}

	has := func(name string) bool {
		_, err := os.Stat(filepath.Join(repoPath, name))
		return err == nil
	}

	switch {
	case has("package.json"):
		return EnvNode
	case has("requirements.txt") || has("app.py"):
		return EnvPython
	case has("main.go"):
		return EnvGo
	default:
		return EnvUnknown
	}
}

// writeDockerfile generates a Dockerfile based on the detected environment.
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
func buildAndRunContainer(repoPath, containerName string, port int) error {
	imageTag := containerName + ":latest"

	// Build image
	buildCmd := exec.Command("docker", "build", "-t", imageTag, ".")
	buildCmd.Dir = repoPath
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// Run container
	runCmd := exec.Command(
		"docker", "run", "-d",
		"-p", fmt.Sprintf("%d:%d", port, port),
		"--name", containerName,
		imageTag,
	)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	return runCmd.Run()
}

// FullPipeline executes the full flow: detects env, generates Dockerfile, builds, and runs container.
func FullPipeline(repoPath, envContent string) error {
	// Step 1: Save .env if provided
	if envContent != "" {
		if err := utils.SaveEnvFile(repoPath, envContent); err != nil {
			return fmt.Errorf("failed to save .env: %w", err)
		}
	}

	// Step 2: Detect environment
	envType := detectEnvironment(repoPath)
	if envType == EnvUnknown {
		return fmt.Errorf("unsupported environment")
	}

	// Step 3: Generate Dockerfile
	if err := writeDockerfile(envType, repoPath); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Step 4: Get a free port
	port, err := utils.GetFreePort()
	if err != nil {
		return fmt.Errorf("could not get free port: %w", err)
	}

	// Step 5: Derive container name from repo path
	repoName := filepath.Base(repoPath)
	containerName := fmt.Sprintf("autoship-%s", strings.ToLower(repoName))

	// Step 6: Build and run container
	if err := buildAndRunContainer(repoPath, containerName, port); err != nil {
		return fmt.Errorf("container error: %w", err)
	}

	return nil
}
