package authentication

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
)

func setupTestApp() *application.App {
	// Create a minimal embedded FS for testing
	views := http.Dir(".")
	app := application.New(views)

	// Note: In newer versions, Render is not directly assignable
	// Tests should use the actual rendering behavior

	return app
}

func TestRequireAuth(t *testing.T) {
	auth := New("test-secret")
	cookieName := "test-session"
	
	// Mock user function
	mockGetUser := func(userID string) (any, error) {
		if userID == "valid-user" {
			return map[string]any{"id": userID, "name": "Test User"}, nil
		}
		return nil, errors.New("user not found")
	}
	
	// Create valid token
	validToken, err := auth.GenerateSessionToken("valid-user", time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	// Create token for invalid user
	invalidUserToken, err := auth.GenerateSessionToken("invalid-user", time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	// Create expired token
	expiredToken, err := auth.GenerateSessionToken("valid-user", -time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	accessCheck := RequireAuth(auth, cookieName, mockGetUser)
	
	tests := []struct {
		name        string
		setupCookie func(*http.Request)
		expected    bool
	}{
		{
			name: "valid token with valid user",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: cookieName, Value: validToken})
			},
			expected: true,
		},
		{
			name: "valid token with invalid user",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: cookieName, Value: invalidUserToken})
			},
			expected: false,
		},
		{
			name: "expired token",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: cookieName, Value: expiredToken})
			},
			expected: false,
		},
		{
			name: "no cookie",
			setupCookie: func(r *http.Request) {
				// No cookie added
			},
			expected: false,
		},
		{
			name: "invalid token",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: cookieName, Value: "invalid-token"})
			},
			expected: false,
		},
		{
			name: "wrong cookie name",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: "wrong-cookie", Value: validToken})
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			
			tt.setupCookie(req)
			
			result := accessCheck(app, w, req)
			if tt.expected != result { t.Errorf("Expected %v, got %v", tt.expected, result) }
		})
	}
}

func TestRequireAuthWithoutUserFunction(t *testing.T) {
	auth := New("test-secret")
	cookieName := "test-session"
	
	// Create valid token
	validToken, err := auth.GenerateSessionToken("any-user", time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	// Test without user function (nil)
	accessCheck := RequireAuth(auth, cookieName, nil)
	
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: validToken})
	w := httptest.NewRecorder()
	
	result := accessCheck(app, w, req)
	if !result { t.Error("Expected true, got false") } // Should pass when no user validation required
}

func TestRequireAuthWithRender(t *testing.T) {
	auth := New("test-secret")
	cookieName := "test-session"
	failTemplate := "signin.html"
	
	// Mock user function
	mockGetUser := func(userID string) (any, error) {
		if userID == "valid-user" {
			return map[string]any{"id": userID}, nil
		}
		return nil, errors.New("user not found")
	}
	
	validToken, err := auth.GenerateSessionToken("valid-user", time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	accessCheck := RequireAuthWithRender(auth, cookieName, mockGetUser, failTemplate)
	
	tests := []struct {
		name        string
		setupCookie func(*http.Request)
		expected    bool
		expectRender bool
	}{
		{
			name: "valid authentication",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: cookieName, Value: validToken})
			},
			expected:     true,
			expectRender: false,
		},
		{
			name: "no authentication",
			setupCookie: func(r *http.Request) {
				// No cookie
			},
			expected:     false,
			expectRender: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			
			tt.setupCookie(req)
			
			result := accessCheck(app, w, req)
			if tt.expected != result { t.Errorf("Expected %v, got %v", tt.expected, result) }
			
			if tt.expectRender {
				// Check that render was called with correct template
				response := w.Body.String()
				if !strings.Contains(response, failTemplate) { t.Errorf("Expected to contain %v", failTemplate) }
			}
		})
	}
}

func TestRequireFunc(t *testing.T) {
	tests := []struct {
		name       string
		checkFunc  func(*http.Request) bool
		setupReq   func(*http.Request)
		expected   bool
	}{
		{
			name: "function returns true",
			checkFunc: func(r *http.Request) bool {
				return true
			},
			setupReq: func(r *http.Request) {},
			expected: true,
		},
		{
			name: "function returns false",
			checkFunc: func(r *http.Request) bool {
				return false
			},
			setupReq: func(r *http.Request) {},
			expected: false,
		},
		{
			name: "function checks header",
			checkFunc: func(r *http.Request) bool {
				return r.Header.Get("X-Special") == "allowed"
			},
			setupReq: func(r *http.Request) {
				r.Header.Set("X-Special", "allowed")
			},
			expected: true,
		},
		{
			name: "function checks header - fails",
			checkFunc: func(r *http.Request) bool {
				return r.Header.Get("X-Special") == "allowed"
			},
			setupReq: func(r *http.Request) {
				r.Header.Set("X-Special", "denied")
			},
			expected: false,
		},
		{
			name: "function checks method",
			checkFunc: func(r *http.Request) bool {
				return r.Method == "POST"
			},
			setupReq: func(r *http.Request) {
				// Request is already GET by default
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessCheck := RequireFunc(tt.checkFunc)
			
			app := setupTestApp()
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			
			tt.setupReq(req)
			
			result := accessCheck(app, w, req)
			if tt.expected != result { t.Errorf("Expected %v, got %v", tt.expected, result) }
		})
	}
}

