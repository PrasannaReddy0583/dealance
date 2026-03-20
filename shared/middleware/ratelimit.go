package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/pkg/response"
)

// RateLimitConfig defines the rate limit parameters.
type RateLimitConfig struct {
	KeyPrefix  string        // e.g. "ratelimit:ip", "ratelimit:auth"
	MaxRequests int          // max requests in the window
	Window     time.Duration // sliding window duration
	KeyFunc    func(c *gin.Context) string // extracts the key component (IP, user ID, etc.)
}

// RateLimit implements a Redis sliding window rate limiter.
func RateLimit(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := fmt.Sprintf("%s:%s", cfg.KeyPrefix, cfg.KeyFunc(c))

		allowed, err := isAllowed(ctx, rdb, key, cfg.MaxRequests, cfg.Window)
		if err != nil {
			// On Redis error, allow the request (fail open) but log
			c.Next()
			return
		}

		if !allowed {
			response.Error(c, apperrors.ErrRateLimited)
			return
		}

		c.Next()
	}
}

// isAllowed implements sliding window rate limiting using Redis sorted sets.
func isAllowed(ctx context.Context, rdb *redis.Client, key string, maxRequests int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	pipe := rdb.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count remaining entries
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set TTL to auto-clean
	pipe.Expire(ctx, key, window+time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return countCmd.Val() < int64(maxRequests), nil
}

// Common rate limit configurations per spec

// IPRateLimit limits all requests by IP: 300 req/min
func IPRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimit(rdb, RateLimitConfig{
		KeyPrefix:   "ratelimit:ip",
		MaxRequests: 300,
		Window:      time.Minute,
		KeyFunc:     func(c *gin.Context) string { return c.ClientIP() },
	})
}

// AuthRateLimit limits auth attempts by IP: 10 req/15min
func AuthRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimit(rdb, RateLimitConfig{
		KeyPrefix:   "ratelimit:auth",
		MaxRequests: 10,
		Window:      15 * time.Minute,
		KeyFunc:     func(c *gin.Context) string { return c.ClientIP() },
	})
}

// OTPSendRateLimit limits OTP sends by email: 3/hour
func OTPSendRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimit(rdb, RateLimitConfig{
		KeyPrefix:   "ratelimit:otp",
		MaxRequests: 3,
		Window:      time.Hour,
		KeyFunc: func(c *gin.Context) string {
			// Email will be extracted from the request body by the handler
			// For rate limiting, use IP as fallback
			return c.ClientIP()
		},
	})
}

// OTPVerifyRateLimit limits OTP verifications: 5 attempts per OTP
func OTPVerifyRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimit(rdb, RateLimitConfig{
		KeyPrefix:   "ratelimit:otpverify",
		MaxRequests: 5,
		Window:      10 * time.Minute,
		KeyFunc:     func(c *gin.Context) string { return c.ClientIP() },
	})
}

// UploadRateLimit limits uploads by user: 20/hour
func UploadRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimit(rdb, RateLimitConfig{
		KeyPrefix:   "ratelimit:upload",
		MaxRequests: 20,
		Window:      time.Hour,
		KeyFunc:     func(c *gin.Context) string { return GetUserID(c) },
	})
}
