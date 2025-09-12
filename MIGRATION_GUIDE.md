# Middleware Migration Guide

This guide helps you migrate from the old `pkg/middleware` package to the new `pkg/application/middleware` structure.

## Key Changes

1. **Import Path**: `github.com/The-Skyscape/devtools/pkg/middleware` → `github.com/The-Skyscape/devtools/pkg/application/middleware`
2. **Interface-Based**: All middleware implements a single-method `Middleware` interface with `Handle(http.Handler) http.Handler`
3. **Simplified API**: One interface, one option (`WithMiddleware`), compose as needed

## Migration Examples

### Old Way (pkg/middleware)

```go
import (
    "github.com/The-Skyscape/devtools/pkg/middleware"
)

func main() {
    // Old rate limiter
    limiter := middleware.NewRateLimiter(100, time.Minute)
    http.Handle("/api", limiter.Limit(handler))
    
    // Old middleware functions
    http.Handle("/", middleware.SecurityHeaders(
        middleware.LogRequests(
            middleware.Recovery(handler),
        ),
    ))
    
    // Old rate limiter config
    limiters := middleware.CreateDefaultRateLimiters()
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        limiter := middleware.GetRateLimiter(limiters, r.URL.Path, r.Method)
        limiter.Limit(handler).ServeHTTP(w, r)
    })
}
```

### New Way (application/middleware)

```go
import (
    "github.com/The-Skyscape/devtools/pkg/application"
    "github.com/The-Skyscape/devtools/pkg/application/middleware"
)

func main() {
    // Option 1: Use standard middleware (recommended for most apps)
    application.Serve(views,
        application.WithStandardMiddleware(),  // Recovery, logging, security, compression, timeout
    )
    
    // Option 2: Compose your own middleware chain
    application.Serve(views,
        application.WithMiddleware(middleware.NewChain(
            middleware.NewRecovery(true),
            middleware.NewLogger(middleware.CommonLoggerFormat),
            middleware.New("api", 100, time.Minute),  // Rate limiter
        )),
    )
    
    // Option 3: Use middleware directly with http.Handler
    limiter := middleware.New("api-limiter", 100, time.Minute)
    http.Handle("/api", limiter.Handle(handler))
}
```

## Specific Migrations

### Rate Limiting

```go
// OLD
limiter := middleware.NewRateLimiter(100, time.Minute)
http.Handle("/", limiter.Limit(handler))

// NEW - Direct usage
limiter := middleware.New("api", 100, time.Minute)
http.Handle("/", limiter.Handle(handler))

// NEW - With options
limiter := middleware.New("api", 60, time.Minute,
    middleware.WithStrategy(middleware.TokenBucket),
    middleware.WithKeyFunc(middleware.APIKeyFunc),
)

// NEW - In application setup
application.Serve(views,
    application.WithMiddleware(limiter),
)
```

### Custom Rate Limiting

```go
// OLD
config := middleware.RateLimiterConfig{
    Rate:     100,
    Window:   time.Minute,
    Strategy: "token",
    KeyBy:    "user",
}
limiter := middleware.CreateRateLimiter(config)

// NEW
limiter := middleware.New("custom", 100, time.Minute,
    middleware.WithStrategy(middleware.TokenBucket),
    middleware.WithKeyFunc(middleware.UserKeyFunc(extractUserID)),
)
```

### Security Headers

```go
// OLD
http.Handle("/", middleware.SecurityHeaders(handler))

// NEW - Direct usage
security := middleware.NewSecurityHeaders([]string{})  // Uses default CSP
http.Handle("/", security.Handle(handler))

// NEW - Custom CSP
security := middleware.NewSecurityHeaders([]string{
    "default-src 'self'",
    "script-src 'self' 'unsafe-inline'",
})
```

### Logging

```go
// OLD
http.Handle("/", middleware.LogRequests(handler))

// NEW - Direct usage
logger := middleware.NewLogger(middleware.CommonLoggerFormat)
http.Handle("/", logger.Handle(handler))
```

### CORS

```go
// OLD
cors := middleware.NewCORS([]string{"*"})
http.Handle("/", cors.Handler(handler))

// NEW - Direct usage
cors := middleware.NewCORS([]string{"*"})
http.Handle("/", cors.Handle(handler))
```

## Application-Level Integration

### Workspace/Website/Workbench Migration

Replace custom rate limiters with framework middleware:

```go
// OLD - Custom rate limiter in each app
type RateLimiter struct {
    visitors map[string]*visitor
    // ... custom implementation
}

func NewRateLimiter() *RateLimiter {
    // ... custom code
}

// NEW - Use framework middleware
import "github.com/The-Skyscape/devtools/pkg/application"

func main() {
    application.Serve(views,
        // Use standard middleware for most apps
        application.WithStandardMiddleware(),
        
        // Or compose your own
        application.WithMiddleware(middleware.NewChain(
            middleware.NewRecovery(true),
            middleware.New("api", 100, time.Minute),
        )),
    )
}
```

## Testing with Middleware

```go
import (
    "testing"
    "net/http/httptest"
    "github.com/The-Skyscape/devtools/pkg/application/middleware"
)

func TestWithRateLimit(t *testing.T) {
    // Create a test limiter
    limiter := middleware.New("test", 2, time.Second)
    
    handler := limiter.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    
    // First two requests succeed
    for i := 0; i < 2; i++ {
        req := httptest.NewRequest("GET", "/", nil)
        rec := httptest.NewRecorder()
        handler.ServeHTTP(rec, req)
        
        if rec.Code != http.StatusOK {
            t.Errorf("Expected 200, got %d", rec.Code)
        }
    }
    
    // Third request is rate limited
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    
    if rec.Code != http.StatusTooManyRequests {
        t.Errorf("Expected 429, got %d", rec.Code)
    }
}
```

## Benefits of Migration

1. **Reduced Code**: Remove ~370 lines of duplicate rate limiting code
2. **Consistency**: All apps use the same middleware
3. **Testability**: Interface-based design makes testing easier
4. **Flexibility**: Functional options pattern for configuration
5. **Maintenance**: Updates to middleware benefit all applications

## Quick Reference

| Old Function | New Equivalent |
|-------------|----------------|
| `middleware.NewRateLimiter()` | `middleware.New("name", rate, window)` |
| `limiter.Limit()` | `limiter.Handle()` |
| `middleware.SecurityHeaders()` | `middleware.NewSecurityHeaders([]string{})` |
| `middleware.LogRequests()` | `middleware.NewLogger(format)` |
| `middleware.Recovery()` | `middleware.NewRecovery(stackTrace)` |
| `middleware.Compress()` | `middleware.NewCompression(level)` |
| Multiple middleware setup | `application.WithStandardMiddleware()` |