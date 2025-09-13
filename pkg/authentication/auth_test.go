package authentication

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func assertEqual(t *testing.T, expected, actual any, message ...string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := "Values not equal"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nExpected: %v\nActual: %v", msg, expected, actual)
	}
}

func assertNotEqual(t *testing.T, expected, actual any, message ...string) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		msg := "Values are equal"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nValue: %v", msg, actual)
	}
}

func assertNil(t *testing.T, value any, message ...string) {
	t.Helper()
	if !isNil(value) {
		msg := "Expected nil"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nGot: %v", msg, value)
	}
}

func assertNotNil(t *testing.T, value any, message ...string) {
	t.Helper()
	if isNil(value) {
		msg := "Expected non-nil value"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s", msg)
	}
}

func assertTrue(t *testing.T, value bool, message ...string) {
	t.Helper()
	if !value {
		msg := "Expected true"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s", msg)
	}
}

func assertFalse(t *testing.T, value bool, message ...string) {
	t.Helper()
	if value {
		msg := "Expected false"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s", msg)
	}
}

func assertNoError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err != nil {
		msg := "Unexpected error"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s: %v", msg, err)
	}
}

func assertError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err == nil {
		msg := "Expected error but got nil"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s", msg)
	}
}

func assertTimeBetween(t *testing.T, actual, start, end time.Time, message ...string) {
	t.Helper()
	if actual.Before(start) || actual.After(end) {
		msg := "Time not in expected range"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nActual: %v\nExpected between: %v and %v", msg, actual, start, end)
	}
}

func assertTimeWithin(t *testing.T, actual, expected time.Time, tolerance time.Duration, message ...string) {
	t.Helper()
	diff := actual.Sub(expected)
	if diff < -tolerance || diff > tolerance {
		msg := "Time not within tolerance"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nActual: %v\nExpected: %v ±%v\nDifference: %v", msg, actual, expected, tolerance, diff)
	}
}

func assertLen(t *testing.T, value any, expected int, message ...string) {
	t.Helper()
	actual := reflect.ValueOf(value).Len()
	if actual != expected {
		msg := "Length mismatch"
		if len(message) > 0 {
			msg = message[0]
		}
		t.Errorf("%s\nExpected length: %d\nActual length: %d", msg, expected, actual)
	}
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		envVar   string
		expected string
	}{
		{
			name:     "with explicit secret",
			secret:   "test-secret",
			envVar:   "",
			expected: "test-secret",
		},
		{
			name:     "with environment variable",
			secret:   "",
			envVar:   "env-secret",
			expected: "env-secret",
		},
		{
			name:     "explicit secret takes precedence",
			secret:   "explicit-secret",
			envVar:   "env-secret",
			expected: "explicit-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if specified
			if tt.envVar != "" {
				os.Setenv("AUTH_SECRET", tt.envVar)
				defer os.Unsetenv("AUTH_SECRET")
			}

			auth := New(tt.secret)
			assertNotNil(t, auth)
			assertEqual(t, tt.expected, auth.secret)
		})
	}
}

func TestHashPassword(t *testing.T) {
	auth := New("test-secret")

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "valid-password",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
		},
		{
			name:        "long password",
			password:    "this-is-a-very-long-password-that-should-still-work-fine",
			expectError: false,
		},
		{
			name:        "special characters",
			password:    "p@ssw0rd!#$%",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := auth.HashPassword(tt.password)

			if tt.expectError {
				assertError(t, err)
				assertEqual(t, "", hash)
			} else {
				assertNoError(t, err)
				assertNotEqual(t, "", hash)
				assertNotEqual(t, tt.password, hash, "Hash should not equal original password")
				
				// Verify the hash can be used for verification
				assertTrue(t, auth.VerifyPassword(hash, tt.password))
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	auth := New("test-secret")
	password := "test-password"
	
	// Create a valid hash
	validHash, err := auth.HashPassword(password)
	assertNoError(t, err)

	tests := []struct {
		name     string
		hash     string
		password string
		expected bool
	}{
		{
			name:     "correct password",
			hash:     validHash,
			password: password,
			expected: true,
		},
		{
			name:     "incorrect password",
			hash:     validHash,
			password: "wrong-password",
			expected: false,
		},
		{
			name:     "empty password",
			hash:     validHash,
			password: "",
			expected: false,
		},
		{
			name:     "empty hash",
			hash:     "",
			password: password,
			expected: false,
		},
		{
			name:     "invalid hash",
			hash:     "invalid-hash",
			password: password,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.VerifyPassword(tt.hash, tt.password)
			assertEqual(t, tt.expected, result)
		})
	}
}

