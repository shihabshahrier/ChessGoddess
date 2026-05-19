// Package config loads and validates application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL    string
	RedisURL       string
	GoogleClientID string
	GoogleSecret   string
	JWTSecret      string
	AllowedOrigins []string
	R2AccessKey    string
	R2SecretKey    string
	R2Bucket       string
	R2Endpoint     string
	OpenRouterKey  string
	StockfishPath  string
	Port           string
	Environment    string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/chesslens?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		R2AccessKey:    getEnv("R2_ACCESS_KEY", ""),
		R2SecretKey:    getEnv("R2_SECRET_KEY", ""),
		R2Bucket:       getEnv("R2_BUCKET", "chesslens"),
		R2Endpoint:     getEnv("R2_ENDPOINT", ""),
		OpenRouterKey:  getEnv("OPENROUTER_API_KEY", ""),
		StockfishPath:  getEnv("STOCKFISH_PATH", "stockfish"),
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		AllowedOrigins: parseOrigins(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
	}

	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if c.GoogleClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	if c.GoogleSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	if c.Environment == "production" && c.JWTSecret == "dev-secret-change-in-production" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value in production")
	}
	return nil
}

func (c *Config) GoogleRedirectURL() string {
	if redirect := getEnv("GOOGLE_REDIRECT_URL", ""); redirect != "" {
		return redirect
	}
	return "http://localhost:8080/api/v1/auth/google/callback"
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseOrigins(raw string) []string {
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}
