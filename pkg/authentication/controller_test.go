package authentication

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
)

func setupTestController(t *testing.T) (*Controller, *application.App) {
	t.Helper()
	
	// Create test collection
	collection, _ := setupTestCollection(t)
	
	// Create controller with default options
	controller := collection.Controller()
	
	// Create test app
	app := application.New()
	
	// Setup controller with app
	controller.Setup(app)
	
	return controller, app
}

func TestControllerNew(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
	tests := []struct {
		name     string
		options  []Option
		validate func(*testing.T, *Controller)
	}{
		{
			name:    "default options",
			options: []Option{},
			validate: func(t *testing.T, ctrl *Controller) {
			},
		},
		{
			name: "with custom cookie name",
			options: []Option{
				WithCookie("custom-cookie"),
			},
			validate: func(t *testing.T, ctrl *Controller) {
			},
		},
		{
			name: "with custom setup view",
			options: []Option{
				WithSetupView("custom-setup.html", "/setup-redirect"),
			},
			validate: func(t *testing.T, ctrl *Controller) {
			},
		},
		{
			name: "with custom signin view",
			options: []Option{
				WithSigninView("custom-signin.html", "/signin-redirect"),
			},
			validate: func(t *testing.T, ctrl *Controller) {
			},
		},
		{
			name: "with custom signout URL",
			options: []Option{
				WithSignoutURL("/custom-signout"),
			},
			validate: func(t *testing.T, ctrl *Controller) {
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := collection.Controller(tt.options...)
			tt.validate(t, controller)
		})
	}
}

func TestOptional(t *testing.T) {
	controller, app := setupTestController(t)
	
	// Optional should always return true
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	result := controller.Optional(app, w, req)
}

func TestRequired(t *testing.T) {
	controller, app := setupTestController(t)
	
	// Mock the render function to track calls
	renderCalled := false
	var renderTemplate string
	app.Render = func(w http.ResponseWriter, r *http.Request, template string, data any) {
		renderCalled = true
		renderTemplate = template
	}
	
	tests := []struct {
		name            string
		setupUser       bool
		setupAuth       bool
		expectedResult  bool
		expectedRender  string
	}{
		{
			name:            "no users exist - should show signup",
			setupUser:       false,
			setupAuth:       false,
			expectedResult:  false,
			expectedRender:  "signup.html",
		},
		{
			name:            "user exists but not authenticated - should show signin",
			setupUser:       true,
			setupAuth:       false,
			expectedResult:  false,
			expectedRender:  "signin.html",
		},
		{
			name:            "user exists and authenticated - should pass",
			setupUser:       true,
			setupAuth:       true,
			expectedResult:  true,
			expectedRender:  "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			renderCalled = false
			renderTemplate = ""
			
			// Setup user if needed
			var user *User
			if tt.setupUser {
				var err error
				user, err = controller.Signup("Test User", "test@example.com", "testuser", "password", false)
			}
			
			// Setup request
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			
			// Setup authentication if needed
			if tt.setupAuth && user != nil {
				session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
				
				token, err := session.Token()
				
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: token,
				})
			}
			
			result := controller.Required(app, w, req)
			
			
			if tt.expectedRender != "" {
			} else {
			}
		})
	}
}

func TestAdminOnly(t *testing.T) {
	controller, app := setupTestController(t)
	
	// Mock the render function
	renderCalled := false
	var renderTemplate string
	app.Render = func(w http.ResponseWriter, r *http.Request, template string, data any) {
		renderCalled = true
		renderTemplate = template
	}
	
	// Create regular user and admin user
	regularUser, err := controller.Signup("Regular User", "user@example.com", "user", "password", false)
	
	adminUser, err := controller.Signup("Admin User", "admin@example.com", "admin", "password", true)
	
	tests := []struct {
		name            string
		user            *User
		expectedResult  bool
		expectedRender  string
	}{
		{
			name:            "no authentication - should show signin",
			user:            nil,
			expectedResult:  false,
			expectedRender:  "signin.html",
		},
		{
			name:            "regular user - should show signin",
			user:            regularUser,
			expectedResult:  false,
			expectedRender:  "signin.html",
		},
		{
			name:            "admin user - should pass",
			user:            adminUser,
			expectedResult:  true,
			expectedRender:  "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			renderCalled = false
			renderTemplate = ""
			
			// Setup request
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			
			// Setup authentication if user provided
			if tt.user != nil {
				session, err := controller.Sessions.Insert(&Session{UserID: tt.user.ID})
				
				token, err := session.Token()
				
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: token,
				})
			}
			
			result := controller.AdminOnly(app, w, req)
			
			
			if tt.expectedRender != "" {
			} else {
			}
		})
	}
}

