package auth

import (
	"os"
	"time"
)

type JWTConfig struct {
	Secret            string
	Issuer            string
	Audience          string
	TokenExpiration   time.Duration
	RefreshExpiration time.Duration
}

func getJWTConfig() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	issuer := os.Getenv("JWT_ISSUER")
	aud := os.Getenv("JWT_AUDIENCE")

	tokenExp := 10 * time.Minute
	refreshExp := 7 * 24 * time.Hour

	cfg := JWTConfig{
		Secret:            secret,
		Issuer:            issuer,
		Audience:          aud,
		TokenExpiration:   tokenExp,
		RefreshExpiration: refreshExp,
	}

	return cfg
}
