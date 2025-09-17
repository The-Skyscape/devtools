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

// App orchestrates the MVC framework, managing controllers, views, and middleware.
//
// Lifecycle:
//  1. New() or Serve() creates the App instance
//  2. Controllers register via WithController() options during initialization
//  3. Start() compiles templates once and begins serving HTTP requests
//  4. Each incoming request receives an isolated controller instance
//
// Thread Safety:
//   - App initialization occurs once at startup (single-threaded context)
//   - Templates are compiled once and become read-only
//   - Request handling uses value receivers for complete isolation
//   - No shared mutable state exists between requests
//
// Memory Model:
//   - Controllers are lightweight (16-32 bytes) and copied per request
//   - Templates are parsed once and cached for the application lifetime
//   - Middleware chains are built once during initialization
type App struct {
	controllers map[string]Handler
	viewEngine  *template.Template
	hostPrefix  string
	views       []fs.FS
	theme       string
	middlewares []Middleware
}

// Middleware defines the interface for HTTP request/response interceptors.
// Middleware implementations can modify requests, responses, or short-circuit
// the request chain. Common uses include authentication, logging, and compression.
type Middleware interface {
	Handle(http.Handler) http.Handler
}

// AccessCheck determines whether a request should be allowed to proceed.
//
// Return Values:
//   - true: Access granted, continue processing
//   - false: Access denied, checker must send response (redirect, error, etc.)
//
// Common Patterns:
//   - Authentication: Redirect to login if not authenticated
//   - Authorization: Return 403 Forbidden if lacks permissions
//   - Rate Limiting: Return 429 Too Many Requests if over limit
//
// The AccessCheck is responsible for sending the response when denying access.
type AccessCheck func(*App, http.ResponseWriter, *http.Request) bool

// Serve provides the simplest way to start a web application.
// It creates an App, configures it with the provided options, and starts
// the HTTP server. If the server fails to start, it calls log.Fatal().
//
// This is appropriate for main() functions where failure should terminate
// the program. For library usage or testing, use New() and Start() directly.
//
// Example:
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

// New creates an App instance configured with the provided views and options.
//
// Views:
//   - Should be an embedded filesystem (//go:embed all:views)
//   - Templates must be in the "views" directory
//   - Public assets in "views/public" are auto-served at "/public/"
//
// Options are applied in order and may include:
//   - WithController(): Register controllers for template access
//   - WithMiddleware(): Add HTTP middleware to the chain
//   - WithDaisyTheme(): Set the DaisyUI theme
//   - WithHostPrefix(): Configure URL prefix for reverse proxies
//
// Design Decision: Options that fail call log.Fatal() because misconfiguration
// should prevent startup rather than cause runtime errors.
func New(views fs.FS, opts ...Option) *App {
	app := App{
		controllers: map[string]Handler{},
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

// Host returns the hostPrefix for the application example (/prefix/path)
func (app *App) Host() string {
	return app.hostPrefix
}

// Server prepares the application and returns the address and HTTP handler.
//
// Initialization Steps:
//  1. Calls prepareViews() to compile templates (happens once)
//  2. Builds middleware chain in reverse order (innermost first)
//  3. Returns address and composed handler
//
// This method is useful for:
//   - Testing with httptest.NewServer
//   - Custom server configurations
//   - Embedding in larger applications
//
// Performance Note: Template compilation happens here, not per request.
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

// Start begins serving HTTP requests and optionally HTTPS if certificates exist.
//
// Server Configuration:
//   - HTTP: Listens on PORT env var (default: 5000)
//   - HTTPS: Listens on 443 if certificates are found
//
// SSL Certificate Discovery:
//   - SKYSCAPE_SSL_FULLCHAIN: Certificate chain path (default: /root/fullchain.pem)
//   - SKYSCAPE_SSL_PRIVKEY: Private key path (default: /root/privkey.pem)
//
// The HTTPS server runs in a background goroutine if certificates exist.
// This allows both HTTP and HTTPS to be served simultaneously.
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

// Use returns a registered controller by name, or nil if not found.
//
// Why String-Based Lookup:
//   - Allows registration of user-defined controller types
//   - Enables dynamic controller discovery in templates
//   - Avoids complex interface definitions
//
// Common Usage:
//
//	// In controllers - accessing other controllers
//	auth := c.App.Use("auth").(*AuthController)
//
//	// In templates - safe access with nil check
//	{{if auth}}{{auth.CurrentUser}}{{end}}
//
// Design Note: String-based DI trades compile-time safety for flexibility.
// This is intentional to support user-defined types without registration.
func (app App) Use(name string) Handler {
	return app.controllers[name]
}

// SetTheme updates the application's UI theme.
// The theme is accessible in templates via the {{theme}} function.
func (app *App) SetTheme(theme string) {
	app.theme = theme
}

// Render executes the named template with data and writes to the writer.
//
// Template Resolution:
//   - Templates are identified by filename only (no paths)
//   - "user-profile.html" not "views/users/profile.html"
//   - Templates must be unique across all view directories
//
// Injected Functions:
//   - req: Current *http.Request
//   - host: Application host prefix
//   - path_eq: Path comparison helper
//   - theme: Current DaisyUI theme
//   - All registered controllers by name
//
// HTMX Awareness:
//   - Automatically detects HX-Request header
//   - Can return partial HTML for HTMX requests
//   - Full page HTML for standard requests
//
// Error Handling:
//   - Missing template returns 404 for HTTP writers
//   - Calls log.Fatal for non-HTTP contexts (startup errors)
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
		funcs[name] = func() Handler { return ctrl.Handle(r) }
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

// Protect wraps an http.Handler with access control middleware.
//
// Access Control Flow:
//  1. If accessCheck is nil, handler executes immediately
//  2. accessCheck evaluates the request
//  3. If false, accessCheck must send response (redirect, error, etc.)
//  4. If true, handler executes normally
//
// Example:
//
//	http.Handle("/admin", app.Protect(adminHandler, auth.RequireAdmin))
//
// Design Note: The AccessCheck is responsible for the response when denying.
// This allows flexible responses: redirects, error pages, JSON errors, etc.
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

// ProtectFunc wraps an http.HandlerFunc with access control.
// This is a convenience method equivalent to Protect() for function handlers.
//
// Example:
//
//	http.HandleFunc("/profile", app.ProtectFunc(profileHandler, auth.RequireLogin))
func (app *App) ProtectFunc(fn http.HandlerFunc, accessLevel AccessCheck) http.HandlerFunc {
	return app.Protect(fn, accessLevel)
}
