# DevTools Example Application

This example demonstrates the **correct patterns** for building applications with the TheSkyscape DevTools MVC framework. It's designed to be clear for both humans and AI coding assistants.

## 🎯 Purpose

This example shows:
- ✅ Proper MVC structure with server-side rendering
- ✅ Type-safe controller callbacks in templates  
- ✅ HTMX for dynamic interactions without client state
- ✅ Database models with embedded composition
- ✅ Authentication integration

## 📁 Structure

```
example/
├── main.go                 # Application entry point
├── controllers/
│   └── ducks.go           # Example controller with CRUD operations
├── models/
│   ├── database.go        # Database setup and collections
│   └── duck.go            # Model with embedded application.Model
└── views/
    ├── dashboard.html     # Main page with HTMX forms
    └── includes.html      # Shared template includes
```

## 🔑 Key Patterns for AI Assistants

### 1. Controller Pattern

**ALWAYS** follow this exact pattern for controllers:

```go
// PATTERN: Factory function returns (prefix, instance)
func ControllerName() (string, *ControllerNameController) {
    return "prefix", &ControllerNameController{}
}

type ControllerNameController struct {
    application.Controller  // ALWAYS embed Controller, never BaseController
}

// PATTERN: Setup registers routes - called once at startup
func (c *ControllerNameController) Setup(app *application.App) {
    c.Controller.Setup(app)  // ALWAYS call parent Setup first
    
    // Register routes using app methods
    http.Handle("GET /", app.Serve("template.html", nil))
    http.Handle("POST /action", app.ProtectFunc(c.handler, nil))
}

// PATTERN: Handle uses VALUE receiver for request isolation
func (c ControllerNameController) Handle(req *http.Request) application.Handler {
    c.Request = req  // Modifies the copy
    return &c        // Returns pointer to the copy
}

// PATTERN: Public methods are accessible in templates
func (c *ControllerNameController) GetData() ([]Model, error) {
    // This can be called in templates as: {{prefix.GetData}}
}

// PATTERN: Private handlers for POST/PUT/DELETE
func (c *ControllerNameController) handler(w http.ResponseWriter, r *http.Request) {
    c.SetRequest(r)  // ALWAYS set request first
    
    // Handle the request
    if err != nil {
        c.RenderError(w, r, err)  // ALWAYS show errors to user
        return
    }
    
    c.Refresh(w, r)  // Use for HTMX page refresh
}
```

### 2. Model Pattern

**ALWAYS** follow this exact pattern for models:

```go
type ModelName struct {
    application.Model  // ALWAYS embed Model for ID, CreatedAt, UpdatedAt
    
    // Your fields here
    Name  string
    Count int
}

// REQUIRED: Table method returns database table name
func (*ModelName) Table() string {
    return "model_names"  // Plural, snake_case
}
```

### 3. Database Setup Pattern

In `models/database.go`:

```go
var (
    DB = local.Database("app.db")
    
    // Collections - use database.Manage for typed repositories
    ModelNames = database.Manage(DB, new(ModelName))
)
```

### 4. Template Pattern

Templates can call controller methods directly:

```html
<!-- Controller methods are available by prefix -->
{{range ducks.AllDucks}}
    <div>{{.Name}}</div>
{{end}}

<!-- HTMX for interactions - NO client state -->
<form hx-post="/action" hx-swap="none">
    <input name="field" />
    <button type="submit">Submit</button>
</form>

<!-- Use hx-swap="none" when using Refresh() -->
<!-- Use hx-target for partial updates -->
```

## ⚠️ Common AI Mistakes to Avoid

### ❌ DON'T Do These:

1. **Don't use BaseController** - It's called `Controller` now
2. **Don't use pointer receiver for Handle()** - Use value receiver
3. **Don't use snake_case in SQL** - Use PascalCase: `WHERE UserID = ?`
4. **Don't add client-side state** - All state lives on server
5. **Don't create new error templates** - Use `c.RenderError()`
6. **Don't forget to call `c.SetRequest(r)`** in handlers
7. **Don't use `http.Redirect`** - Use `c.Redirect()` for HTMX
8. **Don't ignore errors** - Always render them to the user

### ✅ DO These:

1. **DO embed application.Controller** in every controller
2. **DO use value receiver for Handle()** method
3. **DO call parent Setup()** first in Setup method
4. **DO use PascalCase in SQL** queries
5. **DO use c.RenderError()** for all errors
6. **DO use c.Refresh()** for HTMX refreshes
7. **DO make template methods public** (capitalized)
8. **DO use app.ProtectFunc()** for protected routes

## 🚀 Running the Example

```bash
# From the example directory
export AUTH_SECRET="example-secret"
go run .

# Visit http://localhost:5000
```

## 📝 How Controllers Work

1. **Factory Function**: Returns prefix and instance
   - Prefix becomes the template function name
   - Example: `"ducks"` prefix → `{{ducks.Method}}` in templates

2. **Setup Method**: Called once at app startup
   - Registers HTTP routes
   - Initializes services
   - Called in registration order

3. **Handle Method**: Called for EACH request
   - Uses VALUE receiver (creates copy)
   - Each request gets its own controller instance
   - No shared state between requests

4. **Template Methods**: Public methods accessible in templates
   - Must be exported (capitalized)
   - Can return any type templates can handle
   - Called directly: `{{controller.Method}}`

5. **Handler Methods**: Private methods for HTTP handlers
   - Use `c.SetRequest(r)` to set current request
   - Use `c.RenderError()` for errors
   - Use `c.Refresh()` for HTMX updates

## 🔄 HTMX Integration

The framework is built for HTMX:

- `c.Refresh()` - Triggers full page refresh
- `c.Redirect()` - HTMX-aware redirect
- `c.IsHTMX()` - Check if request is from HTMX
- Templates return HTML fragments for swapping

No JavaScript state management needed!

## 🗄️ Database Operations

```go
// Insert
duck, err := models.Ducks.Insert(&models.Duck{Name: "Donald"})

// Get by ID
duck, err := models.Ducks.Get(id)

// Search with SQL
ducks, err := models.Ducks.Search("WHERE Name LIKE ? ORDER BY CreatedAt", "%donald%")

// Update
duck.Name = "Daffy"
err := models.Ducks.Update(duck)

// Delete
err := models.Ducks.Delete(duck)

// Count
count := models.Ducks.Count("WHERE Breed = ?", "mallard")
```

## 🔐 Authentication

The example includes authentication setup:

```go
// In main.go
auth := models.Auth.Controller(
    authentication.WithCookie("token-name"),
    authentication.WithSignoutURL("/"),
)

// Use in controllers for protected routes
http.Handle("POST /admin", app.ProtectFunc(c.adminHandler, auth.Required))
```

## 📚 Learn More

- See `AI_PATTERNS.md` for complete pattern reference
- Check the workbench app for production examples
- Read CLAUDE.md for framework philosophy

## Remember

**This framework is designed for AI assistants.** The patterns are intentionally explicit and consistent. When an AI follows these patterns exactly, it produces working code every time.