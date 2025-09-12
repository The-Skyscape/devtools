package application

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/The-Skyscape/devtools/pkg/application/builtins"
	"github.com/The-Skyscape/devtools/pkg/charting"
)

// appViews contains embedded framework views (error pages, defaults).
// These are always available as fallbacks if user templates are missing.
//go:embed all:views
var appViews embed.FS

// View represents a renderable template with optional access control.
// Views are created via app.Serve() and handle both rendering and authorization.
type View struct {
	app         *App
	name        string
	accessCheck AccessCheck
}

// Serve creates a View that renders the named template with access control.
//
// Example:
//
//	http.Handle("GET /admin", app.Serve("admin.html", auth.RequireAdmin))
//	http.Handle("GET /public", app.Serve("public.html", nil))  // No access check
func (app *App) Serve(name string, accessCheck AccessCheck) *View {
	return &View{app: app, name: name, accessCheck: accessCheck}
}

// Render executes the view's template with the provided data.
// This bypasses access control - use ServeHTTP for protected views.
func (v *View) Render(w http.ResponseWriter, r *http.Request, data any) {
	v.app.Render(w, r, v.name, data)
}

// ServeHTTP implements http.Handler, enforcing access control before rendering.
//
// Access Control Flow:
//  1. If no accessCheck, renders immediately
//  2. Calls accessCheck(app, w, r)
//  3. If false, accessCheck has sent response (redirect, error)
//  4. If true, renders the template
//
// This allows Views to be used directly as HTTP handlers.
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

// prepareViews compiles templates with functions and controllers.
// Called once during app.Start() to prepare all templates.
//
// Template Compilation:
//  1. Loads built-in template functions
//  2. Adds runtime functions (req, host, path, theme)
//  3. Registers controllers as template functions
//  4. Parses all template files from embedded filesystems
//
// Performance Note: This happens once at startup, not per request.
// Templates are compiled and cached for the application lifetime.
func (app *App) prepareViews() {
	// Get all built-in functions
	builtinFuncs := builtins.FuncMap

	// Start with built-in functions as base
	funcs := builtinFuncs

	// Add app-specific functions
	funcs["req"] = func() *http.Request { return nil }
	funcs["host"] = func() string { return app.hostPrefix }
	funcs["path"] = func(parts ...string) string { return fmt.Sprintf("/%s", strings.Join(parts, "/")) }
	funcs["theme"] = func() string { return app.theme }
	funcs["path_eq"] = func(parts ...string) bool { return false }

	// Override the title function to use the specific behavior
	funcs["title"] = func(title string) string { return strings.ReplaceAll(title, "_", " ") }
	funcs["prefix"] = func(s, prefix string) bool { return strings.HasPrefix(s, prefix) }

	// JSON functionsany
	funcs["jsonify"] = func(v any) template.JS {
		data, err := json.Marshal(v)
		if err != nil {
			log.Printf("jsonify error: %v", err)
			return template.JS("{}")
		}
		return template.JS(data)
	}

	// Charting functions
	funcs["renderChart"] = func(dataOrFunc any, placeholder ...string) template.HTML {
		// Handle function call if passed
		var data *charting.ChartData

		// Check if it's a function that returns ChartData
		switch v := dataOrFunc.(type) {
		case func() any:
			if result := v(); result != nil {
				data, _ = result.(*charting.ChartData)
			}
		case func() *charting.ChartData:
			data = v()
		case *charting.ChartData:
			data = v
		}

		// Return placeholder if no data
		if data == nil || len(data.Data) == 0 {
			title := "No data"
			message := "No data available"
			if len(placeholder) > 0 {
				title = placeholder[0]
			}
			if len(placeholder) > 1 {
				message = placeholder[1]
			}
			return charting.PlaceholderChart(title, message)
		}

		return charting.RenderLineChart(data, 600, 300)
	}

	funcs["renderSparkline"] = func(data []float64) template.HTML {
		return charting.RenderSparkline(data, 100, 30)
	}

	funcs["placeholderChart"] = func(title, message string) template.HTML {
		return charting.PlaceholderChart(title, message)
	}

	funcs["chartLoader"] = func(endpoint, title string) template.HTML {
		return template.HTML(fmt.Sprintf(`
			<div hx-get="%s" 
			     hx-trigger="load" 
			     hx-swap="innerHTML"
			     class="chart-container">
				<div class="flex flex-col items-center justify-center h-48 text-base-content/60">
					<span class="loading loading-spinner loading-md"></span>
					<p class="text-sm mt-2">Loading %s...</p>
				</div>
			</div>
		`, endpoint, title))
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Handler { return ctrl }
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
