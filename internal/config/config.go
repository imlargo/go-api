package config

import (
	"log"
	"time"
)

type AppConfig struct {
	Server           ServerConfig           `yaml:"server" validate:"required"`
	Database         DbConfig               `yaml:"database" validate:"required"`
	RateLimiter      RateLimiterConfig      `yaml:"rate_limiter" validate:"required"`
	PushNotification PushNotificationConfig `yaml:"push_notification" validate:"required"`
	Auth             AuthConfig             `yaml:"auth" validate:"required"`
	Storage          StorageConfig          `yaml:"storage" validate:"required"`
	Redis            RedisConfig            `yaml:"redis" validate:"required"`
}

type ServerConfig struct {
	Host string `yaml:"host" validate:"required" example:"localhost"`
	Port string `yaml:"port" validate:"required" example:"8000"`
}

type RateLimiterConfig struct {
	RequestsPerTimeFrame int           `yaml:"requests_per_time_frame" validate:"min=0" example:"100"`
	TimeFrame            time.Duration `yaml:"time_frame" validate:"min=0" example:"60s"`
	Enabled              bool          `yaml:"enabled" example:"true"`
}

type PushNotificationConfig struct {
	VAPIDPublicKey  string `yaml:"vapid_public_key" validate:"required" example:"your-vapid-public-key"`
	VAPIDPrivateKey string `yaml:"vapid_private_key" validate:"required" example:"your-vapid-private-key"`
}

type AuthConfig struct {
	ApiKey string `yaml:"api_key,omitempty" example:"your-api-key"`

	JwtSecret         string        `yaml:"jwt_secret" validate:"required" example:"your-secret-key"`
	JwtIssuer         string        `yaml:"jwt_issuer" validate:"required" example:"your-app"`
	JwtAudience       string        `yaml:"jwt_audience" validate:"required" example:"your-app-users"`
	TokenExpiration   time.Duration `yaml:"token_expiration" validate:"min=1m" example:"15m"`
	RefreshExpiration time.Duration `yaml:"refresh_expiration" validate:"min=1h" example:"168h"`
}

type DbConfig struct {
	URL string `yaml:"url" validate:"required" example:"postgres://user:password@localhost/dbname?sslmode=disable"`
}

type StorageConfig struct {
	Enabled         bool   `yaml:"enabled" example:"true"`
	BucketName      string `yaml:"bucket_name" validate:"required_if=Enabled true" example:"my-bucket"`
	AccountID       string `yaml:"account_id" validate:"required_if=Enabled true" example:"account-id"`
	AccessKeyID     string `yaml:"access_key_id" validate:"required_if=Enabled true" example:"access-key"`
	SecretAccessKey string `yaml:"secret_access_key" validate:"required_if=Enabled true" example:"secret-key"`
	PublicDomain    string `yaml:"public_domain,omitempty" example:"https://cdn.example.com"`
	UsePublicURL    bool   `yaml:"use_public_url" example:"true"`
}

type RedisConfig struct {
	RedisURL string `yaml:"url" validate:"required" example:"redis://localhost:6379"`
}

func LoadConfig() AppConfig {
	// Try to load .env file for environment variables
	err := loadEnv()
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	// Try to load configuration from YAML files
	for _, configPath := range GetDefaultConfigPaths() {
		if config, err := LoadConfigFromYAML(configPath); err == nil {
			return *config
		}
	}

	// Fallback to environment variables only
	log.Println("No YAML configuration found, using environment variables only")
	config, err := LoadConfigFromYAML("") // Empty path forces env-only loading
	if err != nil {
		panic("Error loading configuration: " + err.Error())
	}

	return *config
}