func TestGenerateToken(t *testing.T) {
	auth := New("test-secret")

	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{
			name: "simple claims",
			claims: jwt.MapClaims{
				"user_id": "123",
				"exp":     time.Now().Add(time.Hour).Unix(),
			},
		},
		{
			name: "complex claims",
			claims: jwt.MapClaims{
				"user_id":    "user-456",
				"session_id": "session-789",
				"role":       "admin",
				"exp":        time.Now().Add(time.Hour * 24).Unix(),
				"iat":        time.Now().Unix(),
			},
		},
		{
			name:   "empty claims",
			claims: jwt.MapClaims{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tt.claims)
			assertNoError(t, err)
			assertNotEqual(t, "", token)

			// Verify the token can be parsed back
			parsedClaims, err := auth.ValidateToken(token)
			assertNoError(t, err)
			
			// Check that all original claims are present
			for key, value := range tt.claims {
				actualValue := parsedClaims[key]
				// Handle float64 conversion for numeric values
				if expectedFloat, ok := value.(int64); ok {
					if actualFloat, ok := actualValue.(float64); ok {
						assertEqual(t, float64(expectedFloat), actualFloat)
					} else {
						assertEqual(t, value, actualValue)
					}
				} else {
					assertEqual(t, value, actualValue)
				}
			}
		})
	}
}

func TestGenerateSessionToken(t *testing.T) {
	auth := New("test-secret")

	tests := []struct {
		name      string
		userID    string
		expiresIn time.Duration
	}{
		{
			name:      "one hour session",
			userID:    "user-123",
			expiresIn: time.Hour,
		},
		{
			name:      "one day session",
			userID:    "user-456",
			expiresIn: time.Hour * 24,
		},
		{
			name:      "short session",
			userID:    "user-789",
			expiresIn: time.Minute * 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeGeneration := time.Now()
			token, err := auth.GenerateSessionToken(tt.userID, tt.expiresIn)
			afterGeneration := time.Now()

			assertNoError(t, err)
			assertNotEqual(t, "", token)

			// Validate the token and check claims
			claims, err := auth.ValidateToken(token)
			assertNoError(t, err)

			// Check user_id
			assertEqual(t, tt.userID, claims["user_id"])

			// Check session_id exists and is not empty
			sessionID, ok := claims["session_id"].(string)
			assertTrue(t, ok)
			assertNotEqual(t, "", sessionID)

			// Check timestamps
			iat, ok := claims["iat"].(float64)
			assertTrue(t, ok)
			assertTimeBetween(t, time.Unix(int64(iat), 0), beforeGeneration.Add(-time.Second), afterGeneration.Add(time.Second))

			exp, ok := claims["exp"].(float64)
			assertTrue(t, ok)
			expectedExp := beforeGeneration.Add(tt.expiresIn)
			assertTimeWithin(t, time.Unix(int64(exp), 0), expectedExp, time.Second*30)
		})
	}
}

func TestValidateToken(t *testing.T) {
	auth := New("test-secret")
	otherAuth := New("different-secret")

	// Create a valid token
	validClaims := jwt.MapClaims{
		"user_id": "user-123",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	validToken, err := auth.GenerateToken(validClaims)
	assertNoError(t, err)

	// Create an expired token
	expiredClaims := jwt.MapClaims{
		"user_id": "user-456",
		"exp":     time.Now().Add(-time.Hour).Unix(), // Already expired
		"iat":     time.Now().Add(-time.Hour * 2).Unix(),
	}
	expiredToken, err := auth.GenerateToken(expiredClaims)
	assertNoError(t, err)

	// Create a token with different secret
	wrongSecretToken, err := otherAuth.GenerateToken(validClaims)
	assertNoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid token",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "expired token",
			token:       expiredToken,
			expectError: true,
		},
		{
			name:        "token with wrong secret",
			token:       wrongSecretToken,
			expectError: true,
		},
		{
			name:        "malformed token",
			token:       "invalid.token.format",
			expectError: true,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "random string",
			token:       "random-string-not-jwt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := auth.ValidateToken(tt.token)

			if tt.expectError {
				assertError(t, err)
				assertNil(t, claims)
			} else {
				assertNoError(t, err)
				assertNotNil(t, claims)
			}
		})
	}
}

