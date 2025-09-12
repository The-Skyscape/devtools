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

type Controller interface {
	Setup(*App)
	Handle(*http.Request) Controller
}

type BaseController struct {
	*App
	*http.Request
}

func (base *BaseController) Setup(app *App) {
	base.App = app
}

// SetRequest sets the current request on the controller
func (base *BaseController) SetRequest(r *http.Request) {
	base.Request = r
}

func (base *BaseController) Use(name string) Controller {
	ctrl := base.App.Use(name)
	if ctrl == nil {
		return nil
	}
	return ctrl.Handle(base.Request)
}

// UseWithRequest returns a controller with a specific request set
// This is useful when you need to use another controller but with a different request
func (base *BaseController) UseWithRequest(name string, r *http.Request) Controller {
	ctrl := base.App.Use(name)
	if ctrl == nil {
		return nil
	}
	return ctrl.Handle(r)
}

// Params returns a parameter helper for the current request
func (c *BaseController) Params() *Params {
	return NewParams(c.Request)
}

// MultipartParams returns a multipart parameter helper for file uploads
func (c *BaseController) MultipartParams() *Params {
	return NewMultipartParams(c.Request)
}

// Pagination returns pagination parameters from the request
func (c *BaseController) Pagination(defaultPageSize int) *Pagination {
	return GetPagination(c.Request, defaultPageSize)
}

// Sort returns sort parameters from the request
func (c *BaseController) Sort(defaultField string) *Sort {
	return GetSort(c.Request, defaultField)
}

// Validator returns a new validator for building validation errors
func (c *BaseController) Validator() *Validator {
	return NewValidator()
}

// IsHTMX checks if the request is from HTMX
func (c *BaseController) IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// Refresh triggers a page refresh (HTMX-aware)
func (c *BaseController) Refresh(w http.ResponseWriter, r *http.Request) {
	if c.IsHTMX(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Non-HTMX fallback - redirect to current URL
	http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
}

// Redirect sends a redirect response (HTMX-aware)
func (c *BaseController) Redirect(w http.ResponseWriter, r *http.Request, path string) {
	if c.IsHTMX(r) {
		// Use HX-Location for client-side redirect
		w.Header().Set("HX-Location", c.hostPrefix+path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// RenderError renders an error message (HTMX-aware)
func (c *BaseController) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	// IMPORTANT: We intentionally don't set HTTP status codes here
	// HTMX needs 200 OK to swap the error content into the DOM
	c.Render(w, r, "error-message.html", err)
}

// RenderErrorMsg renders an error message from string
func (c *BaseController) RenderErrorMsg(w http.ResponseWriter, r *http.Request, msg string) {
	c.RenderError(w, r, errors.New(msg))
}

// RenderValidationError renders validation errors
func (c *BaseController) RenderValidationError(w http.ResponseWriter, r *http.Request, err error) {
	if ve, ok := err.(ValidationError); ok {
		c.Render(w, r, "validation-errors.html", ve)
	} else {
		c.RenderError(w, r, err)
	}
}

// RenderNotFound renders a 404 error
func (c *BaseController) RenderNotFound(w http.ResponseWriter, r *http.Request) {
	c.Render(w, r, "error-404.html", ErrNotFound)
}

// RenderForbidden renders a 403 error
func (c *BaseController) RenderForbidden(w http.ResponseWriter, r *http.Request) {
	c.Render(w, r, "error-403.html", ErrForbidden)
}

func (c *BaseController) RenderString(templateName string, data any) (string, error) {
	// Render a template to a string instead of writing to ResponseWriter
	var buf bytes.Buffer
	c.App.Render(&buf, c.Request, templateName, data)
	return buf.String(), nil
}

func (c *BaseController) EventStream(w http.ResponseWriter, r *http.Request) (func(string, any), error) {
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
