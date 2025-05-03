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

// func DetectProjectType(projectPath string) string {
// 	hasHTML := false
// 	hasPackageJSON := false

// 	filepath.Walk(projectPath, func(p string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return nil // skip on error
// 		}
// 		switch filepath.Base(p) {
// 		case "index.html":
// 			hasHTML = true
// 		case "package.json":
// 			hasPackageJSON = true
// 		}
// 		return nil
// 	})

// 	if hasPackageJSON {
// 		return "dynamic"
// 	}
// 	if hasHTML {
// 		return "static"
// 	}
// 	return "unknown"
// }



// func DetectProjectType(projectPath string) string {
// 	if _, err := os.Stat(filepath.Join(projectPath, "package.json")); err == nil {
// 		return "dynamic"
// 	}
// 	if _, err := os.Stat(filepath.Join(projectPath, "index.html")); err == nil {
// 		return "static"
// 	}
// 	return "unknown"
// }
