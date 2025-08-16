package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomID returns a short random hex id (16 chars).
func GenerateRandomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
