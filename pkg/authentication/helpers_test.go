package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

)

func TestGetAuthenticatedUser(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user and session
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	token, err := session.Token()
	
	tests := []struct {
		name        string
		setupCookie bool
		expectUser  bool
	}{
		{
			name:        "with valid authentication",
			setupCookie: true,
			expectUser:  true,
		},
		{
			name:        "without authentication",
			setupCookie: false,
			expectUser:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			
			if tt.setupCookie {
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: token,
				})
			}
			
			resultUser := controller.GetAuthenticatedUser(req)
			
			if tt.expectUser {
			} else {
			}
		})
	}
}

func TestGetAuthenticatedSession(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user and session
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	token, err := session.Token()
	
	tests := []struct {
		name           string
		setupCookie    bool
		expectSession  bool
	}{
		{
			name:          "with valid authentication",
			setupCookie:   true,
			expectSession: true,
		},
		{
			name:          "without authentication",
			setupCookie:   false,
			expectSession: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			
			if tt.setupCookie {
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: token,
				})
			}
			
			resultSession := controller.GetAuthenticatedSession(req)
			
			if tt.expectSession {
			} else {
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create regular user
	regularUser, err := controller.Signup("Regular User", "regular@example.com", "regular", "password", false)
	
	regularSession, err := controller.Sessions.Insert(&Session{UserID: regularUser.ID})
	
	regularToken, err := regularSession.Token()
	
	// Create admin user
	adminUser, err := controller.Signup("Admin User", "admin@example.com", "admin", "password", true)
	
	adminSession, err := controller.Sessions.Insert(&Session{UserID: adminUser.ID})
	
	adminToken, err := adminSession.Token()
	
	tests := []struct {
		name        string
		token       string
		expected    bool
	}{
		{
			name:     "regular user",
			token:    regularToken,
			expected: false,
		},
		{
			name:     "admin user",
			token:    adminToken,
			expected: true,
		},
		{
			name:     "no authentication",
			token:    "",
			expected: false,
		},
		{
			name:     "invalid token",
			token:    "invalid-token",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			
			if tt.token != "" {
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: tt.token,
				})
			}
			
			result := controller.IsAdmin(req)
		})
	}
}

func TestRequireUser(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	token, err := session.Token()
	
	tests := []struct {
		name        string
		setupCookie bool
		expectUser  bool
	}{
		{
			name:        "with valid authentication",
			setupCookie: true,
			expectUser:  true,
		},
		{
			name:        "without authentication",
			setupCookie: false,
			expectUser:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			
			if tt.setupCookie {
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: token,
				})
			}
			
			resultUser := controller.RequireUser(req)
			
			if tt.expectUser {
			} else {
			}
		})
	}
}

func TestLogin(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user (without automatic session)
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	// Test login
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	
	err = controller.Login(w, req, user)
	
	// Should have set a cookie
	cookies := w.Result().Cookies()
	
	cookie := cookies[0]
	
	// Cookie should be valid for authentication
	authReq := httptest.NewRequest("GET", "/", nil)
	authReq.AddCookie(cookie)
	
	authenticatedUser := controller.GetAuthenticatedUser(authReq)
}

func TestLoginError(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Test login with nil user
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	
	err := controller.Login(w, req, nil)
	
	// Should not have set any cookies
	cookies := w.Result().Cookies()
}

func TestLogout(t *testing.T) {
	controller, _ := setupTestController(t)
	
	tests := []struct {
		name    string
		isHTTPS bool
	}{
		{
			name:    "HTTP logout",
			isHTTPS: false,
		},
		{
			name:    "HTTPS logout",
			isHTTPS: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/logout", nil)
			
			if tt.isHTTPS {
				req.TLS = &struct{}{} // Mark as HTTPS
			}
			
			controller.Logout(w, req)
			
			// Should have set a clearing cookie
			cookies := w.Result().Cookies()
			
			cookie := cookies[0]
		})
	}
}

