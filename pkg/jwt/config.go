package jwt

import "time"

type Config struct {
	Secret            string
	Issuer            string
	Audience          string
	TokenExpiration   time.Duration
	RefreshExpiration time.Duration
}
