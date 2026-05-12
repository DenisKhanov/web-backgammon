package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	Port           string
	AllowedOrigins []string
	MigrationsDir  string
}

func Load() (*Config, error) {
	// Load .env if present; ignore error (file may not exist in production).
	// Try current directory first, then parent (for running from backend/).
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://bg_user:bg_pass@localhost:5433/backgammon?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	originsRaw := os.Getenv("ALLOWED_ORIGINS")
	if originsRaw == "" {
		originsRaw = "http://localhost:3000"
	}
	origins := strings.Split(originsRaw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return &Config{
		DatabaseURL:    dsn,
		Port:           port,
		AllowedOrigins: origins,
		MigrationsDir:  migrationsDir,
	}, nil
}
