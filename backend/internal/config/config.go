package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	ServerPort string
	ServerEnv  string

	// Database
	DatabaseURL string

	// JWT
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// Email
	SMTPHost string
	SMTPPort string
	SMTPFrom string

	// AI
	GeminiAPIKey string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file (ignore error if not found)
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	cfg := &Config{
		// Server defaults
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerEnv:  getEnv("SERVER_ENV", "development"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "host=127.0.0.1 port=5432 user=f1rivals dbname=f1rivals sslmode=disable"),

		// JWT
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-key"),
		JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "30m")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),

		// Email
		SMTPHost: getEnv("SMTP_HOST", "localhost"),
		SMTPPort: getEnv("SMTP_PORT", "1025"),
		SMTPFrom: getEnv("SMTP_FROM", "noreply@f1rivals.local"),

		// AI
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
	}

	return cfg, nil
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseDuration parses a duration string, returns default on error
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Minute
	}
	return d
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.ServerEnv == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.ServerEnv == "production"
}
