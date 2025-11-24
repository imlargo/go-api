package ratelimiter

import "time"

type Config struct {
	TimeFrame            time.Duration
	RequestsPerTimeFrame int
}
