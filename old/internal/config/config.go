package config

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/pkg/env"
)

type AppConfig struct {
	Server           ServerConfig
	Database         DbConfig
	RateLimiter      RateLimiterConfig
	PushNotification PushNotificationConfig
	Auth             AuthConfig
	Storage          StorageConfig
	Redis            RedisConfig
	TaskQueue        TaskQueueConfig
	External         ExternalConfig
	Stripe           StripeConfig
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

type TaskQueueConfig struct {
	WorkerCount             int
	TaskTimeout             time.Duration
	MaxRetries              int
	InitialRetryDelay       time.Duration
	MaxRetryDelay           time.Duration
	BackoffFactor           float64 // Hardcoded to 2.0 for standard exponential backoff to ensure consistent retry behavior across the system. Not configurable via env vars. If future requirements demand customization, consider making this value configurable.
	HeartbeatInterval       time.Duration
	OrphanTimeout           time.Duration
	PriorityHighThreshold   enums.TaskPriority
	PriorityNormalThreshold enums.TaskPriority
	DLQAlertThreshold       int
}

type ExternalConfig struct {
	InstagramApiKey string
	TikTokApiKey    string
	OnlyfansApiKey  string
	ShotstackApiKey string
	ResendApiKey    string
}

type StripeConfig struct {
	SecretKey              string
	PublishableKey         string
	WebhookSecret          string
	PriceStarter           string
	PriceGrowth            string
	PriceScale             string
	SubscriptionSuccessURL string
	SubscriptionCancelURL  string
	MarketplaceSuccessURL  string
	MarketplaceCancelURL   string
	PortalConfigurationID  string
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
			ApiKey: env.GetEnvString(PRIVATE_API_KEY, ""),

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
		TaskQueue: TaskQueueConfig{
			WorkerCount:             env.GetEnvInt(TASK_WORKER_COUNT, 7),
			TaskTimeout:             time.Duration(env.GetEnvInt(TASK_TIMEOUT, 30)) * time.Minute,
			MaxRetries:              env.GetEnvInt(TASK_MAX_RETRIES, 3),
			InitialRetryDelay:       time.Duration(env.GetEnvInt(TASK_INITIAL_RETRY_DELAY, 30)) * time.Second,
			MaxRetryDelay:           time.Duration(env.GetEnvInt(TASK_MAX_RETRY_DELAY, 30)) * time.Minute,
			BackoffFactor:           2.0,
			HeartbeatInterval:       time.Duration(env.GetEnvInt(TASK_HEARTBEAT_INTERVAL, 30)) * time.Second,
			OrphanTimeout:           time.Duration(env.GetEnvInt(TASK_ORPHAN_TIMEOUT, 30)) * time.Minute,
			PriorityHighThreshold:   enums.TaskPriority(env.GetEnvInt(TASK_PRIORITY_HIGH_THRESHOLD, enums.TaskPriorityHigh.Int())),
			PriorityNormalThreshold: enums.TaskPriority(env.GetEnvInt(TASK_PRIORITY_NORMAL_THRESHOLD, enums.TaskPriorityNormal.Int())),
			DLQAlertThreshold:       env.GetEnvInt(TASK_DLQ_ALERT_THRESHOLD, 10),
		},
		External: ExternalConfig{
			InstagramApiKey: env.GetEnvString(RAPIDAPI_INSTAGRAM_KEY, ""),
			TikTokApiKey:    env.GetEnvString(RAPIDAPI_TIKTOK_KEY, ""),
			OnlyfansApiKey:  env.GetEnvString(ONLYFANS_API_KEY, ""),
			ShotstackApiKey: env.GetEnvString(SHOTSTACK_API_KEY, ""),
			ResendApiKey:    env.GetEnvString(RESEND_API_KEY, ""),
		},
		Stripe: StripeConfig{
			SecretKey:              env.GetEnvString(STRIPE_SECRET_KEY, ""),
			PublishableKey:         env.GetEnvString(STRIPE_PUBLISHABLE_KEY, ""),
			WebhookSecret:          env.GetEnvString(STRIPE_WEBHOOK_SECRET, ""),
			PriceStarter:           env.GetEnvString(STRIPE_PRICE_STARTER, ""),
			PriceGrowth:            env.GetEnvString(STRIPE_PRICE_GROWTH, ""),
			PriceScale:             env.GetEnvString(STRIPE_PRICE_SCALE, ""),
			SubscriptionSuccessURL: env.GetEnvString(SUBSCRIPTION_SUCCESS_URL, ""),
			SubscriptionCancelURL:  env.GetEnvString(SUBSCRIPTION_CANCEL_URL, ""),
			MarketplaceSuccessURL:  env.GetEnvString(MARKETPLACE_SUCCESS_URL, ""),
			MarketplaceCancelURL:   env.GetEnvString(MARKETPLACE_CANCEL_URL, ""),
			PortalConfigurationID:  env.GetEnvString(STRIPE_PORTAL_CONFIGURATION_ID, ""),
		},
	}
}