func TestCurrentUser(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	// Create session
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	token, err := session.Token()
	
	tests := []struct {
		name           string
		setupCookie    bool
		expectedResult bool
	}{
		{
			name:           "with valid authentication",
			setupCookie:    true,
			expectedResult: true,
		},
		{
			name:           "without authentication",
			setupCookie:    false,
			expectedResult: false,
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
			
			// Create controller instance with request
			controllerHandler := controller.Handle(req)
			ctrl := controllerHandler.(*Controller)
			
			currentUser := ctrl.CurrentUser()
			
			if tt.expectedResult {
			} else {
			}
		})
	}
}

func TestCurrentSession(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	// Create session
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	token, err := session.Token()
	
	tests := []struct {
		name           string
		setupCookie    bool
		expectedResult bool
	}{
		{
			name:           "with valid authentication",
			setupCookie:    true,
			expectedResult: true,
		},
		{
			name:           "without authentication",
			setupCookie:    false,
			expectedResult: false,
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
			
			// Create controller instance with request
			controllerHandler := controller.Handle(req)
			ctrl := controllerHandler.(*Controller)
			
			currentSession := ctrl.CurrentSession()
			
			if tt.expectedResult {
			} else {
			}
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	// Create session
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	token, err := session.Token()
	
	tests := []struct {
		name        string
		setupCookie bool
		expected    bool
	}{
		{
			name:        "authenticated user",
			setupCookie: true,
			expected:    true,
		},
		{
			name:        "unauthenticated user",
			setupCookie: false,
			expected:    false,
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
			
			// Create controller instance with request
			controllerHandler := controller.Handle(req)
			ctrl := controllerHandler.(*Controller)
			
			result := ctrl.IsAuthenticated()
		})
	}
}

func TestHandleSignup(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Mock render function to capture errors
	var renderTemplate string
	var renderData any
	renderCalled := false
	controller.App.Render = func(w http.ResponseWriter, r *http.Request, template string, data any) {
		renderCalled = true
		renderTemplate = template
		renderData = data
	}
	
	tests := []struct {
		name          string
		formData      map[string]string
		expectError   bool
		expectCookie  bool
		isFirstUser   bool
	}{
		{
			name: "valid signup - first user becomes admin",
			formData: map[string]string{
				"name":     "First User",
				"handle":   "first",
				"email":    "first@example.com",
				"password": "password123",
			},
			expectError:  false,
			expectCookie: true,
			isFirstUser:  true,
		},
		{
			name: "valid signup - subsequent user",
			formData: map[string]string{
				"name":     "Regular User",
				"handle":   "regular",
				"email":    "regular@example.com",
				"password": "password123",
			},
			expectError:  false,
			expectCookie: true,
			isFirstUser:  false,
		},
		{
			name: "missing name",
			formData: map[string]string{
				"handle":   "test",
				"email":    "test@example.com",
				"password": "password123",
			},
			expectError:  true,
			expectCookie: false,
		},
		{
			name: "missing handle",
			formData: map[string]string{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "password123",
			},
			expectError:  true,
			expectCookie: false,
		},
		{
			name: "missing email",
			formData: map[string]string{
				"name":     "Test User",
				"handle":   "test",
				"password": "password123",
			},
			expectError:  true,
			expectCookie: false,
		},
		{
			name: "missing password",
			formData: map[string]string{
				"name":   "Test User",
				"handle": "test",
				"email":  "test@example.com",
			},
			expectError:  true,
			expectCookie: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			renderCalled = false
			renderTemplate = ""
			renderData = nil
			
			// Create form data
			formData := url.Values{}
			for key, value := range tt.formData {
				formData.Set(key, value)
			}
			
			// Create request
			req := httptest.NewRequest("POST", "/_auth/signup", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			
			// Call handler
			controller.HandleSignup(w, req)
			
			if tt.expectError {
			} else {
				if renderCalled {
					// If render was called, it should be a redirect or refresh
				}
			}
			
			// Check for cookie
			cookies := w.Result().Cookies()
			if tt.expectCookie {
				cookie := cookies[0]
			} else {
			}
			
			// If successful, verify user was created
			if !tt.expectError {
				user, err := controller.GetUser(tt.formData["email"])
				
				// Check admin status for first user
			}
		})
	}
}

func TestHandleSignin(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user
	password := "testpassword"
	user, err := controller.Signup("Test User", "test@example.com", "testuser", password, false)
	
	// Mock render function
	var renderTemplate string
	var renderData any
	renderCalled := false
	controller.App.Render = func(w http.ResponseWriter, r *http.Request, template string, data any) {
		renderCalled = true
		renderTemplate = template
		renderData = data
	}
	
	tests := []struct {
		name         string
		handle       string
		password     string
		expectError  bool
		expectCookie bool
	}{
		{
			name:         "valid signin with handle",
			handle:       "testuser",
			password:     password,
			expectError:  false,
			expectCookie: true,
		},
		{
			name:         "valid signin with email",
			handle:       "test@example.com",
			password:     password,
			expectError:  false,
			expectCookie: true,
		},
		{
			name:         "invalid password",
			handle:       "testuser",
			password:     "wrongpassword",
			expectError:  true,
			expectCookie: false,
		},
		{
			name:         "nonexistent user",
			handle:       "nonexistent",
			password:     password,
			expectError:  true,
			expectCookie: false,
		},
		{
			name:         "empty handle",
			handle:       "",
			password:     password,
			expectError:  true,
			expectCookie: false,
		},
		{
			name:         "empty password",
			handle:       "testuser",
			password:     "",
			expectError:  true,
			expectCookie: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			renderCalled = false
			renderTemplate = ""
			renderData = nil
			
			// Create form data
			formData := url.Values{}
			formData.Set("handle", tt.handle)
			formData.Set("password", tt.password)
			
			// Create request
			req := httptest.NewRequest("POST", "/_auth/signin", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			
			// Call handler
			controller.HandleSignin(w, req)
			
			if tt.expectError {
			}
			
			// Check for cookie
			cookies := w.Result().Cookies()
			if tt.expectCookie {
				cookie := cookies[0]
				
				// Verify cookie expires in ~4 days
				expectedExpiry := time.Now().Add(time.Hour * 24 * 4)
			} else {
			}
		})
	}
}

func TestHandleSignout(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Create test user and session
	user, err := controller.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	session, err := controller.Sessions.Insert(&Session{UserID: user.ID})
	
	token, err := session.Token()
	
	// Mock redirect function
	redirectCalled := false
	var redirectURL string
	controller.Redirect = func(w http.ResponseWriter, r *http.Request, url string) {
		redirectCalled = true
		redirectURL = url
	}
	
	tests := []struct {
		name        string
		setupCookie bool
	}{
		{
			name:        "signout with session",
			setupCookie: true,
		},
		{
			name:        "signout without session",
			setupCookie: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			redirectCalled = false
			redirectURL = ""
			
			// Create request
			req := httptest.NewRequest("POST", "/_auth/signout", nil)
			w := httptest.NewRecorder()
			
			if tt.setupCookie {
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: token,
				})
			}
			
			// Call handler
			controller.HandleSignout(w, req)
			
			// Should always redirect to signout URL
			
			// Should always clear cookie
			cookies := w.Result().Cookies()
			
			cookie := cookies[0]
		})
	}
}

