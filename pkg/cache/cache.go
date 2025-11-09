package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
)

// InitializeRedis initializes Redis client
func InitializeRedis(addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	redisClient = client
	return client, nil
}

// GetClient returns the Redis client instance
func GetClient() *redis.Client {
	return redisClient
}

// CacheService provides caching functionality
type CacheService struct {
	client *redis.Client
}

// NewCacheService creates a new cache service
func NewCacheService() *CacheService {
	return &CacheService{
		client: redisClient,
	}
}

// Get retrieves a value from cache
func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrCacheMiss
	}
	if err != nil {
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Set stores a value in cache with TTL
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *CacheService) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			// Log error but continue deleting other keys
			continue
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return count > 0, nil
}

// Cache miss error
var ErrCacheMiss = fmt.Errorf("cache miss")

