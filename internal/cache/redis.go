package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Client *redis.Client
}

func New(redisURL string) (*Cache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")
	return &Cache{Client: client}, nil
}

func (c *Cache) Close() error {
	return c.Client.Close()
}

// Health check for Redis
func (c *Cache) HealthCheck(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// JWT Blacklist operations
func (c *Cache) AddToBlacklist(ctx context.Context, jti string, expiration time.Duration) error {
	return c.Client.Set(ctx, fmt.Sprintf("auth:blacklist:%s", jti), "1", expiration).Err()
}

func (c *Cache) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	exists, err := c.Client.Exists(ctx, fmt.Sprintf("auth:blacklist:%s", jti)).Result()
	return exists > 0, err
}

// Session operations
func (c *Cache) SetSession(ctx context.Context, userID string, data interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, fmt.Sprintf("session:%s", userID), data, expiration).Err()
}

func (c *Cache) GetSession(ctx context.Context, userID string) (string, error) {
	return c.Client.Get(ctx, fmt.Sprintf("session:%s", userID)).Result()
}

func (c *Cache) DeleteSession(ctx context.Context, userID string) error {
	return c.Client.Del(ctx, fmt.Sprintf("session:%s", userID)).Err()
}

// Feed caching
func (c *Cache) SetUserFeed(ctx context.Context, userID string, feed interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, fmt.Sprintf("feed:user:%s", userID), feed, expiration).Err()
}

func (c *Cache) GetUserFeed(ctx context.Context, userID string) (string, error) {
	return c.Client.Get(ctx, fmt.Sprintf("feed:user:%s", userID)).Result()
}

func (c *Cache) InvalidateUserFeed(ctx context.Context, userID string) error {
	return c.Client.Del(ctx, fmt.Sprintf("feed:user:%s", userID)).Err()
}

// Rate limiting
func (c *Cache) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	pipe := c.Client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count, err := incr.Result()
	if err != nil {
		return false, err
	}

	return count <= int64(limit), nil
}
