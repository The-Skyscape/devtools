package middleware

import (
	"net/http"
)

// CORS handles Cross-Origin Resource Sharing
type CORS struct {
	allowedOrigins map[string]bool
	allowAll       bool
}

// NewCORS creates a new CORS middleware
func NewCORS(allowedOrigins []string) *CORS {
	cors := &CORS{
		allowedOrigins: map[string]bool{},
	}

	for _, origin := range allowedOrigins {
		if origin == "*" {
			cors.allowAll = true
			break
		}
		cors.allowedOrigins[origin] = true
	}

	return cors
}

// Handle returns the CORS middleware handler
func (c *CORS) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if c.allowAll || c.allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AllowOrigin checks if an origin is allowed
func (c *CORS) AllowOrigin(origin string) bool {
	return c.allowAll || c.allowedOrigins[origin]
}
