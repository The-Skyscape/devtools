# AI Assistant Pattern Reference

This document provides copy-paste ready patterns for AI coding assistants working with the DevTools framework.

## 🎯 Framework Philosophy

1. **Server-side rendering only** - No client JavaScript state
2. **Controllers handle logic** - Templates only display
3. **Type-safe template callbacks** - Controllers available in templates
4. **HTMX for interactivity** - Server returns HTML, not JSON
5. **Embedded composition** - Not inheritance

## 📋 Copy-Paste Patterns

### Complete Controller Pattern

```go
package controllers

import (
    "net/http"
    "github.com/YourOrg/yourapp/models"
    "github.com/The-Skyscape/devtools/pkg/application"
)

// ResourceName returns the controller prefix and instance
// PATTERN: Factory function always returns (string, *Controller)
func ResourceName() (string, *ResourceNameController) {
    return "resource", &ResourceNameController{}
}

// ResourceNameController handles resource operations
type ResourceNameController struct {
    application.Controller  // ALWAYS embed Controller
}

// Setup registers routes and initializes the controller
// PATTERN: Called once at application startup
func (c *ResourceNameController) Setup(app *application.App) {
    c.Controller.Setup(app)  // ALWAYS call parent Setup first
    
    // GET routes - typically render pages
    http.Handle("GET /resources", app.Serve("resources-list.html", nil))
    http.Handle("GET /resources/new", app.Serve("resources-new.html", nil))
    
    // POST routes - typically modify data then refresh
    http.Handle("POST /resources", app.ProtectFunc(c.createResource, nil))
    http.Handle("POST /resources/{id}/delete", app.ProtectFunc(c.deleteResource, nil))
}

// Handle creates a request-scoped controller instance
// PATTERN: MUST use value receiver (not pointer)
func (c ResourceNameController) Handle(req *http.Request) application.IController {
    c.Request = req  // Modifies the copy
    return &c        // Returns pointer to the copy
}

// GetResources is accessible in templates as {{resource.GetResources}}
// PATTERN: Public methods (capitalized) are template-accessible
func (c *ResourceNameController) GetResources() ([]*models.Resource, error) {
    return models.Resources.Search("ORDER BY CreatedAt DESC")
}

// GetResource gets a single resource by ID from URL
func (c *ResourceNameController) GetResource() (*models.Resource, error) {
    id := c.Params().Int("id", 0)
    if id == 0 {
        return nil, application.ErrNotFound
    }
    return models.Resources.Get(id)
}

// createResource handles POST /resources
// PATTERN: Private methods (lowercase) are HTTP handlers
func (c *ResourceNameController) createResource(w http.ResponseWriter, r *http.Request) {
    c.SetRequest(r)  // ALWAYS set request first
    
    // Validate input
    validator := c.Validator()
    validator.CheckRequired("name", r.FormValue("name"))
    validator.CheckLength("description", r.FormValue("description"), 0, 500)
    
    if err := validator.Result(); err != nil {
        c.RenderError(w, r, err)  // ALWAYS render errors
        return
    }
    
    // Create resource
    resource := &models.Resource{
        Name:        r.FormValue("name"),
        Description: r.FormValue("description"),
    }
    
    if _, err := models.Resources.Insert(resource); err != nil {
        c.RenderError(w, r, err)
        return
    }
    
    // Refresh page for HTMX
    c.Refresh(w, r)  // Use Refresh for HTMX updates
}

// deleteResource handles POST /resources/{id}/delete
func (c *ResourceNameController) deleteResource(w http.ResponseWriter, r *http.Request) {
    c.SetRequest(r)
    
    id := c.Params().Int("id", 0)
    resource, err := models.Resources.Get(id)
    if err != nil {
        c.RenderError(w, r, application.ErrNotFound)
        return
    }
    
    if err := models.Resources.Delete(resource); err != nil {
        c.RenderError(w, r, err)
        return
    }
    
    c.Redirect(w, r, "/resources")  // Use Redirect for navigation
}
```

### Complete Model Pattern

