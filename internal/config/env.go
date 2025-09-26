package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Define enums for environment variable keys
const (
	API_URL              = "API_URL"
	PORT                 = "PORT"
	DATABASE_URL         = "DATABASE_URL"
	GOOGLE_CLIENT_ID     = "GOOGLE_CLIENT_ID"
	GOOGLE_CLIENT_SECRET = "GOOGLE_CLIENT_SECRET"
	GOOGLE_REDIRECT_URL  = "GOOGLE_REDIRECT_URL"

	JWT_SECRET             = "JWT_SECRET"
	JWT_ISSUER             = "JWT_ISSUER"
	JWT_AUDIENCE           = "JWT_AUDIENCE"
	JWT_TOKEN_EXPIRATION   = "JWT_TOKEN_EXPIRATION"
	JWT_REFRESH_EXPIRATION = "JWT_REFRESH_EXPIRATION"

	RATE_LIMIT_MAX_REQUESTS = "RATE_LIMIT_MAX_REQUESTS"
	RATE_LIMIT_TIMEFRAME    = "RATE_LIMIT_TIMEFRAME"

	VAPID_PUBLIC_KEY  = "VAPID_PUBLIC_KEY"
	VAPID_PRIVATE_KEY = "VAPID_PRIVATE_KEY"

	STORAGE_ENABLED           = "STORAGE_ENABLED"
	STORAGE_BUCKET_NAME       = "STORAGE_BUCKET_NAME"
	STORAGE_ACCOUNT_ID        = "STORAGE_ACCOUNT_ID"
	STORAGE_ACCESS_KEY_ID     = "STORAGE_ACCESS_KEY_ID"
	STORAGE_SECRET_ACCESS_KEY = "STORAGE_SECRET_ACCESS_KEY"
	STORAGE_PUBLIC_DOMAIN     = "STORAGE_PUBLIC_DOMAIN"
	STORAGE_USE_PUBLIC_URL    = "STORAGE_USE_PUBLIC_URL"

	REDIS_URL = "REDIS_URL"
)

// loadEnv loads environment variables from .env file and validates required ones
// This function is now more lenient since configuration can come from YAML
func loadEnv() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, proceeding with system environment variables")
	}

	// Define required environment variables when no YAML config is present
	// These are now validated in the YAML configuration system
	requiredEnvVars := []string{
		DATABASE_URL, // Always required for database connection
	}

	var missingEnvVars []string

	// Check for missing critical environment variables
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingEnvVars = append(missingEnvVars, envVar)
		}
	}

	// Only fail if critical variables are missing and no YAML config is available
	if len(missingEnvVars) > 0 {
		// Check if any YAML config file exists
		for _, configPath := range GetDefaultConfigPaths() {
			if _, err := os.Stat(configPath); err == nil {
				// YAML config found, environment validation is less strict
				log.Printf("YAML configuration found at %s, environment variable validation is relaxed", configPath)
				return nil
			}
		}
		
		// No YAML config found, require all environment variables
		return fmt.Errorf("no YAML configuration found and missing required environment variables: %v", missingEnvVars)
	}

	return nil
}
