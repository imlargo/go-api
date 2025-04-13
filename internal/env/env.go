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

	// Validate required environment variables
	requiredEnvVars := []string{API_URL, PORT, SIA_URL}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			return fmt.Errorf("required environment variable %s is not set", envVar)
		}
	}

	return nil
}
