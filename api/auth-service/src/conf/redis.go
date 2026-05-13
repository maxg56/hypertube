package conf

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var ctx = context.Background()

// InitRedis initializes the Redis connection
func InitRedis() error {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	db := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if dbInt, err := strconv.Atoi(dbStr); err == nil {
			db = dbInt
		}
	}

	password := os.Getenv("REDIS_PASSWORD")

	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	// Test connection
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

func blacklistKey(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return fmt.Sprintf("blacklist:%x", h)
}

// BlacklistToken adds a JWT token to the blacklist with TTL
func BlacklistToken(tokenString string, ttl time.Duration) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	err := Client.Set(ctx, blacklistKey(tokenString), "blacklisted", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func IsTokenBlacklisted(tokenString string) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("Redis client not initialized")
	}

	exists, err := Client.Exists(ctx, blacklistKey(tokenString)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}

	return exists > 0, nil
}

// InvalidateUserTokens stores a timestamp in Redis so any token issued before
// this moment is considered revoked (e.g. after a password reset).
func InvalidateUserTokens(userID string) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	key := "user_invalidated:" + userID
	return Client.Set(ctx, key, time.Now().Unix(), 7*24*time.Hour).Err()
}

