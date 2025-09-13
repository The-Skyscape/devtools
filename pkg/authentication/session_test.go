package authentication

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSessionTable(t *testing.T) {
	session := &Session{}
}

func TestSessionToken(t *testing.T) {
	// Set auth secret for token generation
	os.Setenv("AUTH_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_SECRET")
	
	session := &Session{
		UserID: "user-123",
	}
	
	// Set session ID manually since it's normally set by database
	session.ID = "session-456"
	
	beforeGeneration := time.Now()
	token, err := session.Token()
	afterGeneration := time.Now()
	
	
	// Parse and validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	
	
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	
	// Check claims
	
	// Check issued at time
	iat, ok := claims["iat"].(float64)
	issuedAt := time.Unix(int64(iat), 0)
	
	// Check expiration time (should be 7 days from now)
	exp, ok := claims["exp"].(float64)
	expiresAt := time.Unix(int64(exp), 0)
	expectedExpiry := beforeGeneration.Add(time.Hour * 24 * 7)
}

func TestSessionTokenWithoutAuthSecret(t *testing.T) {
	// Ensure no auth secret is set
	originalSecret := os.Getenv("AUTH_SECRET")
	os.Unsetenv("AUTH_SECRET")
	defer func() {
		if originalSecret != "" {
			os.Setenv("AUTH_SECRET", originalSecret)
		}
	}()
	
	session := &Session{
		UserID: "user-123",
	}
	session.ID = "session-456"
	
	token, err := session.Token()
	
	// Should still work with empty secret (though not secure)
}

func TestSessionTokenMultiple(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_SECRET")
	
	session1 := &Session{UserID: "user-1"}
	session1.ID = "session-1"
	
	session2 := &Session{UserID: "user-2"}
	session2.ID = "session-2"
	
	token1, err := session1.Token()
	
	token2, err := session2.Token()
	
	// Tokens should be different
	
	// Both tokens should be valid and contain correct session IDs
	for i, tokenData := range []struct {
		token     string
		sessionID string
	}{
		{token1, "session-1"},
		{token2, "session-2"},
	} {
		parsedToken, err := jwt.Parse(tokenData.token, func(token *jwt.Token) (any, error) {
			return []byte("test-secret"), nil
		})
		
		
		claims := parsedToken.Claims.(jwt.MapClaims)
	}
}

func TestControllerAuthenticate(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_SECRET")
	
	controller, _ := setupTestController(t)
	
	// Create test user
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	// Create session for user
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	// Generate token
	token, err := session.Token()
	
	tests := []struct {
		name           string
		setupCookie    func(*http.Request)
		expectUser     bool
		expectSession  bool
		expectError    bool
	}{
		{
			name: "valid session cookie",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: token,
				})
			},
			expectUser:    true,
			expectSession: true,
			expectError:   false,
		},
		{
			name: "no cookie",
			setupCookie: func(r *http.Request) {
				// No cookie added
			},
			expectUser:    false,
			expectSession: false,
			expectError:   true,
		},
		{
			name: "invalid token",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: "invalid-token",
				})
			},
			expectUser:    false,
			expectSession: false,
			expectError:   true,
		},
		{
			name: "empty token",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: "",
				})
			},
			expectUser:    false,
			expectSession: false,
			expectError:   true,
		},
		{
			name: "wrong cookie name",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "wrong-cookie",
					Value: token,
				})
			},
			expectUser:    false,
			expectSession: false,
			expectError:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			tt.setupCookie(req)
			
			resultUser, resultSession, err := controller.Authenticate(req)
			
			if tt.expectError {
			} else {
			}
			
			if tt.expectUser {
			} else {
			}
			
			if tt.expectSession {
			} else {
			}
		})
	}
}

func TestAuthenticateWithExpiredToken(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_SECRET")
	
	controller, _ := setupTestController(t)
	
	// Create test user
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	// Create session for user
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	// Create an expired token manually
	expiredClaims := jwt.MapClaims{
		"sub": session.ID,
		"iat": time.Now().Add(-time.Hour * 2).Unix(),
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
	}
	
	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("test-secret"))
	
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  controller.cookieName,
		Value: expiredToken,
	})
	
	resultUser, resultSession, err := controller.Authenticate(req)
	
}

func TestAuthenticateWithNonexistentSession(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_SECRET")
	
	controller, _ := setupTestController(t)
	
	// Create a token with a nonexistent session ID
	nonexistentClaims := jwt.MapClaims{
		"sub": "nonexistent-session-id",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	
	nonexistentToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, nonexistentClaims).SignedString([]byte("test-secret"))
	
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  controller.cookieName,
		Value: nonexistentToken,
	})
	
	resultUser, resultSession, err := controller.Authenticate(req)
	
}

func TestAuthenticateWithNonexistentUser(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_SECRET")
	
	controller, _ := setupTestController(t)
	
	// Create session with nonexistent user ID
	session, err := controller.Sessions.Insert(&Session{UserID: "nonexistent-user-id"})
	
	// Generate token for this session
	token, err := session.Token()
	
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  controller.cookieName,
		Value: token,
	})
	
	resultUser, resultSession, err := controller.Authenticate(req)
	
}

func TestAuthenticateWithDifferentSecrets(t *testing.T) {
	// Create token with one secret
	os.Setenv("AUTH_SECRET", "secret1")
	
	controller, _ := setupTestController(t)
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	token, err := session.Token()
	
	// Change secret and try to authenticate
	os.Setenv("AUTH_SECRET", "secret2")
	defer os.Unsetenv("AUTH_SECRET")
	
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  controller.cookieName,
		Value: token,
	})
	
	resultUser, resultSession, err := controller.Authenticate(req)
	
}

func TestSessionIntegration(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_SECRET")
	
	controller, _ := setupTestController(t)
	
	// Create multiple users and sessions
	users := []struct {
		name     string
		email    string
		handle   string
		password string
	}{
		{"User One", "user1@example.com", "user1", "pass1"},
		{"User Two", "user2@example.com", "user2", "pass2"},
		{"User Three", "user3@example.com", "user3", "pass3"},
	}
	
	var createdUsers []*User
	var createdSessions []*Session
	var tokens []string
	
	// Create users and sessions
	for _, userData := range users {
		user, err := controller.Signup(userData.name, userData.email, userData.handle, userData.password, false)
		createdUsers = append(createdUsers, user)
		
		session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
		createdSessions = append(createdSessions, session)
		
		token, err := session.Token()
		tokens = append(tokens, token)
	}
	
	// Test authentication for each user
	for i, token := range tokens {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  controller.cookieName,
			Value: token,
		})
		
		resultUser, resultSession, err := controller.Authenticate(req)
		
		
	}
	
	// Test that tokens are unique
	for i, token1 := range tokens {
		for j, token2 := range tokens {
			if i != j {
			}
		}
	}
}

func TestSessionTokenConsistency(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_SECRET")
	
	session := &Session{UserID: "user-123"}
	session.ID = "session-456"
	
	// Generate multiple tokens for the same session
	tokens := make([]string, 5)
	for i := 0; i < 5; i++ {
		token, err := session.Token()
		tokens[i] = token
	}
	
	// All tokens should be different (due to different iat times)
	for i, token1 := range tokens {
		for j, token2 := range tokens {
			if i != j {
			}
		}
	}
	
	// But they should all contain the same session ID
	for i, token := range tokens {
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
			return []byte("test-secret"), nil
		})
		
		
		claims := parsedToken.Claims.(jwt.MapClaims)
	}
}