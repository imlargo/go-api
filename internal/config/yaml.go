package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/imlargo/go-api/pkg/env"
	"gopkg.in/yaml.v3"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// LoadConfigFromYAML loads configuration from a YAML file with strict validation
// Falls back to environment variables for any missing values
func LoadConfigFromYAML(configPath string) (*AppConfig, error) {
	config := &AppConfig{}

	// Check if YAML config file exists
	if _, err := os.Stat(configPath); err == nil {
		// Load from YAML file
		yamlFile, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("error reading YAML config file: %w", err)
		}

		// Parse YAML
		if err := yaml.Unmarshal(yamlFile, config); err != nil {
			return nil, fmt.Errorf("error parsing YAML config: %w", err)
		}

		// Override with environment variables if they exist
		overrideWithEnvVars(config)
	} else {
		// If no YAML file exists, load entirely from environment variables
		config = loadFromEnvVars()
	}

	// Validate the final configuration
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// overrideWithEnvVars overrides YAML configuration with environment variables where they exist
func overrideWithEnvVars(config *AppConfig) {
	// Server configuration
	if host := env.GetEnvString(API_URL, ""); host != "" {
		config.Server.Host = host
	}
	if port := env.GetEnvString(PORT, ""); port != "" {
		config.Server.Port = port
	}

	// Database configuration
	if dbURL := env.GetEnvString(DATABASE_URL, ""); dbURL != "" {
		config.Database.URL = dbURL
	}

	// Rate limiter configuration
	if requests := env.GetEnvInt(RATE_LIMIT_MAX_REQUESTS, -1); requests != -1 {
		config.RateLimiter.RequestsPerTimeFrame = requests
		config.RateLimiter.Enabled = requests > 0
	}
	if timeframe := env.GetEnvInt(RATE_LIMIT_TIMEFRAME, -1); timeframe != -1 {
		config.RateLimiter.TimeFrame = time.Duration(timeframe) * time.Second
		if config.RateLimiter.RequestsPerTimeFrame > 0 {
			config.RateLimiter.Enabled = true
		}
	}

	// Push notification configuration
	if publicKey := env.GetEnvString(VAPID_PUBLIC_KEY, ""); publicKey != "" {
		config.PushNotification.VAPIDPublicKey = publicKey
	}
	if privateKey := env.GetEnvString(VAPID_PRIVATE_KEY, ""); privateKey != "" {
		config.PushNotification.VAPIDPrivateKey = privateKey
	}

	// Auth configuration
	if jwtSecret := env.GetEnvString(JWT_SECRET, ""); jwtSecret != "" {
		config.Auth.JwtSecret = jwtSecret
	}
	if jwtIssuer := env.GetEnvString(JWT_ISSUER, ""); jwtIssuer != "" {
		config.Auth.JwtIssuer = jwtIssuer
	}
	if jwtAudience := env.GetEnvString(JWT_AUDIENCE, ""); jwtAudience != "" {
		config.Auth.JwtAudience = jwtAudience
	}
	if tokenExp := env.GetEnvInt(JWT_TOKEN_EXPIRATION, -1); tokenExp != -1 {
		config.Auth.TokenExpiration = time.Duration(tokenExp) * time.Minute
	}
	if refreshExp := env.GetEnvInt(JWT_REFRESH_EXPIRATION, -1); refreshExp != -1 {
		config.Auth.RefreshExpiration = time.Duration(refreshExp) * time.Minute
	}

	// Storage configuration
	if enabled := os.Getenv(STORAGE_ENABLED); enabled != "" {
		config.Storage.Enabled = env.GetEnvBool(STORAGE_ENABLED, false)
	}
	if bucketName := env.GetEnvString(STORAGE_BUCKET_NAME, ""); bucketName != "" {
		config.Storage.BucketName = bucketName
	}
	if accountID := env.GetEnvString(STORAGE_ACCOUNT_ID, ""); accountID != "" {
		config.Storage.AccountID = accountID
	}
	if accessKeyID := env.GetEnvString(STORAGE_ACCESS_KEY_ID, ""); accessKeyID != "" {
		config.Storage.AccessKeyID = accessKeyID
	}
	if secretAccessKey := env.GetEnvString(STORAGE_SECRET_ACCESS_KEY, ""); secretAccessKey != "" {
		config.Storage.SecretAccessKey = secretAccessKey
	}
	if publicDomain := env.GetEnvString(STORAGE_PUBLIC_DOMAIN, ""); publicDomain != "" {
		config.Storage.PublicDomain = publicDomain
	}
	if usePublicURL := os.Getenv(STORAGE_USE_PUBLIC_URL); usePublicURL != "" {
		config.Storage.UsePublicURL = env.GetEnvBool(STORAGE_USE_PUBLIC_URL, false)
	}

	// Redis configuration
	if redisURL := env.GetEnvString(REDIS_URL, ""); redisURL != "" {
		config.Redis.RedisURL = redisURL
	}
}