func TestPublicAccess(t *testing.T) {
	accessCheck := PublicAccess()
	
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	result := accessCheck(app, w, req)
	if !result { t.Error("Expected true, got false") }
	
	// Should work regardless of headers, cookies, etc.
	req.Header.Set("Authorization", "invalid")
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid"})
	
	result = accessCheck(app, w, req)
	if !result { t.Error("Expected true, got false") }
}

func TestCombineChecks(t *testing.T) {
	// Create individual check functions
	checkAlwaysTrue := func(r *http.Request) bool { return true }
	checkAlwaysFalse := func(r *http.Request) bool { return false }
	checkMethod := func(r *http.Request) bool { return r.Method == "POST" }
	checkHeader := func(r *http.Request) bool { return r.Header.Get("X-Valid") == "true" }
	
	// Convert to AccessCheck functions
	accessAlwaysTrue := RequireFunc(checkAlwaysTrue)
	accessAlwaysFalse := RequireFunc(checkAlwaysFalse)
	accessMethod := RequireFunc(checkMethod)
	accessHeader := RequireFunc(checkHeader)
	publicAccess := PublicAccess()
	
	tests := []struct {
		name      string
		checks    []AccessCheck
		setupReq  func(*http.Request)
		expected  bool
	}{
		{
			name:     "all pass",
			checks:   []AccessCheck{accessAlwaysTrue, publicAccess},
			setupReq: func(r *http.Request) {},
			expected: true,
		},
		{
			name:     "one fails",
			checks:   []AccessCheck{accessAlwaysTrue, accessAlwaysFalse},
			setupReq: func(r *http.Request) {},
			expected: false,
		},
		{
			name:     "empty checks",
			checks:   []AccessCheck{},
			setupReq: func(r *http.Request) {},
			expected: true,
		},
		{
			name:     "method and header both pass",
			checks:   []AccessCheck{accessMethod, accessHeader},
			setupReq: func(r *http.Request) {
				r.Method = "POST"
				r.Header.Set("X-Valid", "true")
			},
			expected: true,
		},
		{
			name:     "method passes, header fails",
			checks:   []AccessCheck{accessMethod, accessHeader},
			setupReq: func(r *http.Request) {
				r.Method = "POST"
				r.Header.Set("X-Valid", "false")
			},
			expected: false,
		},
		{
			name:     "three checks - one fails",
			checks:   []AccessCheck{publicAccess, accessAlwaysTrue, accessAlwaysFalse},
			setupReq: func(r *http.Request) {},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combinedCheck := CombineChecks(tt.checks...)
			
			app := setupTestApp()
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			
			tt.setupReq(req)
			
			result := combinedCheck(app, w, req)
			if tt.expected != result { t.Errorf("Expected %v, got %v", tt.expected, result) }
		})
	}
}

