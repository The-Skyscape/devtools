package testutils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// ControllerTest defines a test case for a controller
type ControllerTest struct {
	Name           string                          // Test name
	Method         string                          // HTTP method
	Path           string                          // Request path
	Body           any                             // Request body (string, []byte, url.Values, or struct for JSON)
	Headers        map[string]string               // Request headers
	SetupFunc      func()                          // Optional setup function
	CleanupFunc    func()                          // Optional cleanup function
	CheckFunc      func(t *testing.T, w *MockResponse) // Custom check function
	ExpectedStatus int                             // Expected HTTP status code
	ExpectedBody   string                          // Expected body substring
	ExpectedHeader map[string]string               // Expected response headers
}

// TestController runs a series of controller tests
func TestController(t *testing.T, app *application.App, tests []ControllerTest) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Run setup if provided
			if tt.SetupFunc != nil {
				tt.SetupFunc()
			}

			// Run cleanup after test
			if tt.CleanupFunc != nil {
				defer tt.CleanupFunc()
			}

			// Create request
			req := createTestRequest(tt.Method, tt.Path, tt.Body)
			
			// Add headers
			for k, v := range tt.Headers {
				req.Header.Set(k, v)
			}

			// Create response recorder
			w := NewMockResponse()

			// Serve the request through the application
			_, handler := app.Server()
			handler.ServeHTTP(w, req)

			// Check status code if specified
			if tt.ExpectedStatus != 0 {
				w.AssertStatus(t, tt.ExpectedStatus)
			}

			// Check body contains expected string if specified
			if tt.ExpectedBody != "" {
				w.AssertBodyContains(t, tt.ExpectedBody)
			}

			// Check headers if specified
			for k, v := range tt.ExpectedHeader {
				w.AssertHeader(t, k, v)
			}

			// Run custom check function if provided
			if tt.CheckFunc != nil {
				tt.CheckFunc(t, w)
			}
		})
	}
}

// createTestRequest creates a test request based on the body type
func createTestRequest(method, path string, body any) *http.Request {
	switch v := body.(type) {
	case nil:
		return MockRequest(method, path, nil)
	case string:
		return MockRequest(method, path, strings.NewReader(v))
	case []byte:
		return MockRequest(method, path, bytes.NewReader(v))
	case url.Values:
		return MockFormRequest(method, path, v)
	default:
		// Assume it's a struct to be marshaled as JSON
		return MockJSONRequest(method, path, v)
	}
}

// ControllerTestSuite provides a test suite for controllers
type ControllerTestSuite struct {
	t        *testing.T
	app      *application.App
	recorder *MockResponse
	request  *http.Request
}

// NewControllerTestSuite creates a new controller test suite
func NewControllerTestSuite(t *testing.T, app *application.App) *ControllerTestSuite {
	return &ControllerTestSuite{
		t:   t,
		app: app,
	}
}

// GET performs a GET request
func (s *ControllerTestSuite) GET(path string) *ControllerTestSuite {
	s.request = MockRequest("GET", path, nil)
	s.recorder = NewMockResponse()
	_, handler := s.app.Server()
	handler.ServeHTTP(s.recorder, s.request)
	return s
}

// POST performs a POST request with form data
func (s *ControllerTestSuite) POST(path string, values url.Values) *ControllerTestSuite {
	s.request = MockFormRequest("POST", path, values)
	s.recorder = NewMockResponse()
	_, handler := s.app.Server()
	handler.ServeHTTP(s.recorder, s.request)
	return s
}

// POSTJson performs a POST request with JSON data
func (s *ControllerTestSuite) POSTJson(path string, body any) *ControllerTestSuite {
	s.request = MockJSONRequest("POST", path, body)
	s.recorder = NewMockResponse()
	_, handler := s.app.Server()
	handler.ServeHTTP(s.recorder, s.request)
	return s
}

// WithHeader adds a header to the request
func (s *ControllerTestSuite) WithHeader(key, value string) *ControllerTestSuite {
	if s.request != nil {
		s.request.Header.Set(key, value)
	}
	return s
}

// WithCookie adds a cookie to the request
func (s *ControllerTestSuite) WithCookie(cookie *http.Cookie) *ControllerTestSuite {
	if s.request != nil {
		s.request.AddCookie(cookie)
	}
	return s
}

// AssertOK checks if the response is 200 OK
func (s *ControllerTestSuite) AssertOK() *ControllerTestSuite {
	s.t.Helper()
	s.recorder.AssertOK(s.t)
	return s
}

// AssertStatus checks the response status code
func (s *ControllerTestSuite) AssertStatus(code int) *ControllerTestSuite {
	s.t.Helper()
	s.recorder.AssertStatus(s.t, code)
	return s
}

// AssertRedirect checks if the response is a redirect
func (s *ControllerTestSuite) AssertRedirect(location string) *ControllerTestSuite {
	s.t.Helper()
	s.recorder.AssertRedirect(s.t, location)
	return s
}

// AssertContains checks if the response body contains a string
func (s *ControllerTestSuite) AssertContains(text string) *ControllerTestSuite {
	s.t.Helper()
	s.recorder.AssertBodyContains(s.t, text)
	return s
}

// AssertNotContains checks if the response body does not contain a string
func (s *ControllerTestSuite) AssertNotContains(text string) *ControllerTestSuite {
	s.t.Helper()
	body := s.recorder.Body.String()
	if strings.Contains(body, text) {
		s.t.Errorf("Response body should not contain: %s", text)
	}
	return s
}

// AssertHeader checks a response header
func (s *ControllerTestSuite) AssertHeader(key, value string) *ControllerTestSuite {
	s.t.Helper()
	s.recorder.AssertHeader(s.t, key, value)
	return s
}

// GetResponse returns the response recorder for custom assertions
func (s *ControllerTestSuite) GetResponse() *MockResponse {
	return s.recorder
}

// GetRequest returns the request for inspection
func (s *ControllerTestSuite) GetRequest() *http.Request {
	return s.request
}

// MockAuthenticatedRequest creates a request with authentication headers/cookies
func MockAuthenticatedRequest(method, path string, body any, token string) *http.Request {
	req := createTestRequest(method, path, body)
	
	// Add JWT token as cookie (typical for devtools apps)
	cookie := &http.Cookie{
		Name:  "auth-token",
		Value: token,
	}
	req.AddCookie(cookie)
	
	// Also add as Authorization header for API-style requests
	req.Header.Set("Authorization", "Bearer "+token)
	
	return req
}

// AssertHTMXRefresh checks if the response has HX-Refresh header (for HTMX)
func AssertHTMXRefresh(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	hxRefresh := w.Header().Get("HX-Refresh")
	if hxRefresh != "true" {
		t.Error("Expected HX-Refresh header to be 'true'")
	}
}

// AssertHTMXRedirect checks if the response has HX-Redirect header (for HTMX)
func AssertHTMXRedirect(t *testing.T, w *httptest.ResponseRecorder, expectedURL string) {
	t.Helper()
	hxRedirect := w.Header().Get("HX-Redirect")
	if hxRedirect != expectedURL {
		t.Errorf("Expected HX-Redirect to %s, got %s", expectedURL, hxRedirect)
	}
}

// TestControllerMethod tests a single controller method directly
func TestControllerMethod(
	t *testing.T,
	handler http.HandlerFunc,
	method string,
	path string,
	body any,
	check func(t *testing.T, w *MockResponse),
) {
	t.Helper()
	
	// Create request
	req := createTestRequest(method, path, body)
	
	// Create response recorder
	w := NewMockResponse()
	
	// Call handler directly
	handler(w, req)
	
	// Run checks
	check(t, w)
}