// Package application provides a web application framework with MVC pattern support,
// template rendering, middleware chains, and HTMX integration.
package application

import (
	"cmp"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/The-Skyscape/devtools/pkg/application/builtins"
)

// App represents a web application with controllers, views, and middleware.
// It provides a complete MVC framework for building web applications with
// server-side rendering and HTMX support.
type App struct {
	controllers map[string]IController
	viewEngine  *template.Template
	hostPrefix  string
	views       []fs.FS
	theme       string
	middlewares []Middleware
}

// Middleware defines the interface for HTTP middleware.
// Middleware can intercept and modify HTTP requests and responses.
type Middleware interface {
	Handle(http.Handler) http.Handler
}

// AccessCheck is a function that determines if a request should be allowed.
// It returns true if access is granted, false otherwise.
// When returning false, the function should handle sending the appropriate
// response (e.g., redirect to login, show error).
type AccessCheck func(*App, http.ResponseWriter, *http.Request) bool

// Serve is a convenience function that creates and starts a new application
// with the provided views and options. It logs startup information and
// terminates the program if the server fails to start.
//
// This is the simplest way to start an application:
//
//	//go:embed all:views
//	var views embed.FS
//
//	func main() {
//		application.Serve(views,
//			application.WithController(controllers.Home()),
//			application.WithDaisyTheme("corporate"),
//		)
//	}
func Serve(views fs.FS, opts ...Option) {
	log.Printf("🚀 Starting Skyscape Application...")
	log.Printf("📱 Visit: http://localhost:%s", cmp.Or(os.Getenv("PORT"), "8080"))

	app := New(views, opts...)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// New creates a new App with the given views and options.
// The views parameter should be an embedded filesystem containing templates.
// Options can be used to configure controllers, middleware, themes, etc.
//
// If views contains a "views/public" directory, it will be automatically
// served as static files at the "/public/" URL path.
func New(views fs.FS, opts ...Option) *App {
	app := App{
		controllers: map[string]IController{},
		views:       []fs.FS{appViews},
		theme:       "retro",
		middlewares: []Middleware{},
	}

	if views != nil {
		app.views = append(app.views, views)

		// Auto-serve public directory if it exists
		if _, err := fs.Sub(views, "views/public"); err == nil {
			public, _ := fs.Sub(views, "views")
			http.Handle("GET /public/", http.FileServerFS(public))
		}
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(&app); err != nil {
			log.Fatal("Failed to setup Application:", err)
		}
	}

	return &app
}

// Server prepares the application for serving and returns the address and handler.
// This method:
//   - Prepares all views and templates
//   - Builds the middleware chain
//   - Returns the configured address and HTTP handler
//
// The returned handler can be used with any HTTP server implementation.
// This is useful for testing or when you need custom server configuration.
func (app *App) Server() (string, http.Handler) {
	log.Println("Preparing Application...")

	app.prepareViews()

	// Build middleware chain in reverse order
	var handler http.Handler = http.DefaultServeMux
	for i := len(app.middlewares) - 1; i >= 0; i-- {
		handler = app.middlewares[i].Handle(handler)
	}

	addr := "0.0.0.0:" + cmp.Or(os.Getenv("PORT"), "5000")
	return addr, handler
}

// Start runs the application HTTP server on the configured port.
// It also starts an HTTPS server if SSL certificates are available.
//
// SSL certificates are configured via environment variables:
//   - SKYSCAPE_SSL_FULLCHAIN: Path to the full certificate chain (default: /root/fullchain.pem)
//   - SKYSCAPE_SSL_PRIVKEY: Path to the private key (default: /root/privkey.pem)
//
// The HTTP server runs on the PORT environment variable (default: 5000).
// The HTTPS server always runs on port 443 if certificates are found.
func (app *App) Start() error {
	addr, handler := app.Server()

	// Start HTTPS server in background if certificates exist
	go func() {
		cert := cmp.Or(os.Getenv("SKYSCAPE_SSL_FULLCHAIN"), "/root/fullchain.pem")
		if _, err := os.Stat(cert); err != nil {
			log.Println("No SSL Certificate found at:", cert)
			return
		}

		key := cmp.Or(os.Getenv("SKYSCAPE_SSL_PRIVKEY"), "/root/privkey.pem")
		if _, err := os.Stat(key); err != nil {
			log.Println("No SSL Key found at:", key)
			return
		}

		if cert != "" && key != "" {
			log.Print("Serving Secure Application @ https://localhost:443")
			log.Fatal(http.ListenAndServeTLS("0.0.0.0:443", cert, key, handler))
		}
	}()

	log.Print("Serving Application @ http://" + addr)
	return http.ListenAndServe(addr, handler)
}

// Use returns the controller registered with the given name.
// Returns nil if no controller is found with that name.
//
// This is primarily used by controllers to access other controllers:
//
//	func (c *MyController) Handle(r *http.Request) Controller {
//		auth := c.App.Use("auth").(*AuthController)
//		// ...
//	}
func (app App) Use(name string) IController {
	return app.controllers[name]
}

// SetTheme updates the application's UI theme.
// The theme is accessible in templates via the {{theme}} function.
func (app *App) SetTheme(theme string) {
	app.theme = theme
}

// Render executes a template with the given data and writes it to the writer.
// It automatically injects runtime functions like:
//   - req: Returns the current *http.Request
//   - host: Returns the application's host prefix
//   - path_eq: Checks if the current path matches the given segments
//   - All registered controllers by name
//
// If the template is not found, it returns a 404 error for HTTP responses
// or exits the program for non-HTTP contexts.
func (app *App) Render(w io.Writer, r *http.Request, page string, data any) {
	// Create a copy of built-in functions to avoid race conditions
	funcs := make(template.FuncMap)
	for k, v := range builtins.FuncMap {
		funcs[k] = v
	}

	// Add runtime-specific functions
	funcs["req"] = func() *http.Request { return r }
	funcs["host"] = func() string { return app.hostPrefix }
	funcs["path_eq"] = func(parts ...string) bool {
		path := fmt.Sprintf("/%s", strings.Join(parts, "/"))
		return r.URL.Path == path
	}

	// Add controller functions
	for name, ctrl := range app.controllers {
		funcs[name] = func() IController { return ctrl.Handle(r) }
	}

	view := app.viewEngine.Lookup(page)
	if view == nil {
		log.Printf("Template not found: %s", page)
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, fmt.Sprintf("Template not found: %s", page), http.StatusNotFound)
			return
		} else {
			fmt.Fprintf(w, "Template not found in non-HTTP context: %s", page)
			os.Exit(1)
		}
	}

	if err := view.Funcs(funcs).Execute(w, data); err != nil {
		log.Print("Error rendering: ", err)
		app.viewEngine.ExecuteTemplate(w, "error-message", err)
	}
}

// Protect wraps an http.Handler with access control.
// If accessCheck is nil, the handler is called directly.
// If accessCheck returns false, it should handle the response itself
// (e.g., redirect to login) and the handler will not be called.
//
// Example:
//
//	http.Handle("/admin", app.Protect(adminHandler, auth.RequireAdmin))
func (app *App) Protect(h http.Handler, accessCheck AccessCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if accessCheck == nil {
			h.ServeHTTP(w, r)
			return
		}

		// If accessCheck returns false, it has already handled the response
		if !accessCheck(app, w, r) {
			return
		}

		h.ServeHTTP(w, r)
	}
}

// ProtectFunc is a convenience wrapper for Protect that accepts an http.HandlerFunc.
// It provides the same access control functionality as Protect.
//
// Example:
//
//	http.HandleFunc("/profile", app.ProtectFunc(profileHandler, auth.RequireLogin))
func (app *App) ProtectFunc(fn http.HandlerFunc, accessLevel AccessCheck) http.HandlerFunc {
	return app.Protect(fn, accessLevel)
}
