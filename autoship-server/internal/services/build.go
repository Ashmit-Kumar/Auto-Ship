// internal/services/build.go
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	// "strings"
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

		// 1. Dynamic: package.json with "start" script
		pkgPath := filepath.Join(fullPath, "package.json")
		if fi, err := os.Stat(pkgPath); err == nil && !fi.IsDir() {
			data, err := os.ReadFile(pkgPath)
			if err == nil {
				var pkg struct {
					Scripts map[string]string `json:"scripts"`
				}
				if err := json.Unmarshal(data, &pkg); err == nil {
					if _, ok := pkg.Scripts["start"]; ok {
						return "dynamic"
					}
				}
			}
		}

		// 2. Dynamic: backend entrypoints
		for _, entry := range []string{"server.js", "app.js", "main.go"} {
			if fi, err := os.Stat(filepath.Join(fullPath, entry)); err == nil && !fi.IsDir() {
				return "dynamic"
			}
		}

		// 3. Static: index.html (untouched)
		if fi, err := os.Stat(filepath.Join(fullPath, "index.html")); err == nil && !fi.IsDir() {
			return "static"
		}
	}
	return "unknown"
}

