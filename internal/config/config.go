package config

import (
	"time"
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
			Host: GetEnv(API_URL, "localhost"),
			Port: GetEnv(PORT, "8000"),
		},
		Database: DbConfig{
			URL: GetEnv(DATABASE_URL, ""),
		},
		RateLimiter: RateLimiterConfig{
			RequestsPerTimeFrame: GetEnvInt(RATE_LIMIT_MAX_REQUESTS, 0),
			TimeFrame:            time.Duration(GetEnvInt(RATE_LIMIT_TIMEFRAME, 0)) * time.Second,
			Enabled:              GetEnvInt(RATE_LIMIT_MAX_REQUESTS, 0) != 0 && GetEnvInt(RATE_LIMIT_TIMEFRAME, 0) != 0,
		},
		PushNotification: PushNotificationConfig{
			VAPIDPublicKey:  GetEnv(VAPID_PUBLIC_KEY, ""),
			VAPIDPrivateKey: GetEnv(VAPID_PRIVATE_KEY, ""),
		},
		Auth: AuthConfig{
			JwtSecret:         GetEnv(JWT_SECRET, "your-secret-key"),
			JwtIssuer:         GetEnv(JWT_ISSUER, "your-app"),
			JwtAudience:       GetEnv(JWT_AUDIENCE, "your-app-users"),
			TokenExpiration:   time.Duration(GetEnvInt(JWT_TOKEN_EXPIRATION, 15)) * time.Minute,
			RefreshExpiration: time.Duration(GetEnvInt(JWT_REFRESH_EXPIRATION, 10080)) * time.Minute,
		},
		Storage: StorageConfig{
			BucketName:      GetEnv(STORAGE_BUCKET_NAME, ""),
			AccountID:       GetEnv(STORAGE_ACCOUNT_ID, ""),
			AccessKeyID:     GetEnv(STORAGE_ACCESS_KEY_ID, ""),
			SecretAccessKey: GetEnv(STORAGE_SECRET_ACCESS_KEY, ""),
			PublicDomain:    GetEnv(STORAGE_PUBLIC_DOMAIN, ""),
			UsePublicURL:    GetEnvBool(STORAGE_USE_PUBLIC_URL, false),
		},
		Redis: RedisConfig{
			RedisURL: GetEnv(REDIS_URL, ""),
		},
	}
}
