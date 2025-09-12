package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// PathBasedRateLimiter applies different rate limits based on request path
type PathBasedRateLimiter struct {
	limiters map[string]*RateLimiter
	configs  map[string]rateLimitConfig
	mu       sync.RWMutex
}

type rateLimitConfig struct {
	rate   int
	window time.Duration
	opts   []Option
}

// NewPathBasedRateLimiter creates a rate limiter with path-specific limits
func NewPathBasedRateLimiter() *PathBasedRateLimiter {
	return &PathBasedRateLimiter{
		limiters: make(map[string]*RateLimiter),
		configs: map[string]rateLimitConfig{
			// Authentication endpoints - strict limits with fixed window
			"auth":    {5, time.Minute, ForAuth()},
			"signin":  {5, time.Minute, ForAuth()},
			"signup":  {3, time.Hour, ForAuth()},
			
			// API endpoints - moderate limits with token bucket for bursts
			"api":     {60, time.Minute, ForAPI()},
			"webhook": {100, time.Minute, ForWebhook()},
			"stripe":  {100, time.Minute, ForWebhook()},
			
			// Admin endpoints - relaxed limits with sliding window
			"admin":   {120, time.Minute, []Option{WithStrategy(SlidingWindow)}},
			
			// General endpoints - default limits
			"default": {100, time.Minute, nil},
		},
	}
}

// Handle returns the HTTP handler that applies path-based rate limiting
func (p *PathBasedRateLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the appropriate limiter based on path
		limiter := p.getLimiterForPath(r.URL.Path)
		
		// Apply the rate limit
		limiter.Handle(next).ServeHTTP(w, r)
	})
}

// getLimiterForPath selects the appropriate rate limiter based on the request path
func (p *PathBasedRateLimiter) getLimiterForPath(path string) *RateLimiter {
	// Clean the path
	path = strings.TrimPrefix(path, "/")
	
	// Determine which config to use
	var configKey string
	switch {
	case strings.HasPrefix(path, "auth/") || path == "auth":
		configKey = "auth"
	case strings.HasPrefix(path, "signin") || strings.HasPrefix(path, "_signin"):
		configKey = "signin"
	case strings.HasPrefix(path, "signup") || strings.HasPrefix(path, "_signup"):
		configKey = "signup"
	case strings.HasPrefix(path, "api/"):
		configKey = "api"
	case strings.HasPrefix(path, "webhook/") || strings.HasPrefix(path, "webhooks/"):
		configKey = "webhook"
	case strings.HasPrefix(path, "stripe/webhook"):
		configKey = "stripe"
	case strings.HasPrefix(path, "admin/"):
		configKey = "admin"
	default:
		configKey = "default"
	}
	
	// Get or create the limiter
	return p.getOrCreateLimiter(configKey)
}

// getOrCreateLimiter returns an existing limiter or creates a new one
func (p *PathBasedRateLimiter) getOrCreateLimiter(key string) *RateLimiter {
	// Fast path: check with read lock
	p.mu.RLock()
	if limiter, ok := p.limiters[key]; ok {
		p.mu.RUnlock()
		return limiter
	}
	p.mu.RUnlock()
	
	// Slow path: create with write lock
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if limiter, ok := p.limiters[key]; ok {
		return limiter
	}
	
	// Get config or use default
	config, ok := p.configs[key]
	if !ok {
		config = p.configs["default"]
	}
	
	// Create new limiter
	limiter := New(key, config.rate, config.window, config.opts...)
	p.limiters[key] = limiter
	return limiter
}

// Stop stops all rate limiters and their cleanup goroutines
func (p *PathBasedRateLimiter) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for _, limiter := range p.limiters {
		limiter.Stop()
	}
	p.limiters = make(map[string]*RateLimiter)
}