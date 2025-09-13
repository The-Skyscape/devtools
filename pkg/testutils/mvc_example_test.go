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

// Use value receiver for proper request isolation
func (c ExampleController) Handle(req *http.Request) application.Handler {
	c.Request = req
	c.CallCount++
	return &c
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
	prefix, controller := NewExampleController()
	testutils.TestController(mvc, prefix, controller, func(ct *testutils.ControllerTest[*ExampleController]) {
		// Test HTTP handlers
		ct.Run("HTTPHandlers", func(t *testing.T, c *ExampleController) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			})

			req := ct.Request("GET", "/api/status")
			rec := ct.TestHandler(handler, req)

			ct.AssertStatus(rec, http.StatusOK)
			ct.AssertJSON(rec)
			ct.AssertContains(rec, "ok")
		})

		// Test form submissions
		ct.Run("FormSubmission", func(t *testing.T, c *ExampleController) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				name := r.FormValue("name")
				if name == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello " + name))
			})

			req := ct.RequestWithForm("/submit", map[string]string{
				"name":  "Alice",
				"email": "alice@example.com",
			})
			rec := ct.TestHandler(handler, req)

			ct.AssertStatus(rec, http.StatusOK)
			ct.AssertContains(rec, "Hello Alice")
		})

		// Test redirects
		ct.Run("Redirects", func(t *testing.T, c *ExampleController) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/dashboard", http.StatusFound)
			})

			req := ct.Request("POST", "/login")
			rec := ct.TestHandler(handler, req)

			ct.AssertRedirect(rec, "/dashboard")
		})

		// Test cookies
		ct.Run("Cookies", func(t *testing.T, c *ExampleController) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.SetCookie(w, &http.Cookie{
					Name:  "session",
					Value: "abc123",
					Path:  "/",
				})
				w.WriteHeader(http.StatusOK)
			})

			req := ct.Request("GET", "/")
			rec := ct.TestHandler(handler, req)

			cookie := ct.AssertCookie(rec, "session")
			if cookie.Value != "abc123" {
				t.Errorf("Expected cookie value 'abc123', got %s", cookie.Value)
			}
		})

		// Test authenticated requests
		ct.Run("AuthenticatedRequest", func(t *testing.T, c *ExampleController) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				cookie, err := r.Cookie("auth")
				if err != nil || cookie.Value != "valid-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			// Test without auth
			req := ct.Request("GET", "/protected")
			rec := ct.TestHandler(handler, req)
			ct.AssertStatus(rec, http.StatusUnauthorized)

			// Test with auth
			req = ct.RequestWithCookie("GET", "/protected", "auth", "valid-token")
			rec = ct.TestHandler(handler, req)
			ct.AssertStatus(rec, http.StatusOK)
		})
	})
}