```go
package models

import "github.com/The-Skyscape/devtools/pkg/application"

// Resource represents a resource in the database
// PATTERN: Always embed application.Model
type Resource struct {
    application.Model  // Provides ID, CreatedAt, UpdatedAt
    
    // Add your fields here
    Name        string
    Description string
    UserID      int
    Status      string
}

// Table returns the database table name
// PATTERN: Required method for all models
func (*Resource) Table() string {
    return "resources"  // Plural, snake_case
}

// User returns the associated user
// PATTERN: Relationship methods are on the model
func (r *Resource) User() (*User, error) {
    return Users.Get(r.UserID)
}

// IsActive checks if the resource is active
// PATTERN: Business logic methods on model
func (r *Resource) IsActive() bool {
    return r.Status == "active"
}
```

### Database Setup Pattern

```go
package models

import (
    "github.com/The-Skyscape/devtools/pkg/authentication"
    "github.com/The-Skyscape/devtools/pkg/database"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

var (
    // DB is the application database
    DB = local.Database("app.db")
    
    // Auth manages authentication
    Auth = authentication.Manage(DB)
    
    // Collections - PATTERN: Use database.Manage for typed repositories
    Resources = database.Manage(DB, new(Resource))
    Users     = database.Manage(DB, new(User))
)
```

### Template Patterns

```html
<!-- Layout Template -->
<!DOCTYPE html>
<html data-theme="{{theme}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@4.6.0/dist/full.min.css" rel="stylesheet">
    <script src="https://cdn.tailwindcss.com"></script>
    <title>{{.Title}}</title>
</head>
<body>
    {{template "content" .}}
</body>
</html>

<!-- List Template -->
{{define "content"}}
<div class="container mx-auto p-4">
    <h1 class="text-2xl font-bold mb-4">Resources</h1>
    
    <!-- PATTERN: Controller methods available by prefix -->
    {{range resource.GetResources}}
    <div class="card bg-base-100 shadow-xl mb-4">
        <div class="card-body">
            <h2 class="card-title">{{.Name}}</h2>
            <p>{{.Description}}</p>
            
            <!-- PATTERN: HTMX for deletions -->
            <form hx-post="/resources/{{.ID}}/delete" 
                  hx-confirm="Are you sure?"
                  hx-swap="none">
                <button type="submit" class="btn btn-error btn-sm">Delete</button>
            </form>
        </div>
    </div>
    {{else}}
    <p>No resources found</p>
    {{end}}
    
    <!-- PATTERN: HTMX form with validation -->
    <form hx-post="/resources" hx-swap="none" class="mt-4">
        <div class="form-control">
            <label class="label">
                <span class="label-text">Name</span>
            </label>
            <input type="text" name="name" class="input input-bordered" required>
        </div>
        
        <div class="form-control">
            <label class="label">
                <span class="label-text">Description</span>
            </label>
            <textarea name="description" class="textarea textarea-bordered"></textarea>
        </div>
        
        <button type="submit" class="btn btn-primary mt-4">Create Resource</button>
    </form>
</div>
{{end}}
```

### Main.go Pattern

```go
package main

import (
    "embed"
    "github.com/The-Skyscape/devtools/pkg/application"
    "yourapp/controllers"
    "yourapp/models"
)

//go:embed all:views
var views embed.FS

func main() {
    // Setup authentication
    auth := models.Auth.Controller()
    
    // Start application with controllers
    application.Serve(views,
        application.WithDaisyTheme("corporate"),
        application.WithController("auth", auth),
        application.WithController(controllers.ResourceName()),
        // Add more controllers here
    )
}
```

## ❌ Anti-Patterns to Avoid

### DON'T: Use pointer receiver for Handle

```go
// ❌ WRONG - Pointer receiver breaks request isolation
func (c *Controller) Handle(req *http.Request) application.IController {
    c.Request = req  // Modifies shared instance!
    return c
}

// ✅ CORRECT - Value receiver creates copy
func (c Controller) Handle(req *http.Request) application.IController {
    c.Request = req  // Modifies the copy
    return &c
}
```

### DON'T: Use snake_case in SQL

```go
// ❌ WRONG - Snake case
resources, err := models.Resources.Search("WHERE user_id = ?", userID)

// ✅ CORRECT - PascalCase
resources, err := models.Resources.Search("WHERE UserID = ?", userID)
```

