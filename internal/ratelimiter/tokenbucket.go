package ratelimiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type TokenBucketLimiter struct {
	sync.RWMutex
	config  Config
	entries map[string]*LimiterEntry
}

type LimiterEntry struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func NewTokenBucketLimiter(cfg Config) *TokenBucketLimiter {

	rl := &TokenBucketLimiter{
		config:  cfg,
		entries: make(map[string]*LimiterEntry),
	}

	go rl.cleanUpEntries()

	return rl
}

func (rl *TokenBucketLimiter) GetEntry(key string) *LimiterEntry {
	rl.Lock()
	_, exists := rl.entries[key]
	rl.Unlock()

	if !exists {
		limiter := rate.NewLimiter(rate.Every(rl.config.TimeFrame), rl.config.RequestsPerTimeFrame)
		rl.Lock()

		rl.entries[key] = &LimiterEntry{
			Limiter:  limiter,
			LastSeen: time.Now(),
		}

		rl.Unlock()
	}

	entry := rl.entries[key]
	entry.LastSeen = time.Now()

	return entry
}

func (rl *TokenBucketLimiter) Allow(key string) (bool, float64) {
	entry := rl.GetEntry(key)

	allowed := entry.Limiter.Allow()
	tokens := entry.Limiter.Tokens()

	return allowed, tokens
}

func (rl *TokenBucketLimiter) cleanUpEntries() {
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
