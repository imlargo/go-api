package config

import (
	"time"

	"github.com/imlargo/go-api/pkg/env"
)

type AppConfig struct {
	Server           ServerConfig
	Database         DbConfig
	RateLimiter      RateLimiterConfig
	PushNotification PushNotificationConfig
	Auth             AuthConfig
	Storage          StorageConfig
	Redis            RedisConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type RateLimiterConfig struct {
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
	Enabled              bool
}

type PushNotificationConfig struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
}

type AuthConfig struct {
	ApiKey string

	JwtSecret         string
	JwtIssuer         string
	JwtAudience       string
	TokenExpiration   time.Duration
	RefreshExpiration time.Duration
}

type DbConfig struct {
	URL string
}

type StorageConfig struct {
	BucketName      string
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	PublicDomain    string // Optional domain
	UsePublicURL    bool   // Use public URL for accessing files
}

type RedisConfig struct {
	RedisURL string
}

func LoadConfig() AppConfig {
	err := loadEnv()
	if err != nil {
		panic("Error loading environment variables: " + err.Error())
	}

	return AppConfig{
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
