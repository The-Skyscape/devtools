package application

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Sentinel errors provide consistent error handling across the application.
// These errors can be checked with errors.Is() for type-safe comparisons.
var (
	// ErrNotFound indicates the requested resource doesn't exist.
	// Templates: error-404.html
	ErrNotFound = errors.New("not found")

	// ErrUnauthorized indicates missing or invalid authentication.
	// Templates: error-401.html
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the user lacks permission for the resource.
	// Templates: error-403.html
	ErrForbidden = errors.New("forbidden")

	// ErrValidation indicates input validation failed.
	// Templates: validation-errors.html
	ErrValidation = errors.New("validation failed")

	// ErrInternal indicates an unexpected server error.
	// Templates: error-message.html
	ErrInternal = errors.New("internal server error")
)

// Handler defines the interface for request handlers in the MVC framework.
// Controllers must implement this interface to be registered with the application.
type Handler interface {
	// Setup is called once during application initialization.
	// Controllers should register routes and initialize resources here.
	Setup(*App)

	// Handle creates a request-scoped instance of the controller.
	// CRITICAL: Must use VALUE receiver to ensure request isolation.
	// The returned Handler has the request set and can access it.
	Handle(*http.Request) Handler
}

// Controller provides base functionality for all application controllers.
//
// Request Isolation Pattern:
//
//	Each request gets its own controller copy through value receivers.
//	This eliminates shared state and prevents data races.
//
// Embedding Pattern:
//
//	Controllers should embed this type to inherit base functionality:
//
//	type MyController struct {
//	    application.Controller
//	}
//
// Template Access:
//
//	Public methods (capitalized) are accessible in templates.
//	Private methods (lowercase) serve as HTTP handlers.
type Controller struct {
	*App
	*http.Request
}

// Setup initializes the controller with the application instance.
// Controllers that override Setup must call this parent method:
//
//	func (c *MyController) Setup(app *application.App) {
//	    c.Controller.Setup(app)  // Required: call parent
//	    // Register routes here
//	}
func (base *Controller) Setup(app *App) {
	base.App = app
}

// Host returns the request Host and the app hostPrefix
func (base *Controller) Host() string {
	if base.Request == nil {
		return base.App.Host()
	}

	return fmt.Sprintf("%s%s", base.Request.Host, base.App.Host())
}

// Use returns another controller configured for the current request.
// This enables controller composition and delegation.
//
// Example:
//
//	auth := c.Use("auth").(*AuthController)
//	if !auth.IsLoggedIn() {
//	    c.Redirect(w, r, "/login")
//	    return
//	}
//
// Returns nil if the controller is not registered.
func (base *Controller) Use(name string) Handler {
	ctrl := base.App.Use(name)
	if ctrl == nil {
		return nil
	}
	return ctrl.Handle(base.Request)
}

// UseWithRequest returns a controller configured with a specific request.
// This is useful for delegation with modified request contexts.
//
// Example:
//
//	// Create a modified request with additional context
//	modifiedReq := r.WithContext(newContext)
//	other := c.UseWithRequest("other", modifiedReq)
func (base *Controller) UseWithRequest(name string, r *http.Request) Handler {
	ctrl := base.App.Use(name)
	if ctrl == nil {
		return nil
	}
	return ctrl.Handle(r)
}

// Params returns a parameter helper for extracting request values.
// It provides type-safe access to form values, query parameters, and path values.
func (c *Controller) Params() *Params {
	return NewParams(c.Request)
}

// MultipartParams returns a parameter helper for multipart forms.
// Use this for handling file uploads and mixed form data.
func (c *Controller) MultipartParams() *Params {
	return NewMultipartParams(c.Request)
}

// Pagination extracts pagination parameters from the request.
// Common query parameters: ?page=2&size=20
func (c *Controller) Pagination(defaultPageSize int) *Pagination {
	return GetPagination(c.Request, defaultPageSize)
}

// Sort extracts sorting parameters from the request.
// Common query parameters: ?sort=name&order=asc
func (c *Controller) Sort(defaultField string) *Sort {
	return GetSort(c.Request, defaultField)
}

