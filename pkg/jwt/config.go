package jwt

import "time"

type JWTConfig struct {
	Secret            string
	Issuer            string
	Audience          string
	TokenExpiration   time.Duration
	RefreshExpiration time.Duration
}
