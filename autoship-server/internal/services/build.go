package services

import (
	"os"
	"path/filepath"
)

func DetectProjectType(projectPath string) string {
	if _, err := os.Stat(filepath.Join(projectPath, "package.json")); err == nil {
		return "dynamic"
	}
	if _, err := os.Stat(filepath.Join(projectPath, "index.html")); err == nil {
		return "static"
	}
	return "unknown"
}
