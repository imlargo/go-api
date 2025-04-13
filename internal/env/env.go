package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Define enums para las variables de entorno
const (
	API_URL                 = "API_URL"
	PORT                    = "PORT"
	SIA_URL                 = "SIA_URL"
	RATE_LIMIT_MAX_REQUESTS = "RATE_LIMIT_MAX_REQUESTS"
	RATE_LIMIT_TIMEFRAME    = "RATE_LIMIT_TIMEFRAME"
)

// Initialize loads environment variables from .env file
func Initialize() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		return err
	}

	// Define required environment variables
	requiredEnvVars := []string{API_URL, PORT, SIA_URL}
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
