package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/labstack/echo/v4"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu        sync.Mutex
	tokens    map[string]*tokenBucket
	rate      int           // tokens per interval
	interval  time.Duration // refill interval
	burst     int           // max tokens (burst capacity)
	cleanupAt time.Time
}

type tokenBucket struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per interval
// interval: time duration for rate limit window
// burst: maximum requests allowed in a burst
func NewRateLimiter(rate int, interval time.Duration, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:    make(map[string]*tokenBucket),
		rate:      rate,
		interval:  interval,
		burst:     burst,
		cleanupAt: time.Now().Add(time.Hour),
	}
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Cleanup old entries periodically
	now := time.Now()
	if now.After(rl.cleanupAt) {
		rl.cleanup(now)
		rl.cleanupAt = now.Add(time.Hour)
	}

	bucket, exists := rl.tokens[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:    rl.burst,
			lastCheck: now,
		}
		rl.tokens[key] = bucket
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(bucket.lastCheck)
	tokensToAdd := int(elapsed / rl.interval) * rl.rate
	bucket.tokens += tokensToAdd
	if bucket.tokens > rl.burst {
		bucket.tokens = rl.burst
	}
	bucket.lastCheck = now

	// Check if request is allowed
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// cleanup removes stale entries
func (rl *RateLimiter) cleanup(now time.Time) {
	staleThreshold := now.Add(-time.Hour)
	for key, bucket := range rl.tokens {
		if bucket.lastCheck.Before(staleThreshold) {
			delete(rl.tokens, key)
		}
	}
}

// RateLimitMiddleware creates a middleware that limits request rate
// Uses user ID from context if available, otherwise uses IP
func RateLimitMiddleware(limiter *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Use user ID if authenticated, otherwise use IP
			var key string
			if userID := c.Get("user_id"); userID != nil {
				key = "user:" + userID.(interface{ String() string }).String()
			} else {
				key = "ip:" + c.RealIP()
			}

			if !limiter.Allow(key) {
				return c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
					Error:   "rate_limit_exceeded",
					Message: "요청이 너무 많습니다. 잠시 후 다시 시도해주세요",
				})
			}

			return next(c)
		}
	}
}

// AIRateLimiter is a pre-configured rate limiter for AI endpoints
// Allows 10 requests per minute with a burst of 5
var AIRateLimiter = NewRateLimiter(10, time.Minute, 5)
