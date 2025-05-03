// internal/services/clone.go
package services

import (
	"fmt"
	"os"
	"os/exec"
)

func CloneRepository(repoURL, username, repoName string) (string, error) {
	path := fmt.Sprintf("static/%s/%s", username, repoName)

	if _, err := os.Stat(path); err == nil {
		return path, nil // already cloned
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create dir: %v", err)
	}

	cmd := exec.Command("git", "clone", repoURL, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failed: %v", err)
	}

	return path, nil
}
