package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements application.Middleware with flexible rate limiting
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex

	// Configuration
	name   string
	rate   int           // requests per window
	window time.Duration // time window

	// Options
	strategy        Strategy
	keyFunc         KeyFunc
	errorHandler    ErrorHandler
	skipFunc        SkipFunc
	cleanupInterval time.Duration
	stopCh          chan struct{}
}

// visitor tracks rate limit state for a single visitor
type visitor struct {
	lastSeen time.Time
	count    int
	tokens   float64 // For token bucket strategy
}

// Strategy defines how rate limiting is applied
type Strategy int

const (
	// FixedWindow resets the counter at fixed intervals
	FixedWindow Strategy = iota
	// SlidingWindow uses a sliding time window
	SlidingWindow
	// TokenBucket allows bursts up to the bucket size
	TokenBucket
)

// KeyFunc generates a rate limit key from the request
type KeyFunc func(*http.Request) string

// ErrorHandler handles rate limit errors
type ErrorHandler func(http.ResponseWriter, *http.Request)

// SkipFunc determines if a request should skip rate limiting
type SkipFunc func(*http.Request) bool

// Option configures a RateLimiter
type Option func(*RateLimiter)

// New creates a new rate limiter middleware
func New(name string, rate int, window time.Duration, opts ...Option) *RateLimiter {
	rl := &RateLimiter{
		visitors: map[string]*visitor{},
		name:     name,
		rate:     rate,
		window:   window,
		// Defaults
		strategy:        FixedWindow,
		keyFunc:         IPKeyFunc,
		errorHandler:    DefaultErrorHandler,
		skipFunc:        nil,
		cleanupInterval: window * 2,
		stopCh:          make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(rl)
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// Name returns the middleware name
func (rl *RateLimiter) Name() string {
	return rl.name
}

// Handle returns the HTTP handler that enforces rate limiting
func (rl *RateLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this request should skip rate limiting
		if rl.skipFunc != nil && rl.skipFunc(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Get rate limit key
		key := rl.keyFunc(r)

		// Check if visitor is rate limited
		if !rl.allow(key) {
			rl.errorHandler(w, r)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// allow checks if a visitor should be allowed to proceed
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	switch rl.strategy {
	case TokenBucket:
		return rl.allowTokenBucket(key)
	case SlidingWindow:
		return rl.allowSlidingWindow(key)
	default: // FixedWindow
		return rl.allowFixedWindow(key)
	}
}

// allowFixedWindow implements fixed window rate limiting
func (rl *RateLimiter) allowFixedWindow(key string) bool {
	v, exists := rl.visitors[key]
	if !exists {
		// First visit
		rl.visitors[key] = &visitor{
			lastSeen: time.Now(),
			count:    1,
		}
		return true
	}

	// Check if window has expired
	if time.Since(v.lastSeen) > rl.window {
		// Reset the visitor
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	// Within window, check count
	if v.count >= rl.rate {
		return false
	}

	// Increment count
	v.count++
	v.lastSeen = time.Now()
	return true
}

// allowSlidingWindow implements sliding window rate limiting
func (rl *RateLimiter) allowSlidingWindow(key string) bool {
	now := time.Now()
	v, exists := rl.visitors[key]

	if !exists {
		rl.visitors[key] = &visitor{
			lastSeen: now,
			count:    1,
		}
		return true
	}

	// Calculate weighted count based on time elapsed
	elapsed := now.Sub(v.lastSeen)
	weight := float64(rl.window-elapsed) / float64(rl.window)
	weightedCount := float64(v.count) * weight

	if weightedCount+1 > float64(rl.rate) {
		return false
	}

	v.count = int(weightedCount) + 1
	v.lastSeen = now
	return true
}

// allowTokenBucket implements token bucket rate limiting
func (rl *RateLimiter) allowTokenBucket(key string) bool {
	now := time.Now()
	v, exists := rl.visitors[key]

	if !exists {
		rl.visitors[key] = &visitor{
			lastSeen: now,
			tokens:   float64(rl.rate) - 1,
		}
		return true
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(v.lastSeen)
	tokensToAdd := elapsed.Seconds() * float64(rl.rate) / rl.window.Seconds()
	v.tokens = min(float64(rl.rate), v.tokens+tokensToAdd)

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	v.lastSeen = now
	return true
}

// cleanupVisitors removes old entries from the visitors map
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for key, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.window*2 {
					delete(rl.visitors, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCh:
			return
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Key functions

// IPKeyFunc rate limits by IP address (default)
func IPKeyFunc(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if multiple are present
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return forwarded
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// UserKeyFunc rate limits by authenticated user ID
func UserKeyFunc(userIDExtractor func(*http.Request) string) KeyFunc {
	return func(r *http.Request) string {
		if userID := userIDExtractor(r); userID != "" {
			return "user:" + userID
		}
		// Fall back to IP if no user
		return IPKeyFunc(r)
	}
}

// PathKeyFunc rate limits by path and IP
func PathKeyFunc(r *http.Request) string {
	return r.URL.Path + ":" + IPKeyFunc(r)
}

// APIKeyFunc rate limits by API key
func APIKeyFunc(r *http.Request) string {
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return "api:" + apiKey
	}
	return IPKeyFunc(r)
}

// DefaultErrorHandler sends a standard rate limit error response
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Retry-After", "60")
	w.Header().Set("X-RateLimit-Limit", "60")
	w.Header().Set("X-RateLimit-Remaining", "0")
	http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
}
