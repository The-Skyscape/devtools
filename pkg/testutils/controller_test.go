package testutils

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/application"
)

//go:embed testdata/views/*.html
var testViews embed.FS

// Example test using TestController
func TestExampleController(t *testing.T) {
	// Setup test database
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create application with embedded test views
	app := application.New(testViews)
	
	// Define test cases
	tests := []ControllerTest{
		{
			Name:           "GET home page",
			Method:         "GET",
			Path:           "/",
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   "Welcome",
		},
		{
			Name:           "POST form data",
			Method:         "POST",
			Path:           "/submit",
			Body:           url.Values{"name": []string{"John"}, "email": []string{"john@example.com"}},
			ExpectedStatus: http.StatusOK,
			CheckFunc: func(t *testing.T, w *MockResponse) {
				// Custom assertions
				w.AssertBodyContains(t, "Thank you, John")
			},
		},
		{
			Name:           "JSON API endpoint",
			Method:         "POST",
			Path:           "/api/users",
			Body:           map[string]string{"name": "Jane", "email": "jane@example.com"},
			Headers:        map[string]string{"Content-Type": "application/json"},
			ExpectedStatus: http.StatusCreated,
			ExpectedHeader: map[string]string{"Content-Type": "application/json"},
			CheckFunc: func(t *testing.T, w *MockResponse) {
				var response map[string]any
				w.AssertJSON(t, &response)
				AssertEqual(t, "Jane", response["name"])
			},
		},
		{
			Name:           "Redirect after POST",
			Method:         "POST",
			Path:           "/login",
			Body:           url.Values{"username": []string{"admin"}, "password": []string{"secret"}},
			ExpectedStatus: http.StatusFound,
			CheckFunc: func(t *testing.T, w *MockResponse) {
				w.AssertRedirect(t, "/dashboard")
			},
		},
	}

	// Run tests
	TestController(t, app, tests)
}

// Example test using ControllerTestSuite for fluent API
func TestExampleControllerSuite(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	app := application.New(testViews)
	suite := NewControllerTestSuite(t, app)

	// Test GET request
	suite.
		GET("/about").
		AssertOK().
		AssertContains("About Us").
		AssertNotContains("Error")

	// Test POST with form data
	suite.
		POST("/contact", url.Values{
			"name":    []string{"Alice"},
			"message": []string{"Hello"},
		}).
		AssertOK().
		AssertContains("Message sent")

	// Test with headers
	suite.
		GET("/api/profile").
		WithHeader("Authorization", "Bearer token123").
		AssertOK()

	// Test redirect
	suite.
		POST("/logout", nil).
		AssertRedirect("/")
}

// Example test for testing a handler directly
func TestHandlerDirectly(t *testing.T) {
	// Define a simple handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "Name required", http.StatusBadRequest)
			return
		}
		w.Write([]byte("Hello, " + name))
	}

	// Test success case
	TestControllerMethod(t, handler, "GET", "/?name=Bob", nil,
		func(t *testing.T, w *MockResponse) {
			w.AssertOK(t)
			w.AssertBodyContains(t, "Hello, Bob")
		})

	// Test error case
	TestControllerMethod(t, handler, "GET", "/", nil,
		func(t *testing.T, w *MockResponse) {
			w.AssertStatus(t, http.StatusBadRequest)
			w.AssertBodyContains(t, "Name required")
		})
}

// Example test with authentication
func TestAuthenticatedEndpoint(t *testing.T) {
	app := application.New(testViews)
	
	// Create a test JWT token (in real tests, use your auth system)
	token := "test-jwt-token"
	
	// Create authenticated request
	req := MockAuthenticatedRequest("GET", "/profile", nil, token)
	w := NewMockResponse()
	
	// Serve request
	_, handler := app.Server()
	handler.ServeHTTP(w, req)
	
	// Assert success
	w.AssertOK(t)
	w.AssertBodyContains(t, "Profile")
}

// Example test for HTMX responses
func TestHTMXEndpoint(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Typical HTMX response with refresh
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
	}

	req := MockRequest("POST", "/update", nil)
	req.Header.Set("HX-Request", "true") // HTMX request header
	
	w := httptest.NewRecorder()
	handler(w, req)
	
	// Check HTMX refresh header
	AssertHTMXRefresh(t, w)
}

// Example test with setup and cleanup
func TestWithSetupAndCleanup(t *testing.T) {
	app := application.New(testViews)
	
	var testData string
	
	tests := []ControllerTest{
		{
			Name:   "Test with setup",
			Method: "GET",
			Path:   "/data",
			SetupFunc: func() {
				// Setup test data
				testData = "test value"
			},
			CleanupFunc: func() {
				// Cleanup test data
				testData = ""
			},
			CheckFunc: func(t *testing.T, w *MockResponse) {
				// Verify setup ran
				AssertEqual(t, "test value", testData)
			},
		},
	}
	
	TestController(t, app, tests)
}