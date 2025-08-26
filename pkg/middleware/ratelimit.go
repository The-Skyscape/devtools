package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple rate limiting middleware
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

// visitor tracks rate limit state for a single visitor
type visitor struct {
	lastSeen time.Time
	count    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	
	// Start cleanup goroutine
	go rl.cleanupVisitors()
	
	return rl
}

// Limit returns a middleware that enforces rate limiting
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get visitor identifier (IP address)
		ip := getIP(r)
		
		// Check if visitor is rate limited
		if !rl.allow(ip) {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}
		
		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// LimitFunc wraps a handler function with rate limiting
func (rl *RateLimiter) LimitFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get visitor identifier (IP address)
		ip := getIP(r)
		
		// Check if visitor is rate limited
		if !rl.allow(ip) {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}
		
		// Call the handler
		handler(w, r)
	}
}

// allow checks if a visitor should be allowed to proceed
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	v, exists := rl.visitors[ip]
	if !exists {
		// First visit
		rl.visitors[ip] = &visitor{
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

// cleanupVisitors removes old entries from the visitors map
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getIP extracts the real IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	
	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// CreateDefaultRateLimiters creates rate limiters for common endpoints
func CreateDefaultRateLimiters() map[string]*RateLimiter {
	return map[string]*RateLimiter{
		// Authentication endpoints - strict limits
		"auth":     NewRateLimiter(5, time.Minute),    // 5 requests per minute
		"signin":   NewRateLimiter(5, time.Minute),    // 5 login attempts per minute
		"signup":   NewRateLimiter(3, time.Hour),      // 3 signups per hour
		
		// API endpoints - moderate limits
		"api":      NewRateLimiter(60, time.Minute),   // 60 requests per minute
		"webhook":  NewRateLimiter(100, time.Minute),  // 100 webhook calls per minute
		
		// Admin endpoints - relaxed limits
		"admin":    NewRateLimiter(120, time.Minute),  // 120 requests per minute
		
		// General endpoints - default limits
		"default":  NewRateLimiter(100, time.Minute),  // 100 requests per minute
	}
}

// GetRateLimiter returns the appropriate rate limiter for a path and method
func GetRateLimiter(limiters map[string]*RateLimiter, path string, method string) *RateLimiter {
	// Check specific paths - only apply strict limits to POST requests
	if method == "POST" {
		if path == "/signin" || path == "/auth/signin" || path == "/_auth/signin" {
			return limiters["signin"]
		}
		if path == "/signup" || path == "/auth/signup" || path == "/_auth/signup" {
			return limiters["signup"]
		}
	}
	
	// Check path prefixes
	if len(path) >= 6 && path[:6] == "/_auth" {
		// Internal auth endpoints get auth rate limit
		return limiters["auth"]
	}
	if len(path) >= 5 && path[:5] == "/auth" {
		return limiters["auth"]
	}
	if len(path) >= 4 && path[:4] == "/api" {
		return limiters["api"]
	}
	if len(path) >= 6 && path[:6] == "/admin" {
		return limiters["admin"]
	}
	if len(path) >= 8 && path[:8] == "/webhook" {
		return limiters["webhook"]
	}
	
	// Default rate limiter
	return limiters["default"]
}

// ApplyRateLimit applies rate limiting to a handler based on the request path and method
func ApplyRateLimit(limiters map[string]*RateLimiter, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := GetRateLimiter(limiters, r.URL.Path, r.Method)
		limiter.Limit(handler).ServeHTTP(w, r)
	})
}