func TestSetSessionCookie(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user and session
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	tests := []struct {
		name    string
		isHTTPS bool
	}{
		{
			name:    "HTTP session cookie",
			isHTTPS: false,
		},
		{
			name:    "HTTPS session cookie",
			isHTTPS: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/login", nil)
			
			if tt.isHTTPS {
				req.TLS = &struct{}{} // Mark as HTTPS
			}
			
			controller.SetSessionCookie(w, req, session)
			
			// Should have set a session cookie
			cookies := w.Result().Cookies()
			
			cookie := cookies[0]
			
			// Cookie should contain a valid token
			
			// Token should be valid for authentication
			authReq := httptest.NewRequest("GET", "/", nil)
			authReq.AddCookie(cookie)
			
			authenticatedUser := controller.GetAuthenticatedUser(authReq)
		})
	}
}

func TestHelperMethodsWithInvalidTokens(t *testing.T) {
	controller, _ := setupTestController(t)
	
	invalidTokens := []string{
		"invalid-token",
		"",
		"malformed.jwt.token",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
	}
	
	for i, invalidToken := range invalidTokens {
		t.Run("invalid token "+string(rune('A'+i)), func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			
			if invalidToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: invalidToken,
				})
			}
			
			// All helper methods should handle invalid tokens gracefully
		})
	}
}

func TestHelperMethodsWithExpiredTokens(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create user and session
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	// Create an expired token manually
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalid"
	
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  controller.cookieName,
		Value: expiredToken,
	})
	
	// All helper methods should handle expired tokens gracefully
}

func TestHelperMethodsIntegration(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create regular user
	regularUser, err := controller.Signup("Regular User", "regular@example.com", "regular", "password", false)
	
	// Create admin user  
	adminUser, err := controller.Signup("Admin User", "admin@example.com", "admin", "password", true)
	
	// Test login and helper methods for regular user
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/login", nil)
	
	err = controller.Login(w1, req1, regularUser)
	
	regularCookie := w1.Result().Cookies()[0]
	
	// Test with regular user cookie
	authReq1 := httptest.NewRequest("GET", "/", nil)
	authReq1.AddCookie(regularCookie)
	
	authenticatedUser := controller.GetAuthenticatedUser(authReq1)
	
	
	// Test login and helper methods for admin user
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/login", nil)
	
	err = controller.Login(w2, req2, adminUser)
	
	adminCookie := w2.Result().Cookies()[0]
	
	// Test with admin user cookie
	authReq2 := httptest.NewRequest("GET", "/", nil)
	authReq2.AddCookie(adminCookie)
	
	authenticatedAdminUser := controller.GetAuthenticatedUser(authReq2)
	
	
	// Test logout
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/logout", nil)
	req3.AddCookie(regularCookie)
	
	controller.Logout(w3, req3)
	
	logoutCookie := w3.Result().Cookies()[0]
	
	// After logout, authentication should fail
	authReq3 := httptest.NewRequest("GET", "/", nil)
	authReq3.AddCookie(logoutCookie)
	
}

func TestCookieSecurityAttributes(t *testing.T) {
	controller, _ := setupTestController(t)
	
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	tests := []struct {
		name       string
		isHTTPS    bool
		cookieFunc func(http.ResponseWriter, *http.Request)
		checkFunc  func(*testing.T, *http.Cookie)
	}{
		{
			name:    "login cookie HTTP",
			isHTTPS: false,
			cookieFunc: func(w http.ResponseWriter, r *http.Request) {
				controller.Login(w, r, user)
			},
			checkFunc: func(t *testing.T, cookie *http.Cookie) {
			},
		},
		{
			name:    "session cookie HTTPS",
			isHTTPS: true,
			cookieFunc: func(w http.ResponseWriter, r *http.Request) {
				controller.SetSessionCookie(w, r, session)
			},
			checkFunc: func(t *testing.T, cookie *http.Cookie) {
			},
		},
		{
			name:    "logout cookie",
			isHTTPS: false,
			cookieFunc: func(w http.ResponseWriter, r *http.Request) {
				controller.Logout(w, r)
			},
			checkFunc: func(t *testing.T, cookie *http.Cookie) {
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", nil)
			
			if tt.isHTTPS {
				req.TLS = &struct{}{}
			}
			
			tt.cookieFunc(w, req)
			
			cookies := w.Result().Cookies()
			
			cookie := cookies[0]
			
			tt.checkFunc(t, cookie)
		})
	}
}