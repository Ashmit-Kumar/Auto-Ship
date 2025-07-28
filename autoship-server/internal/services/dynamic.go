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

		fmt.Printf("Visited: %s\n", path)
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
		fmt.Println("Detected Node.js environment")
		return EnvNode
	case foundPython:
		fmt.Println("Detected Python environment")
		return EnvPython
	case foundGo:
		fmt.Println("Detected Go environment")
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
	var netToolsInstall string
	fmt.Println("repopath:", repoPath)
	switch env {
	case EnvNode:
		baseImage = "node:18"
		netToolsInstall = "RUN apt update && apt install -y net-tools"
		installCmd = "RUN npm install"
	case EnvPython:
		baseImage = "python:3.10"
		netToolsInstall = "RUN apt update && apt install -y net-tools"
		installCmd = "RUN pip install -r requirements.txt"
	case EnvGo:
		baseImage = "golang:1.20"
		netToolsInstall = "RUN apt update && apt install -y net-tools"
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
	%s
	`, baseImage, netToolsInstall, installCmd, cmdLine)

	// Write to Dockerfile
	dockerfilePath := filepath.Join(repoPath, "Dockerfile")
	return os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
}


// buildAndRunContainer builds the Docker image and runs it on a specified port.
func buildAndRunContainerHybrid(repoPath, containerName string) (int, int, error) {
	// Derive image tag from container name
	if containerName == "" {
		return 0, 0, fmt.Errorf("container name cannot be empty")
	}
	if repoPath == "" {
		return 0, 0, fmt.Errorf("repository path cannot be empty")
	}
	containerName = strings.TrimSpace(containerName)
	containerName = strings.ToLower(containerName) // Ensure consistent casing
	// Derive image tag from container name
	imageTag := containerName + ":latest"

	// Step 1: Build image
	buildCmd := exec.Command("docker", "build", "-t", imageTag, ".")
	buildCmd.Dir = repoPath
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("docker build failed: %w", err)
	}

	// Step 2: Run temp container
	fmt.Println("Building a temporary Container 88 88888888 888888888888 88888888888 888888888888   888888888888 888888888888 8888888     888888")
	tmpContainer := containerName + "-tmp"
	fmt.Println("Running temporary container for port detection... ", tmpContainer)
	// Create a temporary container to detect exposed port
	// Use the same image tag but with a different name
	// This avoids conflicts with the final container
	
	
	// if err := exec.Command("docker", "rm", "-f", tmpContainer).Run(); err != nil {
	// 	// Ignore error if container doesn't exist
	// 	log.Printf("Warning: Failed to remove existing temporary container %s: %v", tmpContainer, err)
	// }
	
	
	
	// Run the temporary container in detached mode
	// Use the image tag built earlier
	fmt.Println("Running temporary container for port detection... ", tmpContainer)
	// Use the image tag built earlier
	// Use a simple command that keeps the container running
	// This is just to keep the container alive for port detection
	// You can use any command that suits your needs, like "tail -f /dev/null"
	// or a simple sleep command
	// Here we use "bash" to keep it running
	// This allows us to inspect the container later
	// Note: This is a temporary container, it will be removed after port detection
	// Use the image tag built earlier
	runCmd := exec.Command("docker", "run", "-d", "--name", tmpContainer, imageTag)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("docker run (tmp) failed: %w", err)
	}

	// Step 3: Detect container's exposed port
	fmt.Println("Detecting Exposed Port PPPPPPPPPPPPPPPPPPPPPPPPPPOOOOOOOOOOOOOOOOOOOOOOOOORRRRRRRRRRRRRRRRRRRRRRTTTTTTTTTTTTTTTTTTTTTTTTTTTT")
	containerPort, err := utils.DetectExposedPort(tmpContainer)
	if err != nil {
		logsCmd := exec.Command("docker", "logs", tmpContainer)
		logsOutput, _ := logsCmd.CombinedOutput()
		fmt.Println("Temporary container logs:\n", string(logsOutput))

		// _ = exec.Command("docker", "rm", "-f", tmpContainer).Run()
		return 0, 0, fmt.Errorf("port detection failed: %w", err)
	}

	// Step 4: Pick host port
	fmt.Println("Detected container port in isPortAvailable function:", containerPort)
	hostPort := containerPort
	if !utils.IsPortAvailable(hostPort) {
		hostPort, err = utils.FindFreeHostPort()
		if err != nil {
			logsCmd := exec.Command("docker", "logs", tmpContainer)
			logsOutput, _ := logsCmd.CombinedOutput()
			fmt.Println("Temporary container logs:\n", string(logsOutput))

			// _ = exec.Command("docker", "rm", "-f", tmpContainer).Run()
			return 0, 0, fmt.Errorf("failed to find free host port: %w", err)
		}
	}

	fmt.Println("Using host port through AuthorizeEC2Port: ", hostPort)
	if err := utils.AuthorizeEC2Port(hostPort); err != nil {
		// _ = exec.Command("docker", "rm", "-f", tmpContainer).Run()
		return 0, 0, fmt.Errorf("EC2 SG error: %w", err)
	}

	// Optional: Commit container state (e.g., installed files)
	_ = exec.Command("docker", "commit", tmpContainer, imageTag).Run()
	// _ = exec.Command("docker", "rm", "-f", tmpContainer).Run()

	fmt.Println("Making                                       final                         Container")
	// Step 5: Run final container
	finalCmd := exec.Command(
		"docker", "run", "-d",
		"-p", fmt.Sprintf("%d:%d", hostPort, containerPort),
		"--name", containerName,
		imageTag,
	)
	finalCmd.Stdout = os.Stdout
	finalCmd.Stderr = os.Stderr
	if err := finalCmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("docker final run failed: %w", err)
	}

	return containerPort, hostPort, nil
}



// FullPipeline executes the full flow: detects env, generates Dockerfile, builds, and runs container.
func FullPipeline(username,repoPath, envContent, startCommand string) (int, int, string, error) {
	// Step 1: Save .env if provided
	if envContent != "" {
		if err := utils.SaveEnvFile(repoPath, envContent); err != nil {
			return 0, 0, "", fmt.Errorf("failed to save .env: %w", err)
		}
	}

	// Step 2: Detect environment
	envType := detectEnvironment(repoPath)
	if envType == EnvUnknown {
		return 0, 0, "", fmt.Errorf("unsupported environment")
	}

	// Step 3: Generate Dockerfile dynamically
	if err := GenerateDockerfile(envType, repoPath, startCommand); err != nil {
		return 0, 0, "", fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Step 4: Derive container name from repo
	repoName := filepath.Base(repoPath)
	containerName := fmt.Sprintf("autoship-%s-%s", username, strings.ToLower(repoName))

	// Step 5: Build and run container
	containerPort, hostPort, err := buildAndRunContainerHybrid(repoPath, containerName)
	if err != nil {
		return 0, 0, "", fmt.Errorf("container error: %w", err)
	}

	log.Printf("Container %s started: hostPort=%d, containerPort=%d", containerName, hostPort, containerPort)
	return containerPort, hostPort, containerName, nil
}

