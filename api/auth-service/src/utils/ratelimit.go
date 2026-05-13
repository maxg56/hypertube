package utils

import (
	"context"
	"fmt"
	"time"

	conf "auth-service/src/conf"
	"github.com/go-redis/redis_rate/v10"
)

var rateLimiter *redis_rate.Limiter

// InitRateLimiter initializes the GCRA rate limiter backed by Redis.
// Must be called after conf.InitRedis.
func InitRateLimiter() {
	if conf.Client != nil {
		rateLimiter = redis_rate.NewLimiter(conf.Client)
	}
}

func rateLimitKey(action, identifier string) string {
	return fmt.Sprintf("rl:%s:%s", action, identifier)
}

func makeLimit(maxAttempts int64, window time.Duration) redis_rate.Limit {
	n := int(maxAttempts)
	return redis_rate.Limit{Rate: n, Burst: n, Period: window}
}

// RateLimitRequest atomically increments the counter and returns true when the
// limit is exceeded. Fails open (returns false) when Redis is unavailable.
func RateLimitRequest(action, identifier string, maxAttempts int64, window time.Duration) bool {
	if rateLimiter == nil {
		return false
	}
	res, err := rateLimiter.Allow(context.Background(), rateLimitKey(action, identifier), makeLimit(maxAttempts, window))
	if err != nil {
		return false
	}
	return res.Allowed == 0
}

// IsRateLimited peeks at the current counter (n=0, no token consumed) and
// returns true when no budget remains. Fails open when Redis is unavailable.
func IsRateLimited(action, identifier string, maxAttempts int64, window time.Duration) bool {
	if rateLimiter == nil {
		return false
	}
	res, err := rateLimiter.AllowN(context.Background(), rateLimitKey(action, identifier), makeLimit(maxAttempts, window), 0)
	if err != nil {
		return false
	}
	return res.Remaining == 0
}

// RecordFailure atomically consumes one token from the failure budget.
func RecordFailure(action, identifier string, maxAttempts int64, window time.Duration) {
	if rateLimiter == nil {
		return
	}
	rateLimiter.Allow(context.Background(), rateLimitKey(action, identifier), makeLimit(maxAttempts, window)) //nolint:errcheck
}

// ClearFailures resets the failure counter (deletes both GCRA state keys).
func ClearFailures(action, identifier string) {
	if conf.Client == nil {
		return
	}
	key := rateLimitKey(action, identifier)
	conf.Client.Del(context.Background(), key, key+".ts")
}