### DON'T: Ignore errors

```go
// ❌ WRONG - Silent failure
resource, _ := models.Resources.Get(id)

// ✅ CORRECT - Handle error
resource, err := models.Resources.Get(id)
if err != nil {
    c.RenderError(w, r, err)
    return
}
```

### DON'T: Use http.Redirect directly

```go
// ❌ WRONG - Breaks HTMX
http.Redirect(w, r, "/path", http.StatusSeeOther)

// ✅ CORRECT - HTMX aware
c.Redirect(w, r, "/path")
```

### DON'T: Create custom error templates

```go
// ❌ WRONG - Custom error handling
c.Render(w, r, "my-error.html", err)

// ✅ CORRECT - Use framework error handling
c.RenderError(w, r, err)
```

## 🔍 SQL Query Patterns

```go
// Get single record
resource, err := models.Resources.Get(id)

// Find first matching record
resource, err := models.Resources.Find("WHERE Name = ?", name)

// Search multiple records
resources, err := models.Resources.Search("WHERE Status = ? ORDER BY CreatedAt DESC", "active")

// Count records
count := models.Resources.Count("WHERE UserID = ?", userID)

// Insert new record
resource, err := models.Resources.Insert(&models.Resource{
    Name: "Example",
})

// Update existing record
resource.Name = "Updated"
err := models.Resources.Update(resource)

// Delete record
err := models.Resources.Delete(resource)

// IMPORTANT: All field names in SQL are PascalCase!
// UserID not user_id
// CreatedAt not created_at
// IsActive not is_active
```

## 🎨 Template Functions Available

Built-in functions you can use in templates:

```html
<!-- Formatting -->
{{formatBytes 1024}}        <!-- "1.0 KB" -->
{{formatPrice 19.99}}       <!-- "$19.99" -->
{{formatPercent 0.156}}     <!-- "15.6%" -->
{{formatDate .CreatedAt}}   <!-- "Jan 2, 2006" -->
{{timeAgo .UpdatedAt}}      <!-- "3 hours ago" -->

<!-- String manipulation -->
{{truncate .Description 100}}  <!-- Truncate to 100 chars -->
{{pluralize 1 "item"}}         <!-- "item" -->
{{pluralize 5 "item"}}         <!-- "items" -->
{{lower .Name}}                <!-- lowercase -->
{{upper .Code}}                <!-- UPPERCASE -->
{{title .Name}}                <!-- Title Case -->

<!-- Math -->
{{add 1 2}}                    <!-- 3 -->
{{sub 10 3}}                   <!-- 7 -->
{{mul 4 5}}                    <!-- 20 -->
{{div 10 2}}                   <!-- 5 -->

<!-- Utilities -->
{{default .Name "Unknown"}}    <!-- Use default if empty -->
{{coalesce .Nick .Name .Email}} <!-- First non-empty value -->
```

## 🚦 HTTP Status Codes

The framework handles status codes automatically:

- **200 OK** - Normal responses and HTMX content swaps
- **204 No Content** - HTMX redirects and refreshes
- **404 Not Found** - Template not found
- **500 Internal Server Error** - Template execution errors

Don't set status codes manually unless you have a specific need.

## 🔐 Authentication Patterns

```go
// In controller Setup
http.Handle("POST /admin", app.ProtectFunc(c.adminAction, auth.Required))

// In template
{{if auth.CurrentUser}}
    <p>Welcome {{auth.CurrentUser.Email}}</p>
{{else}}
    <a href="/signin">Sign In</a>
{{end}}

// Custom access check
var adminOnly = func(app *application.App, w http.ResponseWriter, r *http.Request) bool {
    user := auth.CurrentUser()
    if user == nil || !user.IsAdmin {
        c.Redirect(w, r, "/signin")
        return false
    }
    return true
}
```

## 📝 Remember

1. **Controllers are stateless** - Each request gets a fresh copy
2. **Templates are dumb** - Logic belongs in controllers
3. **HTMX handles interactivity** - No JavaScript state needed
4. **Errors always render** - Never fail silently
5. **SQL uses PascalCase** - Match Go struct fields
6. **Embed, don't inherit** - Use composition

When in doubt, look at the example application or the workbench for real patterns!