func TestControllerIntegration(t *testing.T) {
	controller, _ := setupTestController(t)
	
	// Test full signup -> signin -> signout flow
	
	// 1. Signup
	formData := url.Values{}
	formData.Set("name", "Integration User")
	formData.Set("handle", "integration")
	formData.Set("email", "integration@example.com")
	formData.Set("password", "integrationpass")
	
	req := httptest.NewRequest("POST", "/_auth/signup", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	
	controller.HandleSignup(w, req)
	
	// Should have cookie
	signupCookies := w.Result().Cookies()
	signupCookie := signupCookies[0]
	
	// 2. Use cookie to test authentication
	authReq := httptest.NewRequest("GET", "/", nil)
	authReq.AddCookie(signupCookie)
	
	authController := controller.Handle(authReq).(*Controller)
	
	currentUser := authController.CurrentUser()
	
	// 3. Signout
	signoutReq := httptest.NewRequest("POST", "/_auth/signout", nil)
	signoutReq.AddCookie(signupCookie)
	signoutW := httptest.NewRecorder()
	
	redirectCalled := false
	controller.Redirect = func(w http.ResponseWriter, r *http.Request, url string) {
		redirectCalled = true
	}
	
	controller.HandleSignout(signoutW, signoutReq)
	
	
	// Should have clearing cookie
	signoutCookies := signoutW.Result().Cookies()
	clearCookie := signoutCookies[0]
	
	// 4. Test that user is no longer authenticated with cleared cookie
	testReq := httptest.NewRequest("GET", "/", nil)
	testReq.AddCookie(clearCookie)
	
	testController := controller.Handle(testReq).(*Controller)
}