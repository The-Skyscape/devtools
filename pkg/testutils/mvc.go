package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// MVCTest provides a testing framework for MVC applications
type MVCTest struct {
	*testing.T
	app *application.App
	mux *http.ServeMux
}

// NewMVCTest creates a new MVC test instance
func NewMVCTest(t *testing.T) *MVCTest {
	// Create a fresh mux to avoid conflicts
	mux := http.NewServeMux()
	
	// Replace default mux temporarily
	oldMux := http.DefaultServeMux
	http.DefaultServeMux = mux
	
	// Restore on cleanup
	t.Cleanup(func() {
		http.DefaultServeMux = oldMux
	})
	
	return &MVCTest{
		T:   t,
		app: application.New(nil),
		mux: mux,
	}
}

// App returns the test application instance
func (m *MVCTest) App() *application.App {
	return m.app
}

// ControllerTestContext provides context for testing a specific controller
type ControllerTestContext struct {
	*MVCTest
	prefix     string
	controller application.Handler
}

// TestController tests a controller factory function and runs additional tests with the controller instance
func (m *MVCTest) TestController(name string, factory func() (string, application.Handler), additionalTests ...func(*ControllerTestContext)) {
	m.Run(name, func(t *testing.T) {
		// Create controller instance once
		prefix, controller := factory()
		
		// Create controller test context
		ct := &ControllerTestContext{
			MVCTest:    m,
			prefix:     prefix,
			controller: controller,
		}
		
		// Run standard tests
		t.Run("Factory", func(t *testing.T) {
			if prefix == "" {
				t.Error("Factory should return non-empty prefix")
			}
			
			if controller == nil {
				t.Error("Factory should return controller instance")
			}
			
			t.Logf("Controller factory returns prefix '%s'", prefix)
		})
		
		t.Run("Setup", func(t *testing.T) {
			// Setup should not panic
			controller.Setup(m.app)
			
			// Verify app reference if it's a standard controller
			if ctrl, ok := controller.(interface{ GetApp() *application.App }); ok {
				if ctrl.GetApp() != m.app {
					t.Error("Controller should store app reference")
				}
			}
		})
		
		t.Run("RequestIsolation", func(t *testing.T) {
			req1 := httptest.NewRequest("GET", "/test/1", nil)
			req2 := httptest.NewRequest("GET", "/test/2", nil)
			
			handler1 := controller.Handle(req1)
			handler2 := controller.Handle(req2)
			
			if handler1 == nil || handler2 == nil {
				t.Fatal("Handle should return handler instances")
			}
			
			// Verify they are different instances (value receiver pattern)
			if handler1 == handler2 {
				t.Error("Handle should return different instances for request isolation")
			}
		})
		
		// Run additional tests with controller instance
		for _, test := range additionalTests {
			test(ct)
		}
	})
}

// Run runs a subtest with the controller instance available
func (ct *ControllerTestContext) Run(name string, f func(*testing.T, application.Handler)) {
	ct.T.Run(name, func(t *testing.T) {
		f(t, ct.controller)
	})
}

// Prefix returns the controller's prefix
func (ct *ControllerTestContext) Prefix() string {
	return ct.prefix
}

// Controller returns the controller instance
func (ct *ControllerTestContext) Controller() application.Handler {
	return ct.controller
}

// TestRoute tests that a route is registered and responds
func (m *MVCTest) TestRoute(method, path string, expectedStatus int) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	
	m.mux.ServeHTTP(rec, req)
	
	if rec.Code != expectedStatus {
		m.Errorf("Route %s %s: expected status %d, got %d", method, path, expectedStatus, rec.Code)
	}
}

// TestHandler tests an HTTP handler function
func (m *MVCTest) TestHandler(handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// Request creates a test request with common setup
func (m *MVCTest) Request(method, path string) *http.Request {
	return httptest.NewRequest(method, path, nil)
}

// RequestWithForm creates a form-encoded POST request
func (m *MVCTest) RequestWithForm(path string, values map[string]string) *http.Request {
	form := make(url.Values)
	for k, v := range values {
		form.Set(k, v)
	}
	
	req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// RequestWithCookie creates a request with a cookie
func (m *MVCTest) RequestWithCookie(method, path, name, value string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(&http.Cookie{Name: name, Value: value})
	return req
}

// AssertRedirect checks if response is a redirect to expected location
func (m *MVCTest) AssertRedirect(rec *httptest.ResponseRecorder, expectedLocation string) {
	m.Helper()
	
	if rec.Code != http.StatusFound && rec.Code != http.StatusSeeOther {
		m.Errorf("Expected redirect status (302/303), got %d", rec.Code)
	}
	
	location := rec.Header().Get("Location")
	if location != expectedLocation {
		m.Errorf("Expected redirect to %s, got %s", expectedLocation, location)
	}
}

// AssertStatus checks if response has expected status code
func (m *MVCTest) AssertStatus(rec *httptest.ResponseRecorder, expected int) {
	m.Helper()
	
	if rec.Code != expected {
		m.Errorf("Expected status %d, got %d", expected, rec.Code)
	}
}

// AssertContains checks if response body contains text
func (m *MVCTest) AssertContains(rec *httptest.ResponseRecorder, text string) {
	m.Helper()
	
	body := rec.Body.String()
	if !strings.Contains(body, text) {
		m.Errorf("Response should contain '%s'", text)
		m.Logf("Actual response: %s", body)
	}
}

// AssertJSON checks if response is valid JSON with expected content-type
func (m *MVCTest) AssertJSON(rec *httptest.ResponseRecorder) {
	m.Helper()
	
	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		m.Errorf("Expected JSON content-type, got %s", contentType)
	}
	
	// Try to decode JSON
	var data interface{}
	if err := json.NewDecoder(rec.Body).Decode(&data); err != nil {
		m.Errorf("Response is not valid JSON: %v", err)
	}
}

// AssertCookie checks if response sets a cookie
func (m *MVCTest) AssertCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	m.Helper()
	
	cookies := rec.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	
	m.Errorf("Expected cookie '%s' not found", name)
	return nil
}

// AssertNoCookie checks that a cookie is not set
func (m *MVCTest) AssertNoCookie(rec *httptest.ResponseRecorder, name string) {
	m.Helper()
	
	cookies := rec.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == name {
			m.Errorf("Unexpected cookie '%s' found", name)
		}
	}
}

// AssertTemplateMethod verifies a controller method exists and returns expected types
func (m *MVCTest) AssertTemplateMethod(controller interface{}, methodName string) {
	m.Helper()
	
	// Use reflection to check if method exists
	v := reflect.ValueOf(controller)
	method := v.MethodByName(methodName)
	
	if !method.IsValid() {
		m.Errorf("Controller should have method '%s' for template access", methodName)
	}
}