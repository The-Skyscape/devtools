package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

// CSRFProtection provides CSRF protection for the application
type CSRFProtection struct {
	tokens map[string]*csrfToken
	mu     sync.RWMutex
	ttl    time.Duration
}

// csrfToken represents a CSRF token with expiration
type csrfToken struct {
	token   string
	expires time.Time
}

// NewCSRFProtection creates a new CSRF protection middleware
func NewCSRFProtection(ttl time.Duration) *CSRFProtection {
	csrf := &CSRFProtection{
		tokens: make(map[string]*csrfToken),
		ttl:    ttl,
	}
	
	// Start cleanup goroutine
	go csrf.cleanupTokens()
	
	return csrf
}

// GenerateToken generates a new CSRF token for a session
func (c *CSRFProtection) GenerateToken(sessionID string) string {
	// Generate random token
	b := make([]byte, 32)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)
	
	// Store token with expiration
	c.mu.Lock()
	c.tokens[sessionID] = &csrfToken{
		token:   token,
		expires: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
	
	return token
}

// ValidateToken validates a CSRF token
func (c *CSRFProtection) ValidateToken(sessionID, token string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	stored, exists := c.tokens[sessionID]
	if !exists {
		return false
	}
	
	// Check if token has expired
	if time.Now().After(stored.expires) {
		return false
	}
	
	// Constant time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(stored.token), []byte(token)) == 1
}

// Protect wraps a handler with CSRF protection
func (c *CSRFProtection) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Skip CSRF for API endpoints (they should use API keys)
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Skip CSRF for webhook endpoints (they have their own validation)
		if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/webhook" {
			next.ServeHTTP(w, r)
			return
		}
		
		// For HTMX requests, we rely on same-origin policy
		// HTMX requests include the HX-Request header
		if r.Header.Get("HX-Request") == "true" {
			// Verify same origin
			origin := r.Header.Get("Origin")
			if origin != "" && !isSameOrigin(r, origin) {
				http.Error(w, "CSRF validation failed: Invalid origin", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		
		// For form submissions, check CSRF token
		sessionID := getSessionID(r)
		if sessionID == "" {
			http.Error(w, "CSRF validation failed: No session", http.StatusForbidden)
			return
		}
		
		// Get token from form or header
		token := r.FormValue("csrf_token")
		if token == "" {
			token = r.Header.Get("X-CSRF-Token")
		}
		
		if !c.ValidateToken(sessionID, token) {
			http.Error(w, "CSRF validation failed: Invalid token", http.StatusForbidden)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// InjectToken injects a CSRF token into the response for templates
func (c *CSRFProtection) InjectToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only inject for GET requests that might render forms
		if r.Method == "GET" {
			sessionID := getSessionID(r)
			if sessionID != "" {
				token := c.GenerateToken(sessionID)
				// Add token to request context for templates
				r.Header.Set("X-CSRF-Token", token)
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// cleanupTokens removes expired tokens
func (c *CSRFProtection) cleanupTokens() {
	for {
		time.Sleep(5 * time.Minute)
		
		c.mu.Lock()
		now := time.Now()
		for session, token := range c.tokens {
			if now.After(token.expires) {
				delete(c.tokens, session)
			}
		}
		c.mu.Unlock()
	}
}

// getSessionID extracts session ID from cookie
func getSessionID(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		// Try auth cookie as fallback
		cookie, err = r.Cookie("auth")
		if err != nil {
			return ""
		}
	}
	return cookie.Value
}

// isSameOrigin checks if the origin matches the request host
func isSameOrigin(r *http.Request, origin string) bool {
	// Parse origin
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	
	expected := scheme + "://" + r.Host
	return origin == expected
}

// CSRFTokenField returns an HTML input field with CSRF token
func CSRFTokenField(r *http.Request) string {
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		return ""
	}
	return `<input type="hidden" name="csrf_token" value="` + token + `">`
}

// CSRFMetaTag returns an HTML meta tag with CSRF token for JavaScript
func CSRFMetaTag(r *http.Request) string {
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		return ""
	}
	return `<meta name="csrf-token" content="` + token + `">`
}