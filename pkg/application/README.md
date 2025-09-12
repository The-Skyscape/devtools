# Application Package

The `application` package provides an opinionated MVC web framework that prioritizes server-side rendering and eliminates client-side state complexity through HTMX integration.

## Features

- 🚀 **Zero JavaScript Complexity** - Full interactivity with HTMX, no client-side state
- 🎯 **Request Isolation** - Value receiver pattern prevents data races
- 🔧 **Type-Safe Templates** - Controllers directly accessible in templates
- ⚡ **Performance Optimized** - Templates compiled once at startup
- 🔒 **Built-in Security** - CSRF protection, secure sessions, access control
- 🎨 **DaisyUI Themes** - Beautiful UI components out of the box

## Quick Start

```go
package main

import (
    "embed"
    "github.com/The-Skyscape/devtools/pkg/application"
    "myapp/controllers"
)

//go:embed views
var views embed.FS

func main() {
    application.Serve(views,
        application.WithController(controllers.Home()),
        application.WithDaisyTheme("corporate"),
    )
}
```

## Core Concepts

### Request Isolation Pattern

Each request gets its own controller copy through value receivers, eliminating concurrency issues:

```go
// Value receiver creates a copy for each request
func (c MyController) Handle(req *http.Request) application.Handler {
    c.Request = req  // Modifies the copy, not the original
    return &c        // Returns pointer to the copy
}
```

This pattern:
- ✅ Eliminates mutex requirements
- ✅ Prevents data races
- ✅ Costs only 16-32 bytes per request
- ✅ Enables safe concurrent handling

### Controller Pattern

Controllers must implement a factory function and the Handler interface:

```go
// Factory function returns name and controller instance
func MyController() (string, *MyController) {
    return "my", &MyController{}
}

type MyController struct {
    application.Controller
}

// Setup registers routes (called once at startup)
func (c *MyController) Setup(app *application.App) {
    c.Controller.Setup(app)  // Always call parent
    
    // Register routes
    http.Handle("GET /", app.Serve("home.html", nil))
    http.Handle("POST /create", app.ProtectFunc(c.create, auth.Required))
}

// Handle creates request-scoped instance (value receiver!)
func (c MyController) Handle(req *http.Request) application.Handler {
    c.Request = req
    return &c
}

// Public method - accessible in templates as {{my.GetData}}
func (c *MyController) GetData() string {
    return "data"
}

// Private handler - not accessible in templates
func (c *MyController) create(w http.ResponseWriter, r *http.Request) {
    c.SetRequest(r)
    // Handler logic
}
```

### Template Access

Controllers registered with a name become accessible in templates:

```html
<!-- Access controller methods -->
{{my.GetData}}

<!-- Conditional rendering -->
{{if auth.IsLoggedIn}}
    Welcome {{auth.CurrentUser.Name}}!
{{end}}

<!-- Iterate over data -->
{{range todos.AllTodos}}
    <div>{{.Title}}</div>
{{end}}
```

### HTMX Integration

The framework is HTMX-aware and handles requests appropriately:

```go
// In controller methods
func (c *MyController) handler(w http.ResponseWriter, r *http.Request) {
    // Process form
    
    // Trigger page refresh (HTMX-aware)
    c.Refresh(w, r)
    
    // Or redirect (HTMX-aware)
    c.Redirect(w, r, "/success")
}
```

HTMX features:
- Automatic detection via `HX-Request` header
- Returns partial HTML for HTMX requests
- Full page HTML for standard requests
- Errors return 200 OK with HTML content (required for HTMX swapping)

### Error Handling

Errors return 200 OK status with HTML content for HTMX compatibility:

```go
func (c *MyController) handler(w http.ResponseWriter, r *http.Request) {
    if err := validateInput(r); err != nil {
        c.RenderError(w, r, err)  // Returns 200 OK with error HTML
        return
    }
}
```

Error templates are automatically selected:
- `ValidationError` → `validation-errors.html`
- `ErrNotFound` → `error-404.html`
- `ErrForbidden` → `error-403.html`
- `ErrUnauthorized` → `error-401.html`
- Others → `error-message.html`

## Middleware

Add middleware to the application chain:

```go
application.Serve(views,
    application.WithMiddleware(middleware.NewLogger()),
    application.WithMiddleware(middleware.NewRateLimit(100)),
)
```

Built-in middleware:
- Logging
- Recovery
- Compression
- Security headers
- Rate limiting
- CORS

## Access Control

Protect routes with access checks:

```go
// In Setup()
http.Handle("GET /admin", app.Serve("admin.html", auth.RequireAdmin))
http.Handle("POST /api", app.ProtectFunc(c.api, auth.RequireUser))

// Custom access check
customCheck := func(app *App, w http.ResponseWriter, r *http.Request) bool {
    // Check access
    if !hasAccess {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return false
    }
    return true
}
```

## Configuration Options

```go
application.Serve(views,
    // Theme configuration
    application.WithDaisyTheme("corporate"),
    
    // Host prefix for reverse proxies
    application.WithHostPrefix("/app"),
    
    // Custom middleware
    application.WithMiddleware(customMiddleware),
    
    // Register controllers
    application.WithController("auth", authController),
    application.WithController(myController()),
)
```

## Performance Characteristics

- **Template Compilation**: Once at startup, not per request
- **Controller Copying**: 16-32 bytes per request
- **No Mutex Contention**: Value receiver isolation
- **Memory Efficient**: Stack allocation when possible
- **Concurrent Safe**: No shared mutable state

## Testing

```go
func TestController(t *testing.T) {
    app := application.New(testViews)
    ctrl := &MyController{}
    ctrl.Setup(app)
    
    // Test request handling
    req := httptest.NewRequest("GET", "/", nil)
    handler := ctrl.Handle(req)
    
    // Verify isolation
    if handler == ctrl {
        t.Error("Expected new instance")
    }
}
```

## Best Practices

1. **Always use value receivers for Handle()**
2. **Call parent Setup() in controllers**
3. **Use c.SetRequest(r) in handlers**
4. **Return errors as HTML for HTMX**
5. **Keep templates simple - logic in controllers**
6. **Use validation helpers for input**
7. **Public methods for templates, private for handlers**

## Why These Design Choices?

### Why 200 OK for Errors?
HTMX requires 200 OK to swap error content into the DOM. This is intentional design, not a bug.

### Why Value Receivers?
Eliminates entire categories of concurrency bugs without performance penalty.

### Why String-Based Controller Names?
Allows registration of user-defined types without complex interfaces.

### Why Compile Templates at Startup?
Trading hot-reload for guaranteed performance and early error detection.

## Examples

See the `examples/` directory for complete applications:
- **Blog**: Full-featured blog with authentication
- **API**: REST API with the framework
- **Realtime**: Server-Sent Events example

## License

MIT