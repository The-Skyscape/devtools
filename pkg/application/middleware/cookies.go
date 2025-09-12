package middleware

import (
	"net/http"
	"strings"
)

// SecureCookies ensures all cookies have secure flags set
type SecureCookies struct{}

// NewSecureCookies creates a new secure cookies middleware
func NewSecureCookies() *SecureCookies {
	return &SecureCookies{}
}

// Handle returns the HTTP handler that secures cookies
func (s *SecureCookies) Handle(next http.Handler) http.Handler {
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
	secured bool // Track if cookies have been secured
}

// WriteHeader intercepts the status code and modifies cookies before writing
func (w *secureResponseWriter) WriteHeader(code int) {
	if !w.secured {
		w.secureCookies()
		w.secured = true
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write ensures cookies are secured before first write
func (w *secureResponseWriter) Write(b []byte) (int, error) {
	if !w.secured {
		w.secureCookies()
		w.secured = true
	}
	return w.ResponseWriter.Write(b)
}

// secureCookies modifies all Set-Cookie headers to add security flags
func (w *secureResponseWriter) secureCookies() {
	cookies := w.Header()["Set-Cookie"]
	if len(cookies) == 0 {
		return
	}
	
	// Clear existing cookies
	w.Header().Del("Set-Cookie")
	
	// Re-add with security flags
	for _, cookie := range cookies {
		secured := cookie
		
		// Add HttpOnly if not present
		if !strings.Contains(strings.ToLower(cookie), "httponly") {
			secured += "; HttpOnly"
		}
		
		// Add SameSite if not present
		if !strings.Contains(strings.ToLower(cookie), "samesite") {
			secured += "; SameSite=Lax"
		}
		
		// Add Secure flag if on HTTPS
		if (w.request.TLS != nil || w.request.Header.Get("X-Forwarded-Proto") == "https") && 
		   !strings.Contains(strings.ToLower(cookie), "secure") {
			secured += "; Secure"
		}
		
		w.Header().Add("Set-Cookie", secured)
	}
}

// Flush implements http.Flusher
func (w *secureResponseWriter) Flush() {
	if !w.secured {
		w.secureCookies()
		w.secured = true
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}