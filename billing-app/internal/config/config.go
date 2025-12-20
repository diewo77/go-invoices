package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseDSN string
	Env         string
}

// Load loads configuration from environment with sensible defaults.
// Precedence: explicit env var > .env file (if loaded by user) > default.
func Load() Config {
	cfg := Config{}
	cfg.Port = getEnv("PORT", "8080")
	cfg.DatabaseDSN = getEnv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/billing?sslmode=disable")
	cfg.Env = getEnv("APP_ENV", "development")
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ParseBool reads an env var as bool with default.
func ParseBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			log.Printf("invalid boolean for %s: %s", key, v)
			return def
		}
		return b
	}
	return def
}
