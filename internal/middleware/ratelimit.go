package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
)

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

local bucket = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(bucket[1])
local ts = tonumber(bucket[2])

if tokens == nil then
  tokens = capacity
  ts = now
end

local delta = math.max(0, now - ts)
tokens = math.min(capacity, tokens + delta * refill)

local allowed = 0
if tokens >= 1 then
  allowed = 1
  tokens = tokens - 1
end

redis.call("HMSET", key, "tokens", tokens, "ts", now)
redis.call("EXPIRE", key, ttl)

return { allowed, math.floor(tokens) }
`)

type RateLimitConfig struct {
	Prefix string
	Limit  int64
	Window time.Duration
	KeyFn  func(*gin.Context) string
}

func RateLimit(redisClient *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	if cfg.Limit <= 0 {
		cfg.Limit = 60
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Prefix == "" {
		cfg.Prefix = "global"
	}
	if cfg.KeyFn == nil {
		cfg.KeyFn = func(c *gin.Context) string { return c.ClientIP() }
	}

	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		key := "ratelimit:" + cfg.Prefix + ":" + cfg.KeyFn(c)
		allowed, remaining, err := allowRequest(c.Request.Context(), redisClient, key, cfg.Limit, cfg.Window)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(cfg.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		if !allowed {
			app.Error(c, http.StatusTooManyRequests, errcode.RateLimited, "rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

func allowRequest(ctx context.Context, redisClient *redis.Client, key string, limit int64, window time.Duration) (bool, int64, error) {
	refillPerSecond := float64(limit) / window.Seconds()
	if refillPerSecond <= 0 {
		refillPerSecond = 1
	}
	result, err := tokenBucketScript.Run(
		ctx,
		redisClient,
		[]string{key},
		limit,
		refillPerSecond,
		time.Now().Unix(),
		int64(window.Seconds()*2),
	).Slice()
	if err != nil {
		return true, limit, err
	}

	allowed, _ := result[0].(int64)
	remaining, _ := result[1].(int64)
	return allowed == 1, remaining, nil
}
