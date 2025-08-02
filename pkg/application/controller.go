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
	if htmx := r.Header.Get("HX-Request"); htmx != "" {
		w.Header().Add("Hx-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, c.URL.String(), http.StatusSeeOther)
}

func (c *BaseController) Redirect(w http.ResponseWriter, r *http.Request, path string) {
	if htmx := r.Header.Get("HX-Request"); htmx != "" {
		w.Header().Add("Hx-Location", c.hostPrefix+path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
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
