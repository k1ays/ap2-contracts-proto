package redis

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	max    int
	window time.Duration
}

func NewRateLimiter(client *redis.Client, max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		max:    max,
		window: window,
	}
}

func NewRateLimiterFromAddr(addr string, max int, window time.Duration) (*RateLimiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping for rate limiter: %w", err)
	}

	return &RateLimiter{
		client: client,
		max:    max,
		window: window,
	}, nil
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx := c.Request.Context()

		count, err := rl.client.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rl.client.Expire(ctx, key, rl.window)
		}

		if count > int64(rl.max) {
			ttl, _ := rl.client.TTL(ctx, key).Result()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": fmt.Sprintf("%.0fs", ttl.Seconds()),
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(rl.max)-count))

		c.Next()
	}
}
