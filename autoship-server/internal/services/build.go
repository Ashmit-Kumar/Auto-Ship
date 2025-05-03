// internal/services/build.go
package services

import (
	"os"
	"path/filepath"
)

func DetectProjectType(projectPath string) string {
	commonDirs := []string{
		".",             // root
		"public",        // React, Vite, etc.
		"frontend",      // custom structure
		"dist",          // production build
		"build",         // output folder
	}

	for _, dir := range commonDirs {
		fullPath := filepath.Join(projectPath, dir)
		if _, err := os.Stat(filepath.Join(fullPath, "package.json")); err == nil {
			return "dynamic"
		}
		if _, err := os.Stat(filepath.Join(fullPath, "index.html")); err == nil {
			return "static"
		}
	}
	return "unknown"
}

