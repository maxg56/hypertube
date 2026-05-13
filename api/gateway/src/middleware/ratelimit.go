package middleware

import (
	"net/http"
	"sync"
	"time"

	"gateway/src/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limitersMu  sync.Mutex
	ipLimiters  = make(map[string]*ipLimiter)
	globalRPS   rate.Limit
	globalBurst int
)

// InitRateLimiter sets the per-IP token-bucket parameters.
func InitRateLimiter(maxRPS int) {
	globalRPS = rate.Limit(maxRPS)
	globalBurst = maxRPS
	go cleanupLimiters()
}

func getLimiter(ip string) *rate.Limiter {
	limitersMu.Lock()
	defer limitersMu.Unlock()
	l, ok := ipLimiters[ip]
	if !ok {
		l = &ipLimiter{limiter: rate.NewLimiter(globalRPS, globalBurst)}
		ipLimiters[ip] = l
	}
	l.lastSeen = time.Now()
	return l.limiter
}

func cleanupLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		limitersMu.Lock()
		for ip, l := range ipLimiters {
			if time.Since(l.lastSeen) > 10*time.Minute {
				delete(ipLimiters, ip)
			}
		}
		limitersMu.Unlock()
	}
}

// RateLimitMiddleware enforces per-IP rate limiting using a token bucket.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.GlobalConfig == nil || !config.GlobalConfig.RateLimitEnabled {
			c.Next()
			return
		}
		if !getLimiter(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}
		c.Next()
	}
}
