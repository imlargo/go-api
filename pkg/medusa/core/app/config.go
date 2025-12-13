package app

import "time"

type Config struct {
	Server   ServerConfig
	Database DbConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type AuthConfig struct {
	JwtSecret         string
	JwtIssuer         string
	TokenExpiration   time.Duration
	RefreshExpiration time.Duration
}

type DbConfig struct {
	URL string
}
