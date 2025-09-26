# Configuration Management

This document describes the configuration management system for the go-api project, which supports both YAML files and environment variables with strict validation.

## Overview

The configuration system follows best practices by:
- Supporting structured YAML configuration files with strict schema validation
- Maintaining backward compatibility with environment variables
- Providing environment variable override capability
- Implementing comprehensive validation using Go's validator package
- Offering flexible configuration file locations

## Configuration Sources Priority

1. **Environment Variables** (highest priority) - override any YAML values
2. **YAML Configuration File** - structured configuration with validation
3. **Default Values** (lowest priority) - fallback values defined in code

## YAML Configuration

### Configuration File Locations

The system automatically searches for configuration files in the following order:

1. `./config.yaml`
2. `./config.yml`
3. `./configs/config.yaml`
4. `./configs/config.yml`
5. `$HOME/.config/go-api/config.yaml`
6. `/etc/go-api/config.yaml`

### YAML Schema

```yaml
server:
  host: localhost              # required: Server hostname
  port: "8000"                # required: Server port (string)

database:
  url: postgres://user:password@localhost/dbname?sslmode=disable  # required: Database connection URL

rate_limiter:
  requests_per_time_frame: 100  # min=0: Maximum requests per time frame
  time_frame: 60s              # min=0: Time frame duration (e.g., 30s, 5m, 1h)
  enabled: true                # Enable/disable rate limiting

push_notification:
  vapid_public_key: "your-vapid-public-key"    # required: VAPID public key
  vapid_private_key: "your-vapid-private-key"  # required: VAPID private key

auth:
  api_key: "your-api-key"                # optional: API key
  jwt_secret: "your-secret-key"          # required: JWT signing secret
  jwt_issuer: "your-app"                 # required: JWT issuer
  jwt_audience: "your-app-users"         # required: JWT audience
  token_expiration: 15m                  # min=1m: Token expiration (e.g., 15m, 1h)
  refresh_expiration: 168h               # min=1h: Refresh token expiration (e.g., 24h, 168h)

storage:
  enabled: false                         # Enable/disable storage features
  bucket_name: "my-bucket"              # required_if enabled: S3-compatible bucket name
  account_id: "account-id"              # required_if enabled: Storage account ID
  access_key_id: "access-key"           # required_if enabled: Storage access key
  secret_access_key: "secret-key"       # required_if enabled: Storage secret key
  public_domain: "https://cdn.example.com"  # optional: Public domain for file URLs
  use_public_url: true                  # Use public URLs for file access

redis:
  url: "redis://localhost:6379"         # required: Redis connection URL
```

### Time Duration Format

Time durations can be specified using Go's standard duration format:
- `30s` - 30 seconds
- `5m` - 5 minutes  
- `1h` - 1 hour
- `24h` - 24 hours
- `168h` - 168 hours (7 days)

## Environment Variables

All YAML configuration values can be overridden using environment variables:

```bash
# Server configuration
API_URL=localhost
PORT=8000

# Database configuration
DATABASE_URL=postgres://user:password@localhost/dbname?sslmode=disable

# JWT configuration
JWT_SECRET=your-secret-key
JWT_ISSUER=your-app
JWT_AUDIENCE=your-app-users
JWT_TOKEN_EXPIRATION=15        # in minutes
JWT_REFRESH_EXPIRATION=10080   # in minutes (7 days)

# Rate limiting
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_TIMEFRAME=60        # in seconds

# Push notifications
VAPID_PUBLIC_KEY=your-vapid-public-key
VAPID_PRIVATE_KEY=your-vapid-private-key

# Storage configuration
STORAGE_ENABLED=false
STORAGE_BUCKET_NAME=my-bucket
STORAGE_ACCOUNT_ID=account-id
STORAGE_ACCESS_KEY_ID=access-key
STORAGE_SECRET_ACCESS_KEY=secret-key
STORAGE_PUBLIC_DOMAIN=https://cdn.example.com
STORAGE_USE_PUBLIC_URL=true

# Redis configuration
REDIS_URL=redis://localhost:6379
```

## Validation

The configuration system includes comprehensive validation:

### Required Fields
- All `required` fields must have non-empty values
- `required_if` fields are required when the condition is met (e.g., storage fields when storage is enabled)

### Value Constraints
- **Time Durations**: Minimum values enforced (e.g., token expiration ≥ 1 minute)
- **Numeric Values**: Minimum values enforced (e.g., rate limiting ≥ 0)
- **URLs**: Database and Redis URLs must be valid connection strings

### Error Handling
- Validation errors are reported with detailed field information
- Missing required configuration stops application startup
- Invalid YAML syntax is reported with line numbers

## Migration from Environment Variables Only

If you're currently using only environment variables, no changes are required. The system maintains full backward compatibility:

1. **Existing setup**: Continue using environment variables as before
2. **Gradual migration**: Create a YAML file with some values, keep others as environment variables
3. **Full migration**: Move all configuration to YAML, use environment variables only for sensitive values or deployment-specific overrides

## Best Practices

1. **Use YAML for base configuration**: Define your default configuration in YAML files
2. **Use environment variables for secrets**: Override sensitive values like passwords and keys
3. **Use environment variables for deployment differences**: Override values that differ between environments (dev, staging, production)
4. **Version control**: Include `config.example.yaml` in version control, but exclude actual configuration files with secrets
5. **Validation**: Always test your configuration changes to catch validation errors early

## Example Usage

```go
// Load configuration (automatic discovery)
config := config.LoadConfig()

// Load from specific file
config, err := config.LoadConfigFromYAML("./my-config.yaml")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}
```

## Troubleshooting

### Common Issues

1. **Validation errors**: Check that all required fields are provided and meet constraints
2. **YAML syntax errors**: Validate your YAML file syntax using online tools or `yamllint`
3. **File not found**: Ensure your configuration file is in one of the searched locations
4. **Environment variable types**: Remember that environment variables for durations use different units (minutes/seconds) compared to YAML format