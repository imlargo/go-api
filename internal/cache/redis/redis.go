package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nicolailuther/butter/pkg/kv"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) kv.KvProvider {
	return &redisCache{
		client: client,
	}
}

func (r *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
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

func (r *redisCache) Get(key string) (string, error) {
	result, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found", key)
		}
		return "", fmt.Errorf("error getting cache key %s: %w", key, err)
	}

	return result, nil
}

func (r *redisCache) Delete(key string) error {
	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		return fmt.Errorf("error deleting cache key %s: %w", key, err)
	}

	return nil
}

func (r *redisCache) Exists(key string) (bool, error) {
	result, err := r.client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, fmt.Errorf("error checking if key %s exists: %w", key, err)
	}

	return result > 0, nil
}

func (r *redisCache) Ping() error {
	err := r.client.Ping(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("error pinging Redis: %w", err)
	}

	return nil
}
