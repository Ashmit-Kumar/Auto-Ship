package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// generateSubdomain creates a unique subdomain based on the repository name and a random number.
func generateSubdomain(repoName string) string {
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(1000000)
	return fmt.Sprintf("%s%06d.a.com", repoName, random)
}
