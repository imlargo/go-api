package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigFromYAML(t *testing.T) {
	// Create a temporary config file
	configContent := `
server:
  host: "test-host"
  port: "9000"

database:
  url: "postgres://testuser:testpass@testhost/testdb"

rate_limiter:
  requests_per_time_frame: 50
  time_frame: 30s
  enabled: true

push_notification:
  vapid_public_key: "test-public-key"
  vapid_private_key: "test-private-key"

auth:
  jwt_secret: "test-secret"
  jwt_issuer: "test-issuer"
  jwt_audience: "test-audience"
  token_expiration: 10m
  refresh_expiration: 24h

storage:
  enabled: true
  bucket_name: "test-bucket"
  account_id: "test-account"
  access_key_id: "test-access-key"
  secret_access_key: "test-secret-key"
  use_public_url: false

redis:
  url: "redis://testhost:6379"
`

	// Write config to temporary file
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Test loading configuration
	config, err := LoadConfigFromYAML(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load YAML config: %v", err)
	}

	// Verify configuration values
	if config.Server.Host != "test-host" {
		t.Errorf("Expected host 'test-host', got '%s'", config.Server.Host)
	}
	
	if config.Server.Port != "9000" {
		t.Errorf("Expected port '9000', got '%s'", config.Server.Port)
	}

	if config.Database.URL != "postgres://testuser:testpass@testhost/testdb" {
		t.Errorf("Unexpected database URL: %s", config.Database.URL)
	}

	if config.RateLimiter.RequestsPerTimeFrame != 50 {
		t.Errorf("Expected 50 requests per time frame, got %d", config.RateLimiter.RequestsPerTimeFrame)
	}

	if config.RateLimiter.TimeFrame != 30*time.Second {
		t.Errorf("Expected 30s time frame, got %v", config.RateLimiter.TimeFrame)
	}

	if !config.RateLimiter.Enabled {
		t.Error("Expected rate limiter to be enabled")
	}

	if !config.Storage.Enabled {
		t.Error("Expected storage to be enabled")
	}

	if config.Storage.BucketName != "test-bucket" {
		t.Errorf("Expected bucket name 'test-bucket', got '%s'", config.Storage.BucketName)
	}
}

func TestLoadConfigFromYAML_ValidationError(t *testing.T) {
	// Create invalid config
	invalidConfigContent := `
server:
  host: ""  # This should fail validation
  port: "9000"

database:
  url: ""  # This should fail validation
`

	tmpFile, err := os.CreateTemp("", "invalid_config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidConfigContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// This should fail due to validation errors
	_, err = LoadConfigFromYAML(tmpFile.Name())
	if err == nil {
		t.Error("Expected validation error but got none")
	}
}

func TestLoadConfigFromYAML_NonExistentFile(t *testing.T) {
	// Test loading from non-existent file should fail validation when no env vars are set
	_, err := LoadConfigFromYAML("/non/existent/file.yaml")
	if err == nil {
		t.Error("Expected validation error when no config file exists and no env vars are set")
	}
}

func TestLoadConfigFromYAML_WithEnvOverride(t *testing.T) {
	// Create a basic config file
	configContent := `
server:
  host: "yaml-host"
  port: "8000"

database:
  url: "postgres://yamluser:yamlpass@yamlhost/yamldb"

rate_limiter:
  requests_per_time_frame: 100
  time_frame: 60s
  enabled: true

push_notification:
  vapid_public_key: "yaml-public-key"
  vapid_private_key: "yaml-private-key"

auth:
  jwt_secret: "yaml-secret"
  jwt_issuer: "yaml-issuer"
  jwt_audience: "yaml-audience"
  token_expiration: 15m
  refresh_expiration: 168h

storage:
  enabled: false
  bucket_name: "yaml-bucket"
  account_id: "yaml-account"
  access_key_id: "yaml-access-key"
  secret_access_key: "yaml-secret-key"
  use_public_url: false

redis:
  url: "redis://yamlhost:6379"
`

	tmpFile, err := os.CreateTemp("", "config_env_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Set environment variable to override YAML value
	os.Setenv("API_URL", "env-host")
	os.Setenv("DATABASE_URL", "postgres://envuser:envpass@envhost/envdb")
	defer func() {
		os.Unsetenv("API_URL")
		os.Unsetenv("DATABASE_URL")
	}()

	// Test loading configuration with environment override
	config, err := LoadConfigFromYAML(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load YAML config with env override: %v", err)
	}

	// Verify environment variable override
	if config.Server.Host != "env-host" {
		t.Errorf("Expected host 'env-host' (from env), got '%s'", config.Server.Host)
	}

	if config.Database.URL != "postgres://envuser:envpass@envhost/envdb" {
		t.Errorf("Expected database URL from env, got '%s'", config.Database.URL)
	}

	// Verify YAML values still used where no env override
	if config.PushNotification.VAPIDPublicKey != "yaml-public-key" {
		t.Errorf("Expected YAML VAPID key 'yaml-public-key', got '%s'", config.PushNotification.VAPIDPublicKey)
	}
}