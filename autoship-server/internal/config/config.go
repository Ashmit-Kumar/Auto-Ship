// Package config centralizes environment-variable parsing so that .env is
// loaded exactly once at startup. Callers do config.Load() in main, then
// every other package reads typed fields via config.Get() instead of
// calling os.Getenv directly.
//
// Scope: this owns the env vars that previously had duplicate godotenv.Load
// sites (server wiring, Mongo, port allocator) plus what main needs to boot.
// Cloud-specific vars (S3_*, AZURE_*) and JWT secrets are still read lazily
// where they're used — those call sites run after Load, so they see the
// populated env without needing to own the load step.
package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string
	MongoURI        string
	MongoCollection string // collection name used by the port allocator
	CloudProvider   string
}

var (
	cfg      *Config
	loadOnce sync.Once
	loadErr  error
)

// Load reads .env (best-effort — a missing .env is fine for containerised
// runs where env vars come from the orchestrator) and resolves required
// environment variables into a typed Config. Idempotent: subsequent calls
// return the same instance without re-reading anything.
func Load() (*Config, error) {
	loadOnce.Do(func() {
		_ = godotenv.Load()

		mongoURI := os.Getenv("MONGO_URI")
		if mongoURI == "" {
			loadErr = fmt.Errorf("MONGO_URI is required")
			return
		}

		mongoCollection := os.Getenv("MONGO_DB_COLLECTION")
		if mongoCollection == "" {
			mongoCollection = "ports"
		}

		serverPort := os.Getenv("PORT")
		if serverPort == "" {
			serverPort = "3000"
		}

		cfg = &Config{
			ServerPort:      serverPort,
			MongoURI:        mongoURI,
			MongoCollection: mongoCollection,
			CloudProvider:   os.Getenv("CLOUD_PROVIDER"),
		}
	})
	return cfg, loadErr
}

// Get returns the loaded config. Panics if Load wasn't called first — a
// programmer error caught at startup, not a runtime condition worth
// handling at every call site.
func Get() *Config {
	if cfg == nil {
		panic("config.Get() called before config.Load()")
	}
	return cfg
}
