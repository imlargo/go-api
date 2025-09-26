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

// Initialize loads environment variables from .env file
func loadEnv() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file, proceeding with system environment variables")
	}

	// Define required environment variables
	requiredEnvVars := []string{
		API_URL,
		PORT,
		DATABASE_URL,
		JWT_SECRET,
		JWT_ISSUER,
		JWT_AUDIENCE,
		JWT_TOKEN_EXPIRATION,
		JWT_REFRESH_EXPIRATION,
		RATE_LIMIT_MAX_REQUESTS,
		RATE_LIMIT_TIMEFRAME,
		VAPID_PUBLIC_KEY,
		VAPID_PRIVATE_KEY,
		STORAGE_BUCKET_NAME,
		STORAGE_ACCOUNT_ID,
		STORAGE_ACCESS_KEY_ID,
		STORAGE_SECRET_ACCESS_KEY,
		STORAGE_PUBLIC_DOMAIN,
		STORAGE_USE_PUBLIC_URL,
		REDIS_URL,
		STORAGE_ENABLED,
	}

	var missingEnvVars []string

	// Check for missing environment variables
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingEnvVars = append(missingEnvVars, envVar)
		}
	}

	// If there are missing variables, return an error listing them
	if len(missingEnvVars) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missingEnvVars)
	}

	return nil
}
