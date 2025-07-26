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

	fmt.Println("Cloning repository:", repoURL)
	// Run the command and check for errors
	fmt.Println("Executing command:", cmd.String())
	// cmd.Dir = "static" // Set working directory to static
	// cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	// cmd.Env = append(cmd.Env, "GIT_ASKPASS=echo") // Disable password prompts
	// cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no") // Disable strict host key checking
	// cmd.Env = append(cmd.Env, "GIT_USER_NAME="+username) // Set username for commits
	// cmd.Env = append(cmd.Env, "GIT_USER_EMAIL="+username+"@example.com") // Set email for commits
	// cmd.Env = append(cmd.Env, "GIT_COMMITTER_NAME="+username) // Set committer name
	// cmd.Env = append(cmd.Env, "GIT_COMMITTER_EMAIL="+username+"@example.com") // Set committer email
	// cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL_USER_NAME="+username) // Set global user name
	// cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL_USER_EMAIL="+username+"@example.com") // Set global user email
	// cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL_INIT_DEFAULT_BRANCH=main") // Set default branch to main
	// cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL_INIT_TEMPLATE_DIR=~/.git-templates") // Set global template directory
	// cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL_INIT_COMMITTER_NAME="+username		) // Set global committer name

	output, err := cmd.CombinedOutput()
	fmt.Println("Git Output:\n", string(output))
	fmt.Println("Git clone command finished.")
    fmt.Println("Git Output:\n", string(output))


	if err != nil {
	return "", fmt.Errorf("git clone failed: %v\n%s", err, output)
 }
	fmt.Println("Repository cloned successfully to:", path)
	return path, nil
}
