package application

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

func (base *BaseController) Use(name string) Controller {
	ctrl := base.App.Use(name)
	if ctrl == nil {
		return nil
	}
	return ctrl.Handle(base.Request)
}

func (c *BaseController) Atoi(name string, defaultValue int) int {
	value := c.URL.Query().Get(name)
	value = cmp.Or(value, c.FormValue(name))
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func (c *BaseController) Refresh(w http.ResponseWriter, r *http.Request) {
	// HTMX requests should trigger a refresh
	if htmx := r.Header.Get("HX-Request"); htmx != "" {
		w.Header().Add("HX-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Non-HTMX fallback - redirect to current URL
	http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
}

func (c *BaseController) Redirect(w http.ResponseWriter, r *http.Request, path string) {
	if htmx := r.Header.Get("HX-Request"); htmx != "" {
		w.Header().Add("HX-Location", c.hostPrefix+path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

func (c *BaseController) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	// Standardized error rendering for consistency across controllers
	c.Render(w, r, "error-message.html", err)
}

func (c *BaseController) RenderErrorMsg(w http.ResponseWriter, r *http.Request, msg string) {
	// Convenience method for rendering error messages
	c.RenderError(w, r, errors.New(msg))
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
		data = strings.ReplaceAll(buf.String(), "\n", "")
		if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", data); err != nil {
			log.Println("Failed to flush: ", template, data)
		}
		flusher.Flush()
	}, nil
}
