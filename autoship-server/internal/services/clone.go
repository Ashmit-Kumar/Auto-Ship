// internal/services/clone.go
package services

import (
	"fmt"
	"os"
	"os/exec"
)

// In future add support for cloning specific branches or commits

// This only clones main branch of the repo
func CloneRepository(repoURL, username, repoName string) (string, error) {
	path := fmt.Sprintf("static/%s/%s", username, repoName)

	if _, err := os.Stat(path); err == nil {
		return path, nil // already cloned
	}

	// Create new directory for the repository
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create dir: %v", err)
	}

	// Execute the git clone command
	cmd := exec.Command("git", "clone", repoURL, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failed: %v", err)
	}
	fmt.Println("Repository cloned successfully to:", path)
	return path, nil
}
