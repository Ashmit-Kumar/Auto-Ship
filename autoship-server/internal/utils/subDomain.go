package utils

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// GenerateSubdomain creates a unique subdomain based on the repository name
// and the provided domain (falls back to HOSTINGER_DOMAIN env or "autoship.site").
// Returned format: <sanitized-repo-name><6-digit-rand>.<domain>
// Example: my-repo012345.example.com
func GenerateSubdomain(repoName, domain string) string {
	// if domain not provided, try env
	if domain == "" {
		domain = os.Getenv("HOSTINGER_DOMAIN")
		if domain == "" {
			domain = os.Getenv("DOMAIN")
		}
		if domain == "" {
			domain = "autoship.site"
		}
	}

	// sanitize repo name: lowercase, replace spaces/underscores with hyphens, keep alnum and hyphen
	s := strings.ToLower(repoName)
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	sanitized := b.String()
	if sanitized == "" {
		sanitized = "app"
	}

	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(1000000)

	return fmt.Sprintf("%s%06d.%s", sanitized, random, domain)
}
