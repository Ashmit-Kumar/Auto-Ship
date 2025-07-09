// internal/services/build.go
package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DetectProjectType(projectPath string) string {
	commonDirs := []string{
		".",             // root
		"public",        // React, Vite, etc.
		"frontend",      // custom structure
		"dist",          // production build
		"build",         // output folder
	}
	fmt.Println("Detecting project type in:", projectPath)

	for _, dir := range commonDirs {
		fullPath := filepath.Join(projectPath, dir)
		fmt.Println("Checking directory:", fullPath)
		// New dynamic detection logic:
		// 1. If package.json exists and contains a 'start' script, treat as dynamic
		pkgPath := filepath.Join(fullPath, "package.json")
		if fi, err := os.Stat(pkgPath); err == nil && !fi.IsDir() {
			data, err := os.ReadFile(pkgPath)
			if err == nil && strings.Contains(string(data), "\"start\"") {
				return "dynamic"
			}
			// Also treat as dynamic if server.js, app.js, or main.go exists in the same dir
			for _, entry := range []string{"server.js", "app.js", "main.go"} {
				if _, err := os.Stat(filepath.Join(fullPath, entry)); err == nil {
					return "dynamic"
				}
			}
		}
		if _, err := os.Stat(filepath.Join(fullPath, "index.html")); err == nil {
			return "static"
		}
	}
	return "unknown"
}

