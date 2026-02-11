package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/microservices-go/shared/config"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(cfg.RequestsPerSecond),
		burst:    cfg.BurstSize,
	}
}

// getLimiter returns or creates a rate limiter for a key
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := rl.limiters[key]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rl.rps, rl.burst)
	rl.limiters[key] = limiter

	// Cleanup old limiters periodically
	go rl.cleanupLimiter(key, limiter)

	return limiter
}

// cleanupLimiter removes limiter after it has been inactive
func (rl *RateLimiter) cleanupLimiter(key string, limiter *rate.Limiter) {
	time.Sleep(10 * time.Minute)
	rl.mu.Lock()
	delete(rl.limiters, key)
	rl.mu.Unlock()
}

// Middleware returns HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as key (can be customized)
		key := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			key = forwarded
		}

		limiter := rl.getLimiter(key)
		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", string(rune(rl.rps)))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests, please try again later"}}`))
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

		limiter := rl.getLimiter(key)
		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests, please try again later"}}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}