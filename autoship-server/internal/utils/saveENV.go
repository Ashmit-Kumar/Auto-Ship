package utils

import (
    "os"
    "path/filepath"
)

func SaveEnvFile(repoPath, envContent string) error {
    envPath := filepath.Join(repoPath, ".env")
    return os.WriteFile(envPath, []byte(envContent), 0644)
}
