package ratelimiter

import "time"

type Limiter interface {
	Allow(key string) (bool, float64)
}

type Config struct {
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
	Enabled              bool
}
