package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
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
		// Adjusted for HTMX and inline scripts/styles that may be needed
		csp := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://cdn.tailwindcss.com", // HTMX, DaisyUI, Prism.js
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com",                                   // DaisyUI and Prism styles
			"font-src 'self' https://fonts.gstatic.com",
			"img-src 'self' data: https:",
			"connect-src 'self'",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}
		w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))
		
		// Strict-Transport-Security (HSTS) - only on HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			// max-age=31536000 (1 year), includeSubDomains
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		next.ServeHTTP(w, r)
	})
}

// SecureRedirect redirects HTTP to HTTPS
func SecureRedirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request is already HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Skip redirect for health checks and local development
		if r.URL.Path == "/health" || r.Host == "localhost" || strings.HasPrefix(r.Host, "127.0.0.1") {
			next.ServeHTTP(w, r)
			return
		}
		
		// Redirect to HTTPS
		target := "https://" + r.Host + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}

// SecureCookies ensures cookies have secure flags
func SecureCookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap response writer to intercept Set-Cookie headers
		wrapped := &secureResponseWriter{
			ResponseWriter: w,
			request:        r,
		}
		
		next.ServeHTTP(wrapped, r)
	})
}

// secureResponseWriter wraps http.ResponseWriter to modify cookies
type secureResponseWriter struct {
	http.ResponseWriter
	request *http.Request
}

// WriteHeader intercepts and modifies Set-Cookie headers
func (w *secureResponseWriter) WriteHeader(statusCode int) {
	// Get all Set-Cookie headers
	cookies := w.Header()["Set-Cookie"]
	if len(cookies) > 0 {
		w.Header().Del("Set-Cookie")
		
		// Modify each cookie
		for _, cookie := range cookies {
			cookie = ensureSecureCookie(cookie, w.request)
			w.Header().Add("Set-Cookie", cookie)
		}
	}
	
	w.ResponseWriter.WriteHeader(statusCode)
}

// ensureSecureCookie adds security flags to a cookie string
func ensureSecureCookie(cookie string, r *http.Request) string {
	// Check if cookie already has flags
	hasSecure := strings.Contains(cookie, "Secure")
	hasHTTPOnly := strings.Contains(cookie, "HttpOnly")
	hasSameSite := strings.Contains(cookie, "SameSite")
	
	// Add Secure flag for HTTPS
	if !hasSecure && (r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https") {
		cookie += "; Secure"
	}
	
	// Add HttpOnly flag to prevent JavaScript access
	if !hasHTTPOnly {
		cookie += "; HttpOnly"
	}
	
	// Add SameSite flag to prevent CSRF
	if !hasSameSite {
		cookie += "; SameSite=Lax"
	}
	
	return cookie
}

// CORSHeaders adds CORS headers for API endpoints
func CORSHeaders(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to API endpoints
			if !strings.HasPrefix(r.URL.Path, "/api") {
				next.ServeHTTP(w, r)
				return
			}
			
			origin := r.Header.Get("Origin")
			allowed := false
			
			// Check if origin is allowed
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			}
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// CombineMiddleware combines multiple middleware into one
func CombineMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}