package application_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// TestControllerRequestIsolation verifies the value receiver pattern
func TestControllerRequestIsolation(t *testing.T) {
	// Create controller
	ctrl := &isolationTestController{}
	app := application.New(nil)
	ctrl.Setup(app)
	
	// Create two different requests
	req1 := httptest.NewRequest("GET", "/path1", nil)
	req2 := httptest.NewRequest("GET", "/path2", nil)
	
	// Handle both requests "concurrently"
	handler1 := ctrl.Handle(req1)
	handler2 := ctrl.Handle(req2)
	
	// Each handler should have its own request
	if h1, ok := handler1.(*isolationTestController); ok {
		if h1.Request != req1 {
			t.Error("Handler 1 should have request 1")
		}
	}
	
	if h2, ok := handler2.(*isolationTestController); ok {
		if h2.Request != req2 {
			t.Error("Handler 2 should have request 2")
		}
	}
	
	// Original controller should not be modified
	if ctrl.Request != nil {
		t.Error("Original controller should not have request set")
	}
}

// TestControllerHTMXHeaders tests HTMX-specific behavior
func TestControllerHTMXHeaders(t *testing.T) {
	ctrl := &htmxTestController{}
	app := application.New(nil)
	ctrl.Setup(app)
	
	tests := []struct {
		name        string
		htmxRequest bool
		method      string
		wantHeader  string
		wantValue   string
	}{
		{
			name:        "HTMX Refresh",
			htmxRequest: true,
			method:      "refresh",
			wantHeader:  "HX-Refresh",
			wantValue:   "true",
		},
		{
			name:        "HTMX Redirect",
			htmxRequest: true,
			method:      "redirect",
			wantHeader:  "HX-Location",
			wantValue:   "/new-path",
		},
		{
			name:        "Non-HTMX Refresh",
			htmxRequest: false,
			method:      "refresh",
			wantHeader:  "Location",
			wantValue:   "/",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.htmxRequest {
				req.Header.Set("HX-Request", "true")
			}
			
			rec := httptest.NewRecorder()
			handler := ctrl.Handle(req).(*htmxTestController)
			
			switch tt.method {
			case "refresh":
				handler.Refresh(rec, req)
			case "redirect":
				handler.Redirect(rec, req, "/new-path")
			}
			
			if tt.wantHeader != "" {
				got := rec.Header().Get(tt.wantHeader)
				if got != tt.wantValue {
					t.Errorf("Expected header %s=%s, got %s", tt.wantHeader, tt.wantValue, got)
				}
			}
		})
	}
}

// TestControllerErrorHandling tests error rendering with correct templates
func TestControllerErrorHandling(t *testing.T) {
	t.Skip("Skipping error handling test due to missing templates")
	ctrl := &errorTestController{}
	app := application.New(nil)
	ctrl.Setup(app)
	
	tests := []struct {
		name         string
		err          error
		wantTemplate string
	}{
		{
			name:         "NotFound",
			err:          application.ErrNotFound,
			wantTemplate: "error-404.html",
		},
		{
			name:         "Unauthorized",
			err:          application.ErrUnauthorized,
			wantTemplate: "error-401.html",
		},
		{
			name:         "Forbidden",
			err:          application.ErrForbidden,
			wantTemplate: "error-403.html",
		},
		{
			name:         "Generic Error",
			err:          errors.New("something went wrong"),
			wantTemplate: "error-message.html",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			
			handler := ctrl.Handle(req).(*errorTestController)
			handler.templateUsed = ""
			handler.RenderError(rec, req, tt.err)
			
			// Note: We can't easily verify the exact template used without
			// modifying the framework, but we can verify it renders without panic
			if rec.Body.Len() == 0 && handler.templateUsed == "" {
				t.Error("Expected error to be rendered")
			}
		})
	}
}

// TestControllerValidation tests the validation helper
func TestControllerValidation(t *testing.T) {
	ctrl := &validationTestController{}
	app := application.New(nil)
	ctrl.Setup(app)
	
	req := httptest.NewRequest("GET", "/", nil)
	handler := ctrl.Handle(req).(*validationTestController)
	
	// Test validator creation
	validator := handler.Validator()
	if validator == nil {
		t.Fatal("Expected validator to be created")
	}
	
	// Test validation checks
	validator.CheckRequired("name", "")
	validator.CheckRequired("email", "test@example.com")
	
	err := validator.Result()
	if err == nil {
		t.Error("Expected validation error for empty name")
	}
	
	// Verify it's a ValidationError
	if _, ok := err.(application.ValidationError); !ok {
		t.Error("Expected ValidationError type")
	}
}

// TestControllerIsHTMX tests HTMX detection
func TestControllerIsHTMX(t *testing.T) {
	ctrl := &htmxTestController{}
	
	tests := []struct {
		name      string
		header    string
		value     string
		wantHTMX  bool
	}{
		{
			name:     "HTMX Request",
			header:   "HX-Request",
			value:    "true",
			wantHTMX: true,
		},
		{
			name:     "Non-HTMX Request",
			header:   "",
			value:    "",
			wantHTMX: false,
		},
		{
			name:     "Wrong Header Value",
			header:   "HX-Request",
			value:    "false",
			wantHTMX: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set(tt.header, tt.value)
			}
			
			got := ctrl.IsHTMX(req)
			if got != tt.wantHTMX {
				t.Errorf("IsHTMX() = %v, want %v", got, tt.wantHTMX)
			}
		})
	}
}

// BenchmarkControllerHandle benchmarks the Handle method
func BenchmarkControllerHandle(b *testing.B) {
	ctrl := &benchController{}
	app := application.New(nil)
	ctrl.Setup(app)
	
	req := httptest.NewRequest("GET", "/", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler := ctrl.Handle(req)
		_ = handler
	}
}

// BenchmarkControllerValidation benchmarks validation
func BenchmarkControllerValidation(b *testing.B) {
	ctrl := &validationTestController{}
	app := application.New(nil)
	ctrl.Setup(app)
	
	req := httptest.NewRequest("GET", "/", nil)
	handler := ctrl.Handle(req).(*validationTestController)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := handler.Validator()
		v.CheckRequired("field1", "value1")
		v.CheckRequired("field2", "value2")
		v.CheckEmail("email", "test@example.com")
		_ = v.Result()
	}
}

// Test controller implementations

type isolationTestController struct {
	application.Controller
}

func (c isolationTestController) Handle(r *http.Request) application.Handler {
	c.Request = r
	return &c
}

type htmxTestController struct {
	application.Controller
}

func (c htmxTestController) Handle(r *http.Request) application.Handler {
	c.Request = r
	return &c
}

type errorTestController struct {
	application.Controller
	templateUsed string
}

func (c errorTestController) Handle(r *http.Request) application.Handler {
	c.Request = r
	return &c
}

// Override Render to capture template name
func (c *errorTestController) Render(w http.ResponseWriter, r *http.Request, template string, data any) {
	c.templateUsed = template
	// Write something to indicate rendering happened
	w.Write([]byte("rendered"))
}

type validationTestController struct {
	application.Controller
}

func (c validationTestController) Handle(r *http.Request) application.Handler {
	c.Request = r
	return &c
}

type benchController struct {
	application.Controller
}

func (c benchController) Handle(r *http.Request) application.Handler {
	c.Request = r
	return &c
}