// loadFromEnvVars loads configuration entirely from environment variables (legacy behavior)
func loadFromEnvVars() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Host: env.GetEnvString(API_URL, "localhost"),
			Port: env.GetEnvString(PORT, "8000"),
		},
		Database: DbConfig{
			URL: env.GetEnvString(DATABASE_URL, ""),
		},
		RateLimiter: RateLimiterConfig{
			RequestsPerTimeFrame: env.GetEnvInt(RATE_LIMIT_MAX_REQUESTS, 0),
			TimeFrame:            time.Duration(env.GetEnvInt(RATE_LIMIT_TIMEFRAME, 0)) * time.Second,
			Enabled:              env.GetEnvInt(RATE_LIMIT_MAX_REQUESTS, 0) != 0 && env.GetEnvInt(RATE_LIMIT_TIMEFRAME, 0) != 0,
		},
		PushNotification: PushNotificationConfig{
			VAPIDPublicKey:  env.GetEnvString(VAPID_PUBLIC_KEY, ""),
			VAPIDPrivateKey: env.GetEnvString(VAPID_PRIVATE_KEY, ""),
		},
		Auth: AuthConfig{
			JwtSecret:         env.GetEnvString(JWT_SECRET, "your-secret-key"),
			JwtIssuer:         env.GetEnvString(JWT_ISSUER, "your-app"),
			JwtAudience:       env.GetEnvString(JWT_AUDIENCE, "your-app-users"),
			TokenExpiration:   time.Duration(env.GetEnvInt(JWT_TOKEN_EXPIRATION, 15)) * time.Minute,
			RefreshExpiration: time.Duration(env.GetEnvInt(JWT_REFRESH_EXPIRATION, 10080)) * time.Minute,
		},
		Storage: StorageConfig{
			Enabled:         env.GetEnvBool(STORAGE_ENABLED, false),
			BucketName:      env.GetEnvString(STORAGE_BUCKET_NAME, ""),
			AccountID:       env.GetEnvString(STORAGE_ACCOUNT_ID, ""),
			AccessKeyID:     env.GetEnvString(STORAGE_ACCESS_KEY_ID, ""),
			SecretAccessKey: env.GetEnvString(STORAGE_SECRET_ACCESS_KEY, ""),
			PublicDomain:    env.GetEnvString(STORAGE_PUBLIC_DOMAIN, ""),
			UsePublicURL:    env.GetEnvBool(STORAGE_USE_PUBLIC_URL, false),
		},
		Redis: RedisConfig{
			RedisURL: env.GetEnvString(REDIS_URL, ""),
		},
	}
}

// GetDefaultConfigPaths returns the default paths to look for configuration files
func GetDefaultConfigPaths() []string {
	return []string{
		"./config.yaml",
		"./config.yml",
		"./configs/config.yaml",
		"./configs/config.yml",
		filepath.Join(os.Getenv("HOME"), ".config", "go-api", "config.yaml"),
		"/etc/go-api/config.yaml",
	}
}