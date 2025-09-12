package middleware

import (
	"net/http"
	"time"
)

// Options for RateLimiter

// WithStrategy sets the rate limiting strategy
func WithStrategy(strategy Strategy) Option {
	return func(rl *RateLimiter) {
		rl.strategy = strategy
	}
}

// WithKeyFunc sets the function used to generate rate limit keys
func WithKeyFunc(keyFunc KeyFunc) Option {
	return func(rl *RateLimiter) {
		rl.keyFunc = keyFunc
	}
}

// WithErrorHandler sets a custom error handler for rate limit errors
func WithErrorHandler(handler ErrorHandler) Option {
	return func(rl *RateLimiter) {
		rl.errorHandler = handler
	}
}

// WithSkipFunc sets a function to determine if requests should skip rate limiting
func WithSkipFunc(skipFunc SkipFunc) Option {
	return func(rl *RateLimiter) {
		rl.skipFunc = skipFunc
	}
}

// WithCleanupInterval sets how often to clean up old visitor entries
func WithCleanupInterval(interval time.Duration) Option {
	return func(rl *RateLimiter) {
		rl.cleanupInterval = interval
	}
}

// Common skip functions

// SkipPaths returns a skip function that skips specific paths
func SkipPaths(paths ...string) SkipFunc {
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}
	return func(r *http.Request) bool {
		return pathMap[r.URL.Path]
	}
}

// SkipMethods returns a skip function that skips specific HTTP methods
func SkipMethods(methods ...string) SkipFunc {
	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m] = true
	}
	return func(r *http.Request) bool {
		return methodMap[r.Method]
	}
}

// SkipAny combines multiple skip functions with OR logic
func SkipAny(funcs ...SkipFunc) SkipFunc {
	return func(r *http.Request) bool {
		for _, f := range funcs {
			if f(r) {
				return true
			}
		}
		return false
	}
}

// Configuration presets

// ForAPI creates options suitable for API endpoints
func ForAPI() []Option {
	return []Option{
		WithStrategy(TokenBucket),
		WithKeyFunc(APIKeyFunc),
		WithSkipFunc(SkipMethods("OPTIONS")),
	}
}

// ForAuth creates options suitable for authentication endpoints  
func ForAuth() []Option {
	return []Option{
		WithStrategy(FixedWindow),
		WithKeyFunc(IPKeyFunc),
		WithErrorHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too many authentication attempts. Please try again later.", http.StatusTooManyRequests)
		}),
	}
}

// ForUser creates options for per-user rate limiting
func ForUser(extractUserID func(*http.Request) string) []Option {
	return []Option{
		WithStrategy(SlidingWindow),
		WithKeyFunc(UserKeyFunc(extractUserID)),
	}
}

// ForWebhook creates options suitable for webhook endpoints
func ForWebhook() []Option {
	return []Option{
		WithStrategy(TokenBucket),
		WithKeyFunc(PathKeyFunc),
		WithSkipFunc(SkipMethods("HEAD")),
	}
}