func TestGetTokenFromCookie(t *testing.T) {
	auth := New("test-secret")

	tests := []struct {
		name        string
		cookieName  string
		cookieValue string
		expectError bool
	}{
		{
			name:        "valid cookie",
			cookieName:  "auth",
			cookieValue: "token-value",
			expectError: false,
		},
		{
			name:        "empty cookie value",
			cookieName:  "auth",
			cookieValue: "",
			expectError: false,
		},
		{
			name:        "missing cookie",
			cookieName:  "nonexistent",
			cookieValue: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			
			// Add cookie if not testing missing case
			if tt.name != "missing cookie" {
				req.AddCookie(&http.Cookie{
					Name:  tt.cookieName,
					Value: tt.cookieValue,
				})
			}

			token, err := auth.GetTokenFromCookie(req, tt.cookieName)

			if tt.expectError {
				assertError(t, err)
				assertEqual(t, "", token)
			} else {
				assertNoError(t, err)
				assertEqual(t, tt.cookieValue, token)
			}
		})
	}
}

func TestGetTokenFromHeader(t *testing.T) {
	auth := New("test-secret")

	tests := []struct {
		name        string
		header      string
		expected    string
		expectError bool
	}{
		{
			name:        "bearer token",
			header:      "Bearer token-value",
			expected:    "token-value",
			expectError: false,
		},
		{
			name:        "bearer token with extra spaces",
			header:      "Bearer  token-with-spaces  ",
			expected:    " token-with-spaces  ",
			expectError: false,
		},
		{
			name:        "token without bearer prefix",
			header:      "raw-token-value",
			expected:    "raw-token-value",
			expectError: false,
		},
		{
			name:        "empty header",
			header:      "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "bearer with no token",
			header:      "Bearer ",
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			token, err := auth.GetTokenFromHeader(req)

			if tt.expectError {
				assertError(t, err)
			} else {
				assertNoError(t, err)
				assertEqual(t, tt.expected, token)
			}
		})
	}
}

func TestSetCookie(t *testing.T) {
	auth := New("test-secret")

	tests := []struct {
		name    string
		name_   string
		value   string
		expires time.Time
		secure  bool
	}{
		{
			name:    "basic cookie",
			name_:   "session",
			value:   "token-value",
			expires: time.Now().Add(time.Hour),
			secure:  false,
		},
		{
			name:    "secure cookie",
			name_:   "secure-session",
			value:   "secure-token",
			expires: time.Now().Add(time.Hour * 24),
			secure:  true,
		},
		{
			name:    "empty value",
			name_:   "empty",
			value:   "",
			expires: time.Now().Add(time.Hour),
			secure:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			auth.SetCookie(w, tt.name_, tt.value, tt.expires, tt.secure)

			cookies := w.Result().Cookies()
			assertLen(t, cookies, 1)

			cookie := cookies[0]
			assertEqual(t, tt.name_, cookie.Name)
			assertEqual(t, tt.value, cookie.Value)
			assertEqual(t, "/", cookie.Path)
			assertTrue(t, cookie.HttpOnly)
			assertEqual(t, tt.secure, cookie.Secure)
			assertEqual(t, http.SameSiteLaxMode, cookie.SameSite)
			
			// Check expires is approximately correct (within 5 seconds)
			assertTimeWithin(t, cookie.Expires, tt.expires, time.Second*5)
		})
	}
}

func TestClearCookie(t *testing.T) {
	auth := New("test-secret")

	tests := []string{"session", "auth-token", "user-pref"}

	for _, cookieName := range tests {
		t.Run("clear "+cookieName, func(t *testing.T) {
			w := httptest.NewRecorder()
			auth.ClearCookie(w, cookieName)

			cookies := w.Result().Cookies()
			assertLen(t, cookies, 1)

			cookie := cookies[0]
			assertEqual(t, cookieName, cookie.Name)
			assertEqual(t, "", cookie.Value)
			assertEqual(t, "/", cookie.Path)
			assertEqual(t, -1, cookie.MaxAge)
			assertTrue(t, cookie.HttpOnly)
		})
	}
}

