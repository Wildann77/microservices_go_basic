package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/microservices-go/shared/logger"
)

// Cache provides caching operations using Redis
type Cache struct {
	client *redis.Client
	prefix string
}

// NewCache creates a new cache instance
func NewCache(client *redis.Client, prefix string) *Cache {
	return &Cache{
		client: client,
		prefix: prefix,
	}
}

// key generates a cache key with prefix
func (c *Cache) key(key string) string {
	return fmt.Sprintf("cache:%s:%s", c.prefix, key)
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	if c.client == nil {
		return fmt.Errorf("cache client not initialized")
	}

	data, err := c.client.Get(ctx, c.key(key)).Result()
	if err == redis.Nil {
		return fmt.Errorf("key not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}

// Set stores a value in cache with expiration
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("cache client not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := c.client.Set(ctx, c.key(key), data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return nil
	}

	if err := c.client.Del(ctx, c.key(key)).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *Cache) DeletePattern(ctx context.Context, pattern string) error {
	if c.client == nil {
		return nil
	}

	fullPattern := c.key(pattern)
	iter := c.client.Scan(ctx, 0, fullPattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	log := logger.New("cache")
	log.WithField("pattern", pattern).WithField("count", len(keys)).Info("Cache invalidated")

	return nil
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	if c.client == nil {
		return false, nil
	}

	n, err := c.client.Exists(ctx, c.key(key)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return n > 0, nil
}

// Close closes the cache connection
func (c *Cache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
