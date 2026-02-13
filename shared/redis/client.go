package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/microservices-go/shared/config"
	"github.com/microservices-go/shared/logger"
)

// Client wraps redis.Client for rate limiting
type Client struct {
	client *redis.Client
	ctx    context.Context
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log := logger.New("redis")
	log.WithField("addr", cfg.Addr()).Info("Connected to Redis")

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying redis client
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// Increment increments the counter for a key and sets expiration if new
func (c *Client) Increment(key string, window time.Duration) (int64, error) {
	pipe := c.client.Pipeline()

	incr := pipe.Incr(c.ctx, key)
	pipe.Expire(c.ctx, key, window)

	_, err := pipe.Exec(c.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}

	return incr.Val(), nil
}

// GetTTL returns the remaining TTL for a key
func (c *Client) GetTTL(key string) (time.Duration, error) {
	ttl, err := c.client.TTL(c.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// Delete deletes a key
func (c *Client) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}
