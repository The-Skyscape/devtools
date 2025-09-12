package application

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Common sentinel errors for consistent error handling
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrValidation   = errors.New("validation failed")
	ErrInternal     = errors.New("internal server error")
)

type IController interface {
	Setup(*App)
	Handle(*http.Request) IController
}

type Controller struct {
	*App
	*http.Request
}

func (base *Controller) Setup(app *App) {
	base.App = app
}

// SetRequest sets the current request on the controller
func (base *Controller) SetRequest(r *http.Request) {
	base.Request = r
}

func (base *Controller) Use(name string) IController {
	ctrl := base.App.Use(name)
	if ctrl == nil {
		return nil
	}
	return ctrl.Handle(base.Request)
}

// UseWithRequest returns a controller with a specific request set
// This is useful when you need to use another controller but with a different request
func (base *Controller) UseWithRequest(name string, r *http.Request) IController {
	ctrl := base.App.Use(name)
	if ctrl == nil {
		return nil
	}
	return ctrl.Handle(r)
}

// Params returns a parameter helper for the current request
func (c *Controller) Params() *Params {
	return NewParams(c.Request)
}

// MultipartParams returns a multipart parameter helper for file uploads
func (c *Controller) MultipartParams() *Params {
	return NewMultipartParams(c.Request)
}

// Pagination returns pagination parameters from the request
func (c *Controller) Pagination(defaultPageSize int) *Pagination {
	return GetPagination(c.Request, defaultPageSize)
}

// Sort returns sort parameters from the request
func (c *Controller) Sort(defaultField string) *Sort {
	return GetSort(c.Request, defaultField)
}

// Validator returns a new validator for building validation errors
func (c *Controller) Validator() *Validator {
	return NewValidator()
}

// IsHTMX checks if the request is from HTMX
func (c *Controller) IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// Refresh triggers a page refresh (HTMX-aware)
func (c *Controller) Refresh(w http.ResponseWriter, r *http.Request) {
	if c.IsHTMX(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Non-HTMX fallback - redirect to current URL
	http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
}

// Redirect sends a redirect response (HTMX-aware)
func (c *Controller) Redirect(w http.ResponseWriter, r *http.Request, path string) {
	if c.IsHTMX(r) {
		// Use HX-Location for client-side redirect
		w.Header().Set("HX-Location", c.hostPrefix+path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// RenderError renders an error using the appropriate template.
// It automatically selects the right template based on the error type:
//   - ValidationError -> "validation-errors.html"
//   - ErrNotFound -> "error-404.html"
//   - ErrForbidden -> "error-403.html"
//   - ErrUnauthorized -> "error-401.html"
//   - Others -> "error-message.html"
//
// IMPORTANT: Returns 200 OK for HTMX compatibility (needs to swap content)
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

func (c *Controller) RenderString(templateName string, data any) (string, error) {
	// Render a template to a string instead of writing to ResponseWriter
	var buf bytes.Buffer
	c.App.Render(&buf, c.Request, templateName, data)
	return buf.String(), nil
}

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
