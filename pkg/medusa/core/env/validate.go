package env

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file, proceeding with system environment variables")
	}
}

// Initialize loads environment variables from .env file
func CheckEnv(required []string) error {
	var missingEnvVars []string

	LoadEnv()

	// Check for missing environment variables
	for _, envVar := range required {
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
