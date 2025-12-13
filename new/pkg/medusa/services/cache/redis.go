package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyNotFound = errors.New("key not found in cache")
	ErrNilValue    = errors.New("nil value provided")
)

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) Service {
	return &redisCache{
		client: client,
	}
}

// Get retrieves a value from cache and deserializes it into dest
func (r *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return err
	}

	if dest == nil {
		return ErrNilValue
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache with optional TTL (0 = no expiration)
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a key from cache
func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache
func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Clear flushes the entire cache (use with caution)
func (r *redisCache) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Remember gets from cache or executes fn and stores the result
func (r *redisCache) Remember(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) error {
	// Try to get from cache first
	exists, err := r.Exists(ctx, key)
	if err != nil {
		return err
	}

	if exists {
		return nil // Value already in cache
	}

	// Execute function to get value
	value, err := fn()
	if err != nil {
		return err
	}

	// Store in cache
	return r.Set(ctx, key, value, ttl)
}

// GetOrSet retrieves a value or sets a default if it doesn't exist
func (r *redisCache) GetOrSet(ctx context.Context, key string, defaultValue interface{}, ttl time.Duration, dest interface{}) error {
	err := r.Get(ctx, key, dest)
	if err == nil {
		return nil // Value found
	}

	if err != ErrKeyNotFound {
		return err // Real error occurred
	}

	// Key not found, set default value
	if err := r.Set(ctx, key, defaultValue, ttl); err != nil {
		return err
	}

	// Marshal and unmarshal to populate dest
	data, err := json.Marshal(defaultValue)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// SetMultiple stores multiple key-value pairs
func (r *redisCache) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	pipe := r.client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetMultiple retrieves multiple values
func (r *redisCache) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue // Skip missing keys
		}
		if err != nil {
			return nil, err
		}

		var data interface{}
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			return nil, err
		}
		result[key] = data
	}

	return result, nil
}

// DeletePattern removes all keys matching the pattern
func (r *redisCache) DeletePattern(ctx context.Context, pattern string) (int64, error) {
	var cursor uint64
	var deletedCount int64

	for {
		keys, newCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return deletedCount, err
		}

		if len(keys) > 0 {
			deleted, err := r.client.Del(ctx, keys...).Result()
			if err != nil {
				return deletedCount, err
			}
			deletedCount += deleted
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return deletedCount, nil
}

// Increment increments a counter
func (r *redisCache) Increment(ctx context.Context, key string, amount int64) (int64, error) {
	return r.client.IncrBy(ctx, key, amount).Result()
}

// Decrement decrements a counter
func (r *redisCache) Decrement(ctx context.Context, key string, amount int64) (int64, error) {
	return r.client.DecrBy(ctx, key, amount).Result()
}

// TTL gets the remaining time to live of a key
func (r *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Expire sets or updates the expiration time of a key
func (r *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}
