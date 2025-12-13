package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache service
func NewRedisCache(client *redis.Client) Service {
	return &redisCache{
		client: client,
	}
}

// Get retrieves a value from cache and deserializes it into dest
func (r *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	if dest == nil {
		return ErrNilValue
	}

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return nil
}

// Set stores a value in cache with optional TTL (0 = no expiration)
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Delete removes a key from cache
func (r *redisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return count > 0, nil
}

// Clear flushes the entire cache (use with caution)
func (r *redisCache) Clear(ctx context.Context) error {
	if err := r.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

// Remember gets from cache or executes fn, stores and returns the result in dest
func (r *redisCache) Remember(ctx context.Context, key string, ttl time.Duration, dest interface{}, fn func() (interface{}, error)) error {
	if dest == nil {
		return ErrNilValue
	}

	// Try to get from cache first
	err := r.Get(ctx, key, dest)
	if err == nil {
		return nil // Value found in cache
	}

	if err != ErrKeyNotFound {
		return err // Real error occurred
	}

	// Value not in cache, execute function
	value, err := fn()
	if err != nil {
		return fmt.Errorf("function execution failed: %w", err)
	}

	// Store in cache
	if err := r.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Copy value to dest
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal function result: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal function result: %w", err)
	}

	return nil
}

// GetOrSet retrieves a value or sets a default if it doesn't exist
func (r *redisCache) GetOrSet(ctx context.Context, key string, defaultValue interface{}, ttl time.Duration, dest interface{}) error {
	if dest == nil {
		return ErrNilValue
	}

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

	// Copy default value to dest
	data, err := json.Marshal(defaultValue)
	if err != nil {
		return fmt.Errorf("failed to marshal default value: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal default value: %w", err)
	}

	return nil
}

// SetMultiple stores multiple key-value pairs using pipeline for efficiency
func (r *redisCache) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	return nil
}

// GetMultiple retrieves multiple values using pipeline for efficiency
func (r *redisCache) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd, len(keys))

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	result := make(map[string]interface{})
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue // Skip missing keys
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get key %s: %w", key, err)
		}

		var data interface{}
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
		}
		result[key] = data
	}

	return result, nil
}

// DeletePattern removes all keys matching the pattern using SCAN to avoid blocking
func (r *redisCache) DeletePattern(ctx context.Context, pattern string) (int64, error) {
	var cursor uint64
	var deletedCount int64
	const batchSize = 100

	for {
		keys, newCursor, err := r.client.Scan(ctx, cursor, pattern, batchSize).Result()
		if err != nil {
			return deletedCount, fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
		}

		if len(keys) > 0 {
			deleted, err := r.client.Del(ctx, keys...).Result()
			if err != nil {
				return deletedCount, fmt.Errorf("failed to delete keys: %w", err)
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

// Increment increments a counter atomically
func (r *redisCache) Increment(ctx context.Context, key string, amount int64) (int64, error) {
	result, err := r.client.IncrBy(ctx, key, amount).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result, nil
}

// Decrement decrements a counter atomically
func (r *redisCache) Decrement(ctx context.Context, key string, amount int64) (int64, error) {
	result, err := r.client.DecrBy(ctx, key, amount).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return result, nil
}

// TTL gets the remaining time to live of a key
func (r *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return ttl, nil
}

// Expire sets or updates the expiration time of a key
func (r *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}
	return nil
}
