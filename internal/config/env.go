package config

// Define enums for environment variable keys
const (
	HOST                                 = "HOST"
	PORT                                 = "PORT"
	DATABASE_URL                         = "DATABASE_URL"
	JWT_SECRET                           = "JWT_SECRET"
	JWT_TOKEN_EXPIRATION                 = "JWT_TOKEN_EXPIRATION"
	JWT_REFRESH_EXPIRATION               = "JWT_REFRESH_EXPIRATION"
	RATE_LIMITER_ENABLED                 = "RATE_LIMITER_ENABLED"
	RATE_LIMITER_REQUESTS_PER_TIME_FRAME = "RATE_LIMITER_REQUESTS_PER_TIME_FRAME"
	RATE_LIMITER_TIME_FRAME_MINUTES      = "RATE_LIMITER_TIME_FRAME_MINUTES"
)
