package config

import (
	"time"

	"github.com/imlargo/go-api/pkg/medusa/core/app"
	"github.com/imlargo/go-api/pkg/medusa/core/env"
	"github.com/imlargo/go-api/pkg/medusa/services/storage"
)

type Config struct {
	app.Config
	RateLimiter RateLimiterConfig
	Storage     storage.StorageConfig
	Redis       RedisConfig
}

type RateLimiterConfig struct {
	Enabled              bool
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
}

type RedisConfig struct {
	RedisURL string
}

func LoadConfig() Config {
	err := env.CheckEnv([]string{
		HOST,
		PORT,
		DATABASE_URL,
		JWT_SECRET,
		JWT_TOKEN_EXPIRATION,
		JWT_REFRESH_EXPIRATION,
	})

	if err != nil {
		panic("Error loading environment variables: " + err.Error())
	}

	return Config{
		Config: app.Config{
			Server: app.ServerConfig{
				Host: env.GetEnvString(HOST, "localhost"),
				Port: env.GetEnvInt(PORT, 8000),
			},
			Database: app.DbConfig{
				URL: env.GetEnvString(DATABASE_URL, ""),
			},
			Auth: app.AuthConfig{
				JwtSecret:         env.GetEnvString(JWT_SECRET, "your-secret-key"),
				TokenExpiration:   time.Duration(env.GetEnvInt(JWT_TOKEN_EXPIRATION, 15)) * time.Minute,
				RefreshExpiration: time.Duration(env.GetEnvInt(JWT_REFRESH_EXPIRATION, 10080)) * time.Minute,
			},
		},
		RateLimiter: RateLimiterConfig{
			Enabled:              env.GetEnvBool(RATE_LIMITER_ENABLED, true),
			RequestsPerTimeFrame: env.GetEnvInt(RATE_LIMITER_REQUESTS_PER_TIME_FRAME, 100),
			TimeFrame:            time.Duration(env.GetEnvInt(RATE_LIMITER_TIME_FRAME_MINUTES, 1)) * time.Minute,
		},
	}
}
