package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/imlargo/go-api-template/internal/infrastructure/cache"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCacheRepository(redisClient *RedisClient) cache.CacheRepository {
	return &RedisCache{
		client: redisClient.GetClient(),
	}
}

func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshaling value: %w", err)
	}

	err = r.client.Set(context.Background(), key, jsonValue, expiration).Err()
	if err != nil {
		return fmt.Errorf("error setting cache key %s: %w", key, err)
	}

	return nil
}

func (r *RedisCache) Get(key string) (string, error) {
	result, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found", key)
		}
		return "", fmt.Errorf("error getting cache key %s: %w", key, err)
	}

	return result, nil
}

func (r *RedisCache) Delete(key string) error {
	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		return fmt.Errorf("error deleting cache key %s: %w", key, err)
	}

	return nil
}

func (r *RedisCache) Exists(key string) (bool, error) {
	result, err := r.client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, fmt.Errorf("error checking if key %s exists: %w", key, err)
	}

	return result > 0, nil
}

func (r *RedisCache) Ping() error {
	err := r.client.Ping(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("error pinging Redis: %w", err)
	}

	return nil
}
