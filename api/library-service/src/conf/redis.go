package conf

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

const MovieCacheTTL = 24 * time.Hour

func InitRedis() error {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return nil
}

func GetCache(key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("redis not initialized")
	}
	return RedisClient.Get(Ctx, key).Result()
}

func SetCache(key string, value string, ttl time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}
	return RedisClient.Set(Ctx, key, value, ttl).Err()
}

// GetCacheTTL returns the remaining TTL for a key.
func GetCacheTTL(key string) (time.Duration, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("redis not initialized")
	}
	return RedisClient.TTL(Ctx, key).Result()
}
