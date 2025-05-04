// internal/utils/username.go
package utils

import (
	"net/url"
	"strings"
	"fmt"
)

// ExtractUsernameFromRepoURL extracts the username from a GitHub repository URL
func ExtractUsernameFromRepoURL(repoURL string) (string, error) {
	fmt.Println("Extracting username from repo URL:", repoURL);
	// Ensure the URL starts with https:// or git@
	if !strings.HasPrefix(repoURL, "https://github.com/") && !strings.HasPrefix(repoURL, "git@github.com:") {
		return "", fmt.Errorf("invalid GitHub repository URL")
	}

	// If URL starts with git@github.com: convert it to https://github.com/
	if strings.HasPrefix(repoURL, "git@github.com:") {
		repoURL = "https://" + strings.TrimPrefix(repoURL, "git@")
		repoURL = strings.Replace(repoURL, ":", "/", 1)
	}

	// Parse the URL
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	// Extract the path part of the URL (username/repo-name)
	parts := strings.Split(u.Path, "/")

	// Ensure the URL has both username and repo
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid GitHub URL: %v", repoURL)
	}

	// The username is always the second part of the path
	username := parts[1]
	return username, nil
}