func TestGenerateID(t *testing.T) {
	// Test multiple generations to ensure uniqueness
	ids := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		id := GenerateID()
		
		// Check format (should be 32 hex characters)
		assertEqual(t, 32, len(id))
		
		// Check uniqueness
		assertFalse(t, ids[id], "Generated duplicate ID: "+id)
		ids[id] = true
		
		// Check it's valid hex
		for _, char := range id {
			valid := (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')
			assertTrue(t, valid, "Invalid hex character in ID: "+string(char))
		}
	}
}

func TestMiddleware(t *testing.T) {
	auth := New("test-secret")

	// Mock user function
	mockGetUser := func(userID string) (any, error) {
		if userID == "valid-user" {
			return map[string]any{"id": userID, "name": "Test User"}, nil
		}
		return nil, jwt.ErrInvalidKey // User not found
	}

	// Create test handler that checks context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		claims := GetClaims(r)
		
		if user != nil {
			w.Header().Set("X-User-Found", "true")
		}
		if claims != nil {
			w.Header().Set("X-Claims-Found", "true")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.Middleware("session", mockGetUser)
	handler := middleware(testHandler)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectUser     bool
		expectClaims   bool
	}{
		{
			name: "valid cookie with valid user",
			setupRequest: func(r *http.Request) {
				token, _ := auth.GenerateSessionToken("valid-user", time.Hour)
				r.AddCookie(&http.Cookie{Name: "session", Value: token})
			},
			expectUser:   true,
			expectClaims: true,
		},
		{
			name: "valid header with valid user",
			setupRequest: func(r *http.Request) {
				token, _ := auth.GenerateSessionToken("valid-user", time.Hour)
				r.Header.Set("Authorization", "Bearer "+token)
			},
			expectUser:   true,
			expectClaims: true,
		},
		{
			name: "valid token but invalid user",
			setupRequest: func(r *http.Request) {
				token, _ := auth.GenerateSessionToken("invalid-user", time.Hour)
				r.AddCookie(&http.Cookie{Name: "session", Value: token})
			},
			expectUser:   false,
			expectClaims: false,
		},
		{
			name: "invalid token",
			setupRequest: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: "session", Value: "invalid-token"})
			},
			expectUser:   false,
			expectClaims: false,
		},
		{
			name: "no token",
			setupRequest: func(r *http.Request) {
				// No token provided
			},
			expectUser:   false,
			expectClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			tt.setupRequest(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assertEqual(t, http.StatusOK, w.Code)
			
			if tt.expectUser {
				assertEqual(t, "true", w.Header().Get("X-User-Found"))
			} else {
				assertEqual(t, "", w.Header().Get("X-User-Found"))
			}
			
			if tt.expectClaims {
				assertEqual(t, "true", w.Header().Get("X-Claims-Found"))
			} else {
				assertEqual(t, "", w.Header().Get("X-Claims-Found"))
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	// Test with no user in context
	user := GetUser(req)
	assertNil(t, user)

	// Test with user in context
	expectedUser := map[string]any{"id": "123", "name": "Test User"}
	ctx := context.WithValue(req.Context(), UserContextKey, expectedUser)
	req = req.WithContext(ctx)

	user = GetUser(req)
	assertEqual(t, expectedUser, user)
}

func TestGetClaims(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	// Test with no claims in context
	claims := GetClaims(req)
	assertNil(t, claims)

	// Test with claims in context
	expectedClaims := jwt.MapClaims{"user_id": "123", "role": "admin"}
	ctx := context.WithValue(req.Context(), ClaimsContextKey, expectedClaims)
	req = req.WithContext(ctx)

	claims = GetClaims(req)
	assertEqual(t, expectedClaims, claims)

	// Test with wrong type in context
	ctx = context.WithValue(req.Context(), ClaimsContextKey, "wrong-type")
	req = req.WithContext(ctx)

	claims = GetClaims(req)
	assertNil(t, claims)
}