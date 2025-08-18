package application

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

//go:embed all:views
var appViews embed.FS

type View struct {
	app         *App
	name        string
	accessCheck AccessCheck
}

func (app *App) Serve(name string, accessCheck AccessCheck) *View {
	return &View{app: app, name: name, accessCheck: accessCheck}
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data any) {
	v.app.Render(w, r, v.name, data)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if v.accessCheck == nil {
		v.app.Render(w, r, v.name, nil)
		return
	}

	// If accessCheck returns false, it has already handled the response
	if !v.accessCheck(v.app, w, r) {
		return
	}

	v.app.Render(w, r, v.name, nil)
}

func (app *App) prepareViews() {
	funcs := template.FuncMap{
		"req":     func() *http.Request { return nil },
		"host":    func() string { return app.hostPrefix },
		"path":    func(parts ...string) string { return fmt.Sprintf("/%s", strings.Join(parts, "/")) },
		"theme":   func() string { return app.theme },
		"title":   func(title string) string { return strings.ReplaceAll(title, "_", " ") },
		"prefix":  func(s, prefix string) bool { return strings.HasPrefix(s, prefix) },
		"path_eq": func(parts ...string) bool { return false },
		// Math functions
		"add": func(a, b interface{}) interface{} {
			switch va := a.(type) {
			case int:
				if vb, ok := b.(int); ok {
					return va + vb
				}
			case float64:
				if vb, ok := b.(float64); ok {
					return va + vb
				}
			}
			return 0
		},
		"sub": func(a, b interface{}) interface{} {
			switch va := a.(type) {
			case int:
				if vb, ok := b.(int); ok {
					return va - vb
				}
			case float64:
				if vb, ok := b.(float64); ok {
					return va - vb
				}
			}
			return 0
		},
		"mul": func(a, b interface{}) interface{} {
			switch va := a.(type) {
			case int:
				if vb, ok := b.(int); ok {
					return va * vb
				}
				if vb, ok := b.(float64); ok {
					return float64(va) * vb
				}
			case float64:
				if vb, ok := b.(float64); ok {
					return va * vb
				}
				if vb, ok := b.(int); ok {
					return va * float64(vb)
				}
			}
			return 0
		},
		"div": func(a, b interface{}) interface{} {
			switch va := a.(type) {
			case int:
				if vb, ok := b.(int); ok && vb != 0 {
					return va / vb
				}
				if vb, ok := b.(float64); ok && vb != 0 {
					return float64(va) / vb
				}
			case float64:
				if vb, ok := b.(float64); ok && vb != 0 {
					return va / vb
				}
				if vb, ok := b.(int); ok && vb != 0 {
					return va / float64(vb)
				}
			}
			return 0
		},
		// Type conversion functions
		"float": func(v interface{}) float64 {
			switch val := v.(type) {
			case int:
				return float64(val)
			case float64:
				return val
			case string:
				// Try to parse string to float
				var f float64
				fmt.Sscanf(val, "%f", &f)
				return f
			}
			return 0
		},
		"toString": func(v interface{}) string {
			return fmt.Sprintf("%v", v)
		},
		// Utility functions
		"slice": func(s string, start, end int) string {
			if start < 0 {
				start = 0
			}
			if end > len(s) {
				end = len(s)
			}
			if start > end {
				return ""
			}
			return s[start:end]
		},
		"head": func(n int, arr interface{}) interface{} {
			// Simple head implementation for slices
			switch v := arr.(type) {
			case []interface{}:
				if n > len(v) {
					return v
				}
				return v[:n]
			}
			return arr
		},
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" || val == 0 {
				return def
			}
			return val
		},
		"set": func(m map[string]interface{}, key string, val interface{}) map[string]interface{} {
			if m == nil {
				m = make(map[string]interface{})
			}
			m[key] = val
			return m
		},
		"dict": func() map[string]interface{} {
			return make(map[string]interface{})
		},
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl }
	}

	if app.viewEngine == nil {
		app.viewEngine = template.New("")
	}

	app.viewEngine = app.viewEngine.Funcs(funcs)
	for _, source := range app.views {
		if tmpl, err := app.viewEngine.ParseFS(source, "views/*.html"); err == nil {
			app.viewEngine = tmpl
		} else {
			log.Fatal("Failed to parse root views", err)
		}

		if tmpl, err := app.viewEngine.ParseFS(source, "views/**/*.html"); err == nil {
			app.viewEngine = tmpl
		} else {
			log.Print("Failed to parse views", err)
		}

		if tmpl, err := app.viewEngine.ParseFS(source, "views/**/**/*.html"); err == nil {
			app.viewEngine = tmpl
		} else {
			log.Print("Failed to parse views", err)
		}
	}
}
