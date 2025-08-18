package ports

type RateLimiter interface {
	Allow(key string) (bool, float64)
}
