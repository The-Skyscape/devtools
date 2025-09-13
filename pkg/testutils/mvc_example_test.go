package testutils_test

import (
	"net/http"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/testutils"
)

// ExampleController shows how to test a controller with the MVC framework
type ExampleController struct {
	application.Controller
	CallCount int
}

func (c *ExampleController) Handle(req *http.Request) application.Handler {
	c.Request = req
	c.CallCount++
	return c
}

func (c *ExampleController) Setup(app *application.App) {
	c.Controller.Setup(app)
	// Register routes here
}

// NewExampleController is a factory function for the example controller
func NewExampleController() (string, *ExampleController) {
	return "example", &ExampleController{}
}

func TestMVCFramework(t *testing.T) {
	mvc := testutils.NewMVCTest(t)
	
	// Test controller factory and setup
	mvc.TestController("ExampleController", func() (string, application.Handler) {
		return NewExampleController()
	})
	
	// Test HTTP handlers
	mvc.Run("HTTPHandlers", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		
		req := mvc.Request("GET", "/api/status")
		rec := mvc.TestHandler(handler, req)
		
		mvc.AssertStatus(rec, http.StatusOK)
		mvc.AssertJSON(rec)
		mvc.AssertContains(rec, "ok")
	})
	
	// Test form submissions
	mvc.Run("FormSubmission", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			name := r.FormValue("name")
			if name == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello " + name))
		})
		
		req := mvc.RequestWithForm("/submit", map[string]string{
			"name": "Alice",
			"email": "alice@example.com",
		})
		rec := mvc.TestHandler(handler, req)
		
		mvc.AssertStatus(rec, http.StatusOK)
		mvc.AssertContains(rec, "Hello Alice")
	})
	
	// Test redirects
	mvc.Run("Redirects", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		})
		
		req := mvc.Request("POST", "/login")
		rec := mvc.TestHandler(handler, req)
		
		mvc.AssertRedirect(rec, "/dashboard")
	})
	
	// Test cookies
	mvc.Run("Cookies", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{
				Name:  "session",
				Value: "abc123",
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)
		})
		
		req := mvc.Request("GET", "/")
		rec := mvc.TestHandler(handler, req)
		
		cookie := mvc.AssertCookie(rec, "session")
		if cookie.Value != "abc123" {
			t.Errorf("Expected cookie value 'abc123', got %s", cookie.Value)
		}
	})
	
	// Test authenticated requests
	mvc.Run("AuthenticatedRequest", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth")
			if err != nil || cookie.Value != "valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		
		// Test without auth
		req := mvc.Request("GET", "/protected")
		rec := mvc.TestHandler(handler, req)
		mvc.AssertStatus(rec, http.StatusUnauthorized)
		
		// Test with auth
		req = mvc.RequestWithCookie("GET", "/protected", "auth", "valid-token")
		rec = mvc.TestHandler(handler, req)
		mvc.AssertStatus(rec, http.StatusOK)
	})
}