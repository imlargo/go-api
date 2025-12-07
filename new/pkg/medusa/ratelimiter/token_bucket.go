package ratelimiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type tokenBucketLimiter struct {
	sync.RWMutex
	config  Config
	entries map[string]*tblEntry
}

type tblEntry struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func NewTokenBucketLimiter(cfg Config) RateLimiter {
	rl := &tokenBucketLimiter{
		config:  cfg,
		entries: make(map[string]*tblEntry),
	}

	go rl.cleanUpEntries()

	return rl
}

func (rl *tokenBucketLimiter) getEntry(key string) *tblEntry {
	rl.Lock()
	_, exists := rl.entries[key]
	rl.Unlock()

	if !exists {
		limiter := rate.NewLimiter(rate.Every(rl.config.TimeFrame), rl.config.RequestsPerTimeFrame)
		rl.Lock()

		rl.entries[key] = &tblEntry{
			Limiter:  limiter,
			LastSeen: time.Now(),
		}

		rl.Unlock()
	}

	entry := rl.entries[key]
	entry.LastSeen = time.Now()

	return entry
}

func (rl *tokenBucketLimiter) Allow(key string) (bool, float64) {
	entry := rl.getEntry(key)

	allowed := entry.Limiter.Allow()
	tokens := entry.Limiter.Tokens()

	return allowed, tokens
}

func (rl *tokenBucketLimiter) cleanUpEntries() {
	maxAge := 3 * time.Minute
	timeInterval := time.Minute

	for {
		time.Sleep(timeInterval)

		rl.Lock()
		for key, entry := range rl.entries {
			if time.Since(entry.LastSeen) > maxAge {
				delete(rl.entries, key)
			}
		}
		rl.Unlock()
	}
}
