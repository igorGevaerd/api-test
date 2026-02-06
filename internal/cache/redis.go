package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// Client is a wrapper around Redis client.
type Client struct {
	client    *redis.Client
	Underlying *redis.Client // For testing purposes
}

// New creates a new Redis cache client.
func New(host, port string) *Client {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("✓ Connected to Redis")
	return &Client{client: client, Underlying: client}
}

// Get retrieves a value from cache.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set stores a value in cache with TTL.
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from cache.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// Underlying returns the underlying Redis client for advanced operations.
func (c *Client) Underlying() *redis.Client {
	return c.client
}
