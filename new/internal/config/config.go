package config

import (
	"time"

	"github.com/imlargo/go-api/pkg/medusa/app"
	"github.com/imlargo/go-api/pkg/medusa/env"
)

type Config struct {
	app.Config
}

func LoadConfig() Config {
	err := env.CheckEnv([]string{
		API_URL,
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
				Host: env.GetEnvString(API_URL, "localhost"),
				Port: env.GetEnvString(PORT, "8000"),
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
	}
}
