package services

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// DeleteProject deletes a project's deployment by stopping and removing the Docker container.
func DeleteProject(containerName string) error {
	if containerName == "" {
		return fmt.Errorf("container name required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop the container (best-effort)
	log.Printf("Stopping container %s", containerName)
	stopCmd := exec.CommandContext(ctx, "docker", "stop", containerName)
	if out, err := stopCmd.CombinedOutput(); err != nil {
		log.Printf("failed to stop container %s: %v, output: %s", containerName, err, string(out))
	} else {
		log.Printf("Successfully stopped container %s: %s", containerName, string(out))
	}

	// Remove the container (force + remove volumes)
	log.Printf("Removing container %s", containerName)
	rmCmd := exec.CommandContext(ctx, "docker", "rm", "-f", "-v", containerName)
	if out, err := rmCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove container %s: %v, output: %s", containerName, err, string(out))
	} else {
		log.Printf("Successfully removed container %s: %s", containerName, string(out))
	}

	return nil
}
