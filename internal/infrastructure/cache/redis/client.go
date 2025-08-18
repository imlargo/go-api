package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(redisURL string) (*RedisClient, error) {
	parsedURL, err := url.Parse(redisURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing Redis URL: %w", err)
	}

	username := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()

	opts := &redis.Options{
		Addr:     parsedURL.Host,
		Username: username,
		Password: password,
		DB:       0,

		PoolSize:     10,
		MinIdleConns: 5,
		MaxIdleConns: 10,

		DialTimeout:  10 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	}

	if parsedURL.Scheme == "rediss" {
		opts.TLSConfig = &tls.Config{}
	}

	fmt.Printf("[Redis] Connecting to %s (TLS: %v)\n", opts.Addr, opts.TLSConfig != nil)
	fmt.Printf("[Redis] Username: %s | Password present: %v\n", opts.Username, opts.Password != "")

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Printf("[Redis] Ping failed: %+v\n", err)
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	log.Println("[Redis] Connected successfully.")
	return &RedisClient{client: client}, nil
}

func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