func TestAccessControlIntegration(t *testing.T) {
	auth := New("test-secret")
	cookieName := "session"
	
	// Mock user database
	users := map[string]any{
		"user-1": map[string]any{"id": "user-1", "name": "User One", "role": "user"},
		"user-2": map[string]any{"id": "user-2", "name": "User Two", "role": "admin"},
	}
	
	getUserFunc := func(userID string) (any, error) {
		if user, exists := users[userID]; exists {
			return user, nil
		}
		return nil, errors.New("user not found")
	}
	
	// Create tokens
	userToken, err := auth.GenerateSessionToken("user-1", time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	adminToken, err := auth.GenerateSessionToken("user-2", time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	invalidToken, err := auth.GenerateSessionToken("user-999", time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	// Create access checks
	requireAuth := RequireAuth(auth, cookieName, getUserFunc)
	requireAuthWithRender := RequireAuthWithRender(auth, cookieName, getUserFunc, "signin.html")
	requireAdmin := RequireFunc(func(r *http.Request) bool {
		// Get token from cookie
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return false
		}
		
		// Validate token
		claims, err := auth.ValidateToken(cookie.Value)
		if err != nil {
			return false
		}
		
		userID, ok := claims["user_id"].(string)
		if !ok {
			return false
		}
		
		// Get user
		user, err := getUserFunc(userID)
		if err != nil {
			return false
		}
		
		// Check if admin
		userMap := user.(map[string]any)
		return userMap["role"] == "admin"
	})
	
	// Combine auth + admin check
	requireAuthAndAdmin := CombineChecks(requireAuth, requireAdmin)
	
	tests := []struct {
		name        string
		accessCheck AccessCheck
		token       string
		expected    bool
	}{
		// Basic auth tests
		{
			name:        "require auth - valid user",
			accessCheck: requireAuth,
			token:       userToken,
			expected:    true,
		},
		{
			name:        "require auth - valid admin",
			accessCheck: requireAuth,
			token:       adminToken,
			expected:    true,
		},
		{
			name:        "require auth - invalid user",
			accessCheck: requireAuth,
			token:       invalidToken,
			expected:    false,
		},
		{
			name:        "require auth - no token",
			accessCheck: requireAuth,
			token:       "",
			expected:    false,
		},
		
		// Auth with render tests
		{
			name:        "require auth with render - valid",
			accessCheck: requireAuthWithRender,
			token:       userToken,
			expected:    true,
		},
		{
			name:        "require auth with render - invalid",
			accessCheck: requireAuthWithRender,
			token:       invalidToken,
			expected:    false,
		},
		
		// Admin-only tests
		{
			name:        "require admin - regular user",
			accessCheck: requireAdmin,
			token:       userToken,
			expected:    false,
		},
		{
			name:        "require admin - admin user",
			accessCheck: requireAdmin,
			token:       adminToken,
			expected:    true,
		},
		{
			name:        "require admin - no token",
			accessCheck: requireAdmin,
			token:       "",
			expected:    false,
		},
		
		// Combined checks
		{
			name:        "require auth and admin - regular user",
			accessCheck: requireAuthAndAdmin,
			token:       userToken,
			expected:    false,
		},
		{
			name:        "require auth and admin - admin user",
			accessCheck: requireAuthAndAdmin,
			token:       adminToken,
			expected:    true,
		},
		{
			name:        "require auth and admin - invalid user",
			accessCheck: requireAuthAndAdmin,
			token:       invalidToken,
			expected:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()
			req := httptest.NewRequest("GET", "/protected", nil)
			w := httptest.NewRecorder()
			
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: cookieName, Value: tt.token})
			}
			
			result := tt.accessCheck(app, w, req)
			if tt.expected != result { t.Errorf("Expected %v, got %v", tt.expected, result) }
		})
	}
}

func TestAccessControlEdgeCases(t *testing.T) {
	auth := New("test-secret")
	
	t.Run("require auth with panicking user function", func(t *testing.T) {
		panicUserFunc := func(userID string) (any, error) {
			panic("database error")
		}
		
		accessCheck := RequireAuth(auth, "session", panicUserFunc)
		validToken, _ := auth.GenerateSessionToken("user-1", time.Hour)
		
		app := setupTestApp()
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: validToken})
		w := httptest.NewRecorder()
		
		// Should handle panic gracefully and return false
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Access check should handle panics gracefully, but panicked with: %v", r)
			}
		}()
		
		result := accessCheck(app, w, req)
		if result { t.Error("Expected false, got true") } // Should fail gracefully
	})
	
	t.Run("combine checks with nil checks", func(t *testing.T) {
		// This should not panic
		combinedCheck := CombineChecks()
		
		app := setupTestApp()
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		
		result := combinedCheck(app, w, req)
		if !result { t.Error("Expected true, got false") } // Empty check list should pass
	})
	
	t.Run("require func with nil function", func(t *testing.T) {
		// This will panic, which is expected behavior
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when passing nil function to RequireFunc")
			}
		}()
		
		RequireFunc(nil)
	})
}

func TestAccessControlPerformance(t *testing.T) {
	auth := New("test-secret")
	
	// Create user function that simulates database lookup
	getUserFunc := func(userID string) (any, error) {
		// Simulate some work
		time.Sleep(time.Microsecond)
		if userID == "valid-user" {
			return map[string]any{"id": userID}, nil
		}
		return nil, errors.New("not found")
	}
	
	accessCheck := RequireAuth(auth, "session", getUserFunc)
	validToken, _ := auth.GenerateSessionToken("valid-user", time.Hour)
	
	// Run multiple times to test performance
	iterations := 100
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		app := setupTestApp()
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: validToken})
		w := httptest.NewRecorder()
		
		result := accessCheck(app, w, req)
		if !result { t.Error("Expected true, got false") }
	}
	
	elapsed := time.Since(start)
	averageTime := elapsed / time.Duration(iterations)
	
	// Should be reasonably fast (less than 1ms per check on average)
	if averageTime > time.Millisecond {
		t.Errorf("Access check too slow: %v average per check", averageTime)
	}
}

func TestAccessControlWithDifferentSecrets(t *testing.T) {
	auth1 := New("secret-1")
	auth2 := New("secret-2")
	
	getUserFunc := func(userID string) (any, error) {
		return map[string]any{"id": userID}, nil
	}
	
	// Create token with first auth
	token, err := auth1.GenerateSessionToken("user-1", time.Hour)
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	
	// Try to validate with second auth (different secret)
	accessCheck := RequireAuth(auth2, "session", getUserFunc)
	
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	
	result := accessCheck(app, w, req)
	if result { t.Error("Expected false, got true") } // Should fail due to different secret
}