// Validator creates a new validator for input validation.
// Build validation errors incrementally and check with Result().
//
// Example:
//
//	v := c.Validator()
//	v.CheckRequired("name", r.FormValue("name"))
//	v.CheckEmail("email", r.FormValue("email"))
//	if err := v.Result(); err != nil {
//	    c.RenderError(w, r, err)
//	    return
//	}
func (c *Controller) Validator() *Validator {
	return NewValidator()
}

// IsHTMX checks if the request originates from HTMX.
// HTMX sets the "HX-Request: true" header on all requests.
// Use this to return partial HTML for HTMX vs full pages for direct navigation.
func (c *Controller) IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// Refresh triggers a full page refresh.
//
// HTMX Behavior:
//   - Sets "HX-Refresh: true" header
//   - Returns 204 No Content
//   - HTMX reloads the entire page
//
// Standard Behavior:
//   - Redirects to current URL (POST-Redirect-GET pattern)
//
// Use this after successful form submissions to refresh the page state.
func (c *Controller) Refresh(w http.ResponseWriter, r *http.Request) {
	if c.IsHTMX(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Non-HTMX fallback - redirect to current URL
	http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
}

// Redirect navigates to a different path.
//
// HTMX Behavior:
//   - Sets "HX-Location" header with target path
//   - Returns 204 No Content
//   - HTMX performs client-side navigation
//
// Standard Behavior:
//   - Returns 303 See Other redirect
//
// IMPORTANT: Always use this instead of http.Redirect for HTMX compatibility.
func (c *Controller) Redirect(w http.ResponseWriter, r *http.Request, path string) {
	if c.IsHTMX(r) {
		// Use HX-Location for client-side redirect
		w.Header().Set("HX-Location", c.hostPrefix+path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// RenderError renders an error using the appropriate error template.
//
// Template Selection:
//   - ValidationError → "validation-errors.html"
//   - ErrNotFound → "error-404.html"
//   - ErrForbidden → "error-403.html"
//   - ErrUnauthorized → "error-401.html"
//   - Others → "error-message.html"
//
// CRITICAL: Returns 200 OK status, not error codes.
// Why: HTMX requires 200 OK to swap error content into the DOM.
// This is intentional design, not a bug.
//
// For HTMX forms, target error containers:
//
//	<form hx-post="/submit" hx-target=".error-message">
func (c *Controller) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	// Select template based on error type
	template := "error-message.html"

	switch {
	case errors.Is(err, ErrNotFound):
		template = "error-404.html"
	case errors.Is(err, ErrForbidden):
		template = "error-403.html"
	case errors.Is(err, ErrUnauthorized):
		template = "error-401.html"
	case err != nil:
		if _, ok := err.(ValidationError); ok {
			template = "validation-errors.html"
		}
	}

	c.Render(w, r, template, err)
}

// RenderString renders a template to a string instead of HTTP response.
// Useful for generating HTML for emails, reports, or embedded content.
//
// Example:
//
//	html, err := c.RenderString("email-template.html", data)
//	// Send html via email service
func (c *Controller) RenderString(templateName string, data any) (string, error) {
	var buf bytes.Buffer
	c.App.Render(&buf, c.Request, templateName, data)
	return buf.String(), nil
}

// EventStream establishes a Server-Sent Events connection.
// Returns a function to send events to the client.
//
// Example:
//
//	send, err := c.EventStream(w, r)
//	if err != nil {
//	    return
//	}
//	for i := 0; i < 10; i++ {
//	    send("update", map[string]int{"progress": i * 10})
//	    time.Sleep(time.Second)
//	}
func (c *Controller) EventStream(w http.ResponseWriter, r *http.Request) (func(string, any), error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("event streaming not supported")
	}

	fmt.Fprintf(w, "event: ping\ndata: pong\n\n")
	flusher.Flush()

	return func(template string, data any) {
		var buf bytes.Buffer
		c.Render(&buf, r, template, data)
		// For SSE, we need to escape newlines in the data field
		// Each line of data must be prefixed with "data: "
		lines := strings.Split(buf.String(), "\n")
		fmt.Fprintf(w, "event: message\n")
		for _, line := range lines {
			if line != "" {
				fmt.Fprintf(w, "data: %s\n", line)
			}
		}
		if _, err := fmt.Fprintf(w, "\n"); err != nil {
			log.Println("Failed to flush: ", template, data)
		}
		flusher.Flush()
	}, nil
}
