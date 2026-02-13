package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/microservices-go/shared/config"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/redis"
)

// RateLimiter implements fixed window rate limiting using Redis
type RateLimiter struct {
	redis  *redis.Client
	config *config.RateLimitConfig
	prefix string
}

// NewRateLimiter creates a new Redis-based rate limiter
func NewRateLimiter(redisClient *redis.Client, cfg *config.RateLimitConfig, prefix string) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		config: cfg,
		prefix: prefix,
	}
}

// Allow checks if the request should be allowed
func (rl *RateLimiter) Allow(key string) (bool, int, int, int) {
	window := time.Duration(rl.config.WindowSeconds) * time.Second
	limit := rl.config.RequestsPerMinute

	fullKey := fmt.Sprintf("ratelimit:%s:%s", rl.prefix, key)

	count, err := rl.redis.Increment(fullKey, window)
	if err != nil {
		log := logger.New("rate-limiter")
		log.WithError(err).WithField("key", fullKey).Error("Failed to increment rate limit counter")
		// Allow request on Redis error to avoid blocking legitimate traffic
		return true, 0, limit, 0
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	allowed := count <= int64(limit)

	return allowed, int(count), limit, remaining
}

// Middleware returns HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as key
		key := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			key = forwarded
		}

		allowed, current, limit, remaining := rl.Allow(key)

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests, please try again later"}}`))

			log := logger.New("rate-limiter")
			log.WithField("key", key).
				WithField("current", current).
				WithField("limit", limit).
				Warn("Rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// PerUserMiddleware returns rate limiting middleware per user
func (rl *RateLimiter) PerUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := "anonymous"

		if claims, ok := GetUserFromContext(r.Context()); ok {
			key = claims.UserID
		}

		allowed, current, limit, remaining := rl.Allow(key)

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests, please try again later"}}`))

			log := logger.New("rate-limiter")
			log.WithField("user", key).
				WithField("current", current).
				WithField("limit", limit).
				Warn("Rate limit exceeded for user")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Close closes the rate limiter and its Redis connection
func (rl *RateLimiter) Close() error {
	return rl.redis.Close()
}
