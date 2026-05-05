package utils

import (
	"context"
	"fmt"
	"time"

	conf "auth-service/src/conf"
)

// rateLimitKey builds a namespaced Redis key.
func rateLimitKey(action, identifier string) string {
	return fmt.Sprintf("rl:%s:%s", action, identifier)
}

// RateLimitRequest increments the counter for (action, identifier) on every
// call and returns true when the limit is exceeded. The window TTL is set only
// on the first increment, so the counter resets naturally after the window.
// Fails open (returns false) when Redis is unavailable.
func RateLimitRequest(action, identifier string, maxAttempts int64, window time.Duration) bool {
	if conf.Client == nil {
		return false
	}
	ctx := context.Background()
	key := rateLimitKey(action, identifier)
	count, err := conf.Client.Incr(ctx, key).Result()
	if err != nil {
		return false
	}
	if count == 1 {
		conf.Client.Expire(ctx, key, window)
	}
	return count > maxAttempts
}

// IsRateLimited returns true if the failure counter for (action, identifier)
// has reached or exceeded maxAttempts, without modifying the counter.
func IsRateLimited(action, identifier string, maxAttempts int64) bool {
	if conf.Client == nil {
		return false
	}
	ctx := context.Background()
	count, err := conf.Client.Get(ctx, rateLimitKey(action, identifier)).Int64()
	if err != nil {
		return false
	}
	return count >= maxAttempts
}

// RecordFailure increments the failure counter for (action, identifier).
// The TTL window is set only on the first failure.
func RecordFailure(action, identifier string, window time.Duration) {
	if conf.Client == nil {
		return
	}
	ctx := context.Background()
	key := rateLimitKey(action, identifier)
	count, err := conf.Client.Incr(ctx, key).Result()
	if err != nil {
		return
	}
	if count == 1 {
		conf.Client.Expire(ctx, key, window)
	}
}

// ClearFailures deletes the failure counter for (action, identifier).
func ClearFailures(action, identifier string) {
	if conf.Client == nil {
		return
	}
	conf.Client.Del(context.Background(), rateLimitKey(action, identifier))
}
