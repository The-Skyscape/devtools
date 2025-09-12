// Package application provides an opinionated MVC web framework that prioritizes
// server-side rendering and eliminates client-side state complexity through HTMX integration.
//
// # Core Philosophy
//
// This framework rejects the complexity of modern JavaScript frameworks in favor of:
//   - HTMX for dynamic updates without client-side state
//   - Server-side rendering as the source of truth
//   - Type safety over runtime flexibility
//   - Explicit patterns over magic
//   - Zero JavaScript complexity
//
// # Architecture
//
// Request Isolation:
// Each HTTP request receives its own controller instance through a value receiver pattern.
// This eliminates mutex requirements and prevents data races between concurrent requests.
// The pattern copies only 16-32 bytes per request - less than a CPU cache line.
//
// Template Integration:
// Controllers are directly accessible in templates through their registered names.
// Public methods (capitalized) can be called from templates, while private methods
// (lowercase) serve as HTTP handlers.
//
// State Management:
// All application state lives on the server. There is no client-side state to synchronize,
// no cache invalidation problems, and no version mismatches between API and client.
//
// # Quick Start
//
//	//go:embed all:views
//	var views embed.FS
//
//	func main() {
//	    application.Serve(views,
//	        application.WithController(controllers.Home()),
//	        application.WithController("auth", auth.Controller()),
//	        application.WithDaisyTheme("corporate"),
//	    )
//	}
//
// # Controller Pattern
//
// Controllers must implement a factory function and the Handler interface:
//
//	func MyController() (string, *MyController) {
//	    return "my", &MyController{}
//	}
//
//	type MyController struct {
//	    application.Controller
//	}
//
//	// Value receiver for request isolation
//	func (c MyController) Handle(req *http.Request) application.Handler {
//	    c.Request = req  // Modifies the copy
//	    return &c        // Returns pointer to copy
//	}
//
// # Template Access
//
// Controllers registered with a name become accessible in templates:
//
//	{{my.PublicMethod}}        // Calls controller's PublicMethod
//	{{if my.IsLoggedIn}}...{{end}}  // Conditional based on controller method
//
// # HTMX Integration
//
// The framework automatically detects HTMX requests and responds appropriately:
//   - HX-Request header triggers partial HTML responses
//   - Errors return 200 OK with HTML content (required for HTMX swapping)
//   - c.Refresh() triggers full page refresh via HX-Refresh header
//   - c.Redirect() uses HX-Location for client-side navigation
//
// # Design Decisions
//
// Why 200 OK for Errors:
// HTMX requires 200 OK to swap error content into the DOM. This is intentional,
// not a bug. Error templates are rendered as HTML fragments.
//
// Why String-Based Dependency Injection:
// The Use() method accepts strings to allow user-defined controller types
// without requiring interface definitions or type registration.
//
// Why Templates Compiled at Startup:
// Templates are parsed and compiled once during app.Start() for performance.
// This trades hot-reloading for zero per-request compilation overhead.
//
// # Thread Safety
//
// The framework ensures thread safety through:
//   - Controllers initialized once at startup (single-threaded)
//   - Request isolation via value receivers (no shared state)
//   - Immutable template compilation (prepared once, read-only after)
//   - Database operations use connection pooling (SQLite handles concurrency)
package application