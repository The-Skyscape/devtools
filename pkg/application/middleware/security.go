package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders provides security headers middleware
type SecurityHeaders struct {
	csp []string
}

// NewSecurityHeaders creates a new security headers middleware
func NewSecurityHeaders(csp []string) *SecurityHeaders {
	if len(csp) == 0 {
		csp = []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://cdn.tailwindcss.com", // HTMX, DaisyUI, Prism.js
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com",                                 // DaisyUI and Prism styles
			"font-src 'self' https://fonts.gstatic.com",
			"img-src 'self' data: https:",
			"connect-src 'self'",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}
	}
	return &SecurityHeaders{csp}
}

// Handle returns the HTTP handler that adds security headers
func (s *SecurityHeaders) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Content-Type-Options prevents MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options prevents clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// X-XSS-Protection enables XSS filter in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy controls referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (formerly Feature-Policy) restricts browser features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content-Security-Policy helps prevent XSS and other attacks
		w.Header().Set("Content-Security-Policy", strings.Join(s.csp, "; "))

		// Strict-Transport-Security (HSTS) - only on HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			// max-age=31536000 (1 year), includeSubDomains
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// WithCSP sets custom Content-Security-Policy directives
func (s *SecurityHeaders) WithCSP(directives []string) *SecurityHeaders {
	s.csp = directives
	return s
}
