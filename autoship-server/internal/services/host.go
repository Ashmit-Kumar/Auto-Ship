package services

import (
	"os"
	"path/filepath"
)

func DetectHostProjectType(path string) string {
	hasHTML := false
	hasPackageJSON := false

	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		switch filepath.Base(p) {
		case "index.html":
			hasHTML = true
		case "package.json":
			hasPackageJSON = true
		}
		return nil
	})

		if hasPackageJSON {
			return "dynamic"
		}
		if hasHTML {
			return "static"
		}
		return "unknown"
	}

