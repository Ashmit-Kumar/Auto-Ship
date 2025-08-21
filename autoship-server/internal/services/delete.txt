package services

import (
	"context"
	"fmt"
	"log"

	"github.com/moby/moby/api/types"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// DeleteProject deletes a project's deployment.
// It stops and removes the Docker container associated with the project.
func DeleteProject(containerName string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	ctx := context.Background()

	// Stop the container
	log.Printf("Stopping container %s", containerName)
	if err := cli.ContainerStop(ctx, containerName, container.StopOptions{}); err != nil {
		// Log the error but continue, as we want to remove it even if stopping fails
		log.Printf("failed to stop container %s: %v. Proceeding with removal.", containerName, err)
	} else {
		log.Printf("Successfully stopped container %s", containerName)
	}

	// Remove the container
	log.Printf("Removing container %s", containerName)
	if err := cli.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true, // Force removal even if it's running
	}); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerName, err)
	}

	log.Printf("Successfully removed container %s", containerName)
	return nil
}
