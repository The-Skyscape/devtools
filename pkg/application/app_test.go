package application_test

import (
	"bytes"
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/application"
)

//go:embed testdata/views/*.html
var testViews embed.FS

// TestAppLifecycle tests the complete application lifecycle
func TestAppLifecycle(t *testing.T) {
	// Create app with test views
	app := application.New(testViews)
	if app == nil {
		t.Fatal("Expected app to be created")
	}

	// Get server configuration
	addr, handler := app.Server()
	if addr == "" {
		t.Error("Expected server address")
	}
	if handler == nil {
		t.Error("Expected HTTP handler")
	}
}

// TestAppWithController tests controller registration
func TestAppWithController(t *testing.T) {
	// Create a test controller
	ctrl := &testController{}
	
	app := application.New(testViews,
		application.WithController("test", ctrl),
	)
	
	// Verify controller is registered
	if app.Use("test") != ctrl {
		t.Error("Controller not properly registered")
	}
	
	// Verify unknown controller returns nil
	if app.Use("unknown") != nil {
		t.Error("Expected nil for unknown controller")
	}
}

// TestAppRender tests template rendering
func TestAppRender(t *testing.T) {
	app := application.New(testViews)
	
	// Create a buffer to capture output
	var buf bytes.Buffer
	req := httptest.NewRequest("GET", "/", nil)
	
	// This should render without error
	app.Render(&buf, req, "test.html", map[string]string{
		"Title": "Test Page",
	})
	
	output := buf.String()
	if output == "" {
		t.Error("Expected rendered output")
	}
}

// TestAppProtect tests access control middleware
func TestAppProtect(t *testing.T) {
	app := application.New(testViews)
	
	// Create a handler that should be protected
	protectedCalled := false
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		protectedCalled = true
		w.WriteHeader(http.StatusOK)
	})
	
	// Test with no access check (should allow)
	handler := app.Protect(protected, nil)
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	if !protectedCalled {
		t.Error("Protected handler should have been called")
	}
	
	// Test with denying access check
	protectedCalled = false
	denyAccess := func(app *application.App, w http.ResponseWriter, r *http.Request) bool {
		w.WriteHeader(http.StatusForbidden)
		return false
	}
	
	handler = app.Protect(protected, denyAccess)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	
	if protectedCalled {
		t.Error("Protected handler should not have been called")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", rec.Code)
	}
}

// TestAppMiddleware tests middleware chaining
func TestAppMiddleware(t *testing.T) {
	callOrder := []string{}
	
	// Create test middleware
	middleware1 := &testMiddleware{
		name: "first",
		callOrder: &callOrder,
	}
	middleware2 := &testMiddleware{
		name: "second",
		callOrder: &callOrder,
	}
	
	app := application.New(testViews,
		application.WithMiddleware(middleware1),
		application.WithMiddleware(middleware2),
	)
	
	// Get the server handler which applies middleware
	_, handler := app.Server()
	
	// Make a request
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	
	// Verify middleware executed in correct order
	if len(callOrder) != 2 {
		t.Fatalf("Expected 2 middleware calls, got %d", len(callOrder))
	}
	if callOrder[0] != "first" {
		t.Errorf("Expected first middleware to run first, got %s", callOrder[0])
	}
	if callOrder[1] != "second" {
		t.Errorf("Expected second middleware to run second, got %s", callOrder[1])
	}
}

// BenchmarkAppRender benchmarks template rendering
func BenchmarkAppRender(b *testing.B) {
	app := application.New(testViews)
	req := httptest.NewRequest("GET", "/", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		app.Render(&buf, req, "test.html", map[string]string{
			"Title": "Benchmark",
		})
	}
}

// BenchmarkAppWithMiddleware benchmarks request handling with middleware
func BenchmarkAppWithMiddleware(b *testing.B) {
	app := application.New(testViews,
		application.WithMiddleware(&passthroughMiddleware{}),
		application.WithMiddleware(&passthroughMiddleware{}),
		application.WithMiddleware(&passthroughMiddleware{}),
	)
	
	_, handler := app.Server()
	req := httptest.NewRequest("GET", "/", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	app := application.New(testViews)
	_, handler := app.Server()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}
	})
}

// Helper types for testing

type testController struct {
	application.Controller
}

func (c *testController) Setup(app *application.App) {
	c.Controller.Setup(app)
}

func (c testController) Handle(r *http.Request) application.Handler {
	c.Request = r
	return &c
}

type testMiddleware struct {
	name      string
	callOrder *[]string
}

func (m *testMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*m.callOrder = append(*m.callOrder, m.name)
		next.ServeHTTP(w, r)
	})
}

type passthroughMiddleware struct{}

func (m *passthroughMiddleware) Handle(next http.Handler) http.Handler {
	return next
}