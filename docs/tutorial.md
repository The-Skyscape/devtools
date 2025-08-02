# Tutorial: Building a Todo Application

This tutorial walks you through building a complete todo list application with TheSkyscape DevTools, covering authentication, database operations, and real-time updates with HTMX.

## What We'll Build

- ✅ User authentication (signup/signin/logout)
- ✅ Create, read, update, delete todos
- ✅ Mark todos as complete/incomplete
- ✅ Real-time updates without page refreshes
- ✅ Responsive UI with DaisyUI

## Prerequisites

- Go 1.24 or later
- Basic understanding of Go and HTML
- Docker (optional, for deployment)

## Step 1: Project Setup

Create a new project:

```bash
mkdir todo-app
cd todo-app
go mod init todo-app
go get github.com/The-Skyscape/devtools
```

Create the project structure:

```bash
mkdir -p models
mkdir -p controllers  
mkdir -p views
```

Your structure should look like:

```
todo-app/
├── controllers/
├── models/
├── views/
├── main.go
└── go.mod
```

## Step 2: Database Models

Create the database setup and Todo model:

**`models/database.go`**:
```go
package models

import (
	"github.com/The-Skyscape/devtools/pkg/authentication"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/local"
)

var (
	// DB is our SQLite database
	DB = local.Database("todos.db")
	
	// Auth manages user authentication
	Auth = authentication.Manage(DB)
	
	// Todos manages our todo items
	Todos = database.Manage(DB, new(Todo))
)
```

**`models/todo.go`**:
```go
package models

import (
	"time"
	"github.com/The-Skyscape/devtools/pkg/application"
)

// Todo represents a task in our todo list
type Todo struct {
	application.Model
	Title       string
	Description string
	Completed   bool
	DueDate     *time.Time
	UserID      string // Link todos to users
}

// Table returns the database table name
func (*Todo) Table() string {
	return "todos"
}

// Business logic methods

// GetUserTodos returns all todos for a specific user
func GetUserTodos(userID string) ([]*Todo, error) {
	return Todos.Search("UserID = ? ORDER BY CreatedAt DESC", userID)
}

// GetPendingTodos returns incomplete todos for a user
func GetPendingTodos(userID string) ([]*Todo, error) {
	return Todos.Search("UserID = ? AND Completed = ? ORDER BY CreatedAt DESC", userID, false)
}

// GetCompletedTodos returns completed todos for a user
func GetCompletedTodos(userID string) ([]*Todo, error) {
	return Todos.Search("UserID = ? AND Completed = ? ORDER BY UpdatedAt DESC", userID, true)
}

// ToggleComplete flips the completed status
func (t *Todo) ToggleComplete() error {
	t.Completed = !t.Completed
	t.UpdatedAt = time.Now()
	return Todos.Update(t)
}

// MarkComplete marks the todo as completed
func (t *Todo) MarkComplete() error {
	t.Completed = true
	t.UpdatedAt = time.Now()
	return Todos.Update(t)
}
```

## Step 3: Todo Controller

Create the controller to handle HTTP requests:

**`controllers/todos.go`**:
```go
package controllers

import (
	"net/http"
	"strconv"
	"time"
	"todo-app/models"
	"github.com/The-Skyscape/devtools/pkg/application"
)

// Todos returns the controller name and instance
func Todos() (string, *TodoController) {
	return "todos", &TodoController{}
}

// TodoController handles todo-related requests
type TodoController struct {
	application.BaseController
}

// Setup registers routes when the application starts
func (c *TodoController) Setup(app *application.App) {
	c.BaseController.Setup(app)
	
	// Main page
	http.Handle("GET /", app.Serve("dashboard.html", c.requireAuth))
	
	// Todo CRUD operations
	http.Handle("POST /todos", app.ProtectFunc(c.createTodo, false))
	http.Handle("PUT /todos/{id}/toggle", app.ProtectFunc(c.toggleTodo, false))
	http.Handle("DELETE /todos/{id}", app.ProtectFunc(c.deleteTodo, false))
	http.Handle("PUT /todos/{id}", app.ProtectFunc(c.updateTodo, false))
	
	// Partial views for HTMX
	http.Handle("GET /todos/list", app.Serve("todo-list.html", c.requireAuth))
}

// Handle returns controller instance for each request
func (c TodoController) Handle(r *http.Request) application.Controller {
	c.Request = r
	return &c
}

// requireAuth is an access check function
func (c *TodoController) requireAuth(app *application.App, r *http.Request) string {
	auth := app.Use("auth")
	if auth == nil {
		return "signin.html"
	}
	
	// Get auth controller and check if user is authenticated
	if user := auth.(*application.BaseController).Use("auth"); user == nil {
		return "signin.html" 
	}
	
	return "" // Allow access
}

// Template methods - accessible in views

// AllTodos returns all todos for the current user
func (c *TodoController) AllTodos() ([]*models.Todo, error) {
	user := c.getCurrentUser()
	if user == nil {
		return nil, nil
	}
	return models.GetUserTodos(user.ID)
}

// PendingTodos returns incomplete todos for the current user
func (c *TodoController) PendingTodos() ([]*models.Todo, error) {
	user := c.getCurrentUser()
	if user == nil {
		return nil, nil
	}
	return models.GetPendingTodos(user.ID)
}

// CompletedTodos returns completed todos for the current user
func (c *TodoController) CompletedTodos() ([]*models.Todo, error) {
	user := c.getCurrentUser()
	if user == nil {
		return nil, nil
	}
	return models.GetCompletedTodos(user.ID)
}

// HTTP handler methods

func (c *TodoController) createTodo(w http.ResponseWriter, r *http.Request) {
	user := c.getCurrentUser()
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDateStr := r.FormValue("due_date")
	
	if title == "" {
		c.Render(w, r, "error-message", "Title is required")
		return
	}
	
	todo := &models.Todo{
		Title:       title,
		Description: description,
		UserID:      user.ID,
		Completed:   false,
	}
	
	// Parse due date if provided
	if dueDateStr != "" {
		if dueDate, err := time.Parse("2006-01-02", dueDateStr); err == nil {
			todo.DueDate = &dueDate
		}
	}
	
	if _, err := models.Todos.Insert(todo); err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	
	// Return updated todo list
	c.Render(w, r, "todo-list.html", nil)
}

func (c *TodoController) toggleTodo(w http.ResponseWriter, r *http.Request) {
	user := c.getCurrentUser()
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	todoID := r.PathValue("id")
	todo, err := models.Todos.Get(todoID)
	if err != nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	
	// Check ownership
	if todo.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	if err := todo.ToggleComplete(); err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	
	// Return updated todo list
	c.Render(w, r, "todo-list.html", nil)
}

func (c *TodoController) updateTodo(w http.ResponseWriter, r *http.Request) {
	user := c.getCurrentUser()
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	todoID := r.PathValue("id")
	todo, err := models.Todos.Get(todoID)
	if err != nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	
	// Check ownership
	if todo.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	// Update fields
	if title := r.FormValue("title"); title != "" {
		todo.Title = title
	}
	if description := r.FormValue("description"); description != "" {
		todo.Description = description
	}
	
	if err := models.Todos.Update(todo); err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	
	c.Render(w, r, "todo-list.html", nil)
}

func (c *TodoController) deleteTodo(w http.ResponseWriter, r *http.Request) {
	user := c.getCurrentUser()
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	todoID := r.PathValue("id")
	todo, err := models.Todos.Get(todoID)
	if err != nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	
	// Check ownership
	if todo.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	if err := models.Todos.Delete(todo); err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	
	c.Render(w, r, "todo-list.html", nil)
}

// Helper method to get current authenticated user
func (c *TodoController) getCurrentUser() *models.User {
	// This would typically use the auth controller
	// For now, we'll implement a simple version
	auth := c.Use("auth")
	if auth == nil {
		return nil
	}
	
	// Get current user from auth controller
	// Implementation depends on your auth setup
	return nil // Placeholder
}
```

## Step 4: HTML Templates

Create the user interface with HTMX for dynamic updates:

**`views/dashboard.html`**:
```html
<!DOCTYPE html>
<html data-theme="{{theme}}" lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Todo App</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@4.4.19/dist/full.css" rel="stylesheet">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-base-200 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <!-- Header -->
        <div class="navbar bg-base-100 rounded-box shadow-lg mb-8">
            <div class="flex-1">
                <h1 class="text-xl font-bold">📝 My Todos</h1>
            </div>
            <div class="flex-none">
                {{if auth.CurrentUser}}
                    <div class="dropdown dropdown-end">
                        <div tabindex="0" role="button" class="btn btn-ghost">
                            👤 {{auth.CurrentUser.Name}}
                        </div>
                        <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-[1] w-52 p-2 shadow">
                            <li><a href="/profile">Profile</a></li>
                            <li>
                                <form method="POST" action="/_auth/signout">
                                    <button type="submit" class="w-full text-left">Logout</button>
                                </form>
                            </li>
                        </ul>
                    </div>
                {{else}}
                    <a href="/signin" class="btn btn-primary">Sign In</a>
                {{end}}
            </div>
        </div>

        {{if auth.CurrentUser}}
            <!-- Add Todo Form -->
            <div class="card bg-base-100 shadow-lg mb-8">
                <div class="card-body">
                    <h2 class="card-title">Add New Todo</h2>
                    <form hx-post="{{host}}/todos" hx-target="#todo-list" hx-swap="outerHTML" class="space-y-4">
                        <div class="form-control">
                            <label class="label">
                                <span class="label-text">Title</span>
                            </label>
                            <input name="title" type="text" placeholder="What do you need to do?" 
                                   class="input input-bordered w-full" required>
                        </div>
                        
                        <div class="form-control">
                            <label class="label">
                                <span class="label-text">Description (optional)</span>
                            </label>
                            <textarea name="description" placeholder="Add more details..." 
                                      class="textarea textarea-bordered"></textarea>
                        </div>
                        
                        <div class="form-control">
                            <label class="label">
                                <span class="label-text">Due Date (optional)</span>
                            </label>
                            <input name="due_date" type="date" class="input input-bordered">
                        </div>
                        
                        <div class="card-actions justify-end">
                            <button type="submit" class="btn btn-primary">
                                <span class="loading loading-spinner loading-sm htmx-indicator"></span>
                                Add Todo
                            </button>
                        </div>
                    </form>
                </div>
            </div>

            <!-- Todo List -->
            <div id="todo-list">
                {{template "todo-list.html"}}
            </div>
        {{else}}
            <!-- Not authenticated -->
            <div class="hero bg-base-100 rounded-box shadow-lg">
                <div class="hero-content text-center">
                    <div class="max-w-md">
                        <h1 class="text-5xl font-bold">Welcome!</h1>
                        <p class="py-6">Sign in to start managing your todos.</p>
                        <a href="/signin" class="btn btn-primary">Get Started</a>
                    </div>
                </div>
            </div>
        {{end}}
    </div>

    <!-- Error Toast -->
    <div id="toast" class="toast toast-top toast-end" style="display: none;">
        <div class="alert alert-error">
            <span id="toast-message"></span>
        </div>
    </div>

    <script>
        // Show toast on errors
        document.body.addEventListener('htmx:responseError', function(e) {
            const toast = document.getElementById('toast');
            const message = document.getElementById('toast-message');
            message.textContent = 'Something went wrong. Please try again.';
            toast.style.display = 'block';
            setTimeout(() => {
                toast.style.display = 'none';
            }, 3000);
        });
    </script>
</body>
</html>
```

**`views/todo-list.html`**:
```html
<div class="grid gap-6 md:grid-cols-2">
    <!-- Pending Todos -->
    <div class="card bg-base-100 shadow-lg">
        <div class="card-body">
            <h2 class="card-title">
                📋 Pending
                <div class="badge badge-secondary">{{len (todos.PendingTodos)}}</div>
            </h2>
            
            <div class="space-y-3">
                {{range todos.PendingTodos}}
                <div class="border border-base-300 rounded-lg p-4 hover:bg-base-50 transition-colors">
                    <div class="flex items-start justify-between">
                        <div class="flex items-start space-x-3 flex-1">
                            <button hx-put="{{host}}/todos/{{.ID}}/toggle" 
                                    hx-target="#todo-list" 
                                    hx-swap="outerHTML"
                                    class="checkbox">
                            </button>
                            <div class="flex-1">
                                <h3 class="font-semibold">{{.Title}}</h3>
                                {{if .Description}}
                                    <p class="text-sm text-base-content/70 mt-1">{{.Description}}</p>
                                {{end}}
                                {{if .DueDate}}
                                    <p class="text-xs text-warning mt-1">Due: {{.DueDate.Format "Jan 2, 2006"}}</p>
                                {{end}}
                                <p class="text-xs text-base-content/50 mt-1">Created {{.CreatedAt.Format "Jan 2, 15:04"}}</p>
                            </div>
                        </div>
                        <div class="dropdown dropdown-end">
                            <div tabindex="0" role="button" class="btn btn-ghost btn-sm">⋮</div>
                            <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-[1] w-32 p-2 shadow">
                                <li>
                                    <button hx-delete="{{host}}/todos/{{.ID}}" 
                                            hx-target="#todo-list" 
                                            hx-swap="outerHTML"
                                            hx-confirm="Are you sure?"
                                            class="text-error">
                                        Delete
                                    </button>
                                </li>
                            </ul>
                        </div>
                    </div>
                </div>
                {{else}}
                <div class="text-center py-8 text-base-content/50">
                    <p>No pending todos! 🎉</p>
                    <p class="text-sm">Add one above to get started.</p>
                </div>
                {{end}}
            </div>
        </div>
    </div>

    <!-- Completed Todos -->
    <div class="card bg-base-100 shadow-lg">
        <div class="card-body">
            <h2 class="card-title">
                ✅ Completed
                <div class="badge badge-success">{{len (todos.CompletedTodos)}}</div>
            </h2>
            
            <div class="space-y-3">
                {{range todos.CompletedTodos}}
                <div class="border border-base-300 rounded-lg p-4 hover:bg-base-50 transition-colors opacity-60">
                    <div class="flex items-start justify-between">
                        <div class="flex items-start space-x-3 flex-1">
                            <button hx-put="{{host}}/todos/{{.ID}}/toggle" 
                                    hx-target="#todo-list" 
                                    hx-swap="outerHTML"
                                    class="checkbox checkbox-success" checked>
                            </button>
                            <div class="flex-1">
                                <h3 class="font-semibold line-through">{{.Title}}</h3>
                                {{if .Description}}
                                    <p class="text-sm text-base-content/70 mt-1 line-through">{{.Description}}</p>
                                {{end}}
                                <p class="text-xs text-base-content/50 mt-1">Completed {{.UpdatedAt.Format "Jan 2, 15:04"}}</p>
                            </div>
                        </div>
                        <div class="dropdown dropdown-end">
                            <div tabindex="0" role="button" class="btn btn-ghost btn-sm">⋮</div>
                            <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-[1] w-32 p-2 shadow">
                                <li>
                                    <button hx-delete="{{host}}/todos/{{.ID}}" 
                                            hx-target="#todo-list" 
                                            hx-swap="outerHTML"
                                            hx-confirm="Are you sure?"
                                            class="text-error">
                                        Delete
                                    </button>
                                </li>
                            </ul>
                        </div>
                    </div>
                </div>
                {{else}}
                <div class="text-center py-8 text-base-content/50">
                    <p>No completed todos yet.</p>
                </div>
                {{end}}
            </div>
        </div>
    </div>
</div>
```

## Step 5: Main Application

Create the main application entry point:

**`main.go`**:
```go
package main

import (
	"embed"
	"os"
	"todo-app/controllers"
	"todo-app/models"
	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/authentication"
)

//go:embed all:views
var views embed.FS

func main() {
	// Create authentication controller with options
	auth := models.Auth.Controller(
		authentication.WithCookie("todo-session"),
		authentication.WithSignoutURL("/"),
	)

	// Start the application
	application.Serve(views,
		application.WithController("auth", auth),
		application.WithController(controllers.Todos()),
		application.WithDaisyTheme(getEnv("THEME", "corporate")),
		application.WithHostPrefix(os.Getenv("HOST_PREFIX")),
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

## Step 6: Environment Setup

Create a `.env` file (don't commit this to git):

```bash
# Required
AUTH_SECRET=your-super-secret-jwt-key-change-this-in-production

# Optional
PORT=8080
THEME=corporate
HOST_PREFIX=
DEBUG=true
```

## Step 7: Running the Application

1. **Set environment variables**:
   ```bash
   export AUTH_SECRET="your-super-secret-jwt-key"
   export THEME="corporate"
   ```

2. **Run the application**:
   ```bash
   go run .
   ```

3. **Visit** `http://localhost:5000` (or the port you specified)

## Step 8: Testing the Application

1. **Sign up** for a new account
2. **Add some todos** with different titles and descriptions
3. **Mark todos as complete** by clicking the checkboxes
4. **Delete todos** using the dropdown menu
5. **Notice** how the page updates without refreshing (thanks to HTMX!)

## Next Steps

Now that you have a working todo application, you can extend it with:

### Features to Add

1. **Categories/Tags**: Add a category field to todos
2. **Due Date Reminders**: Highlight overdue todos
3. **Bulk Operations**: Select multiple todos for bulk actions
4. **Search/Filter**: Add search functionality
5. **Todo Sharing**: Share todos with other users
6. **Dark Mode Toggle**: Let users switch themes
7. **File Attachments**: Attach files to todos

### Technical Improvements

1. **Add Tests**: Write unit and integration tests
2. **Add Validation**: Better form validation and error handling
3. **Add Pagination**: For users with many todos
4. **Add Caching**: Cache frequently accessed data
5. **Add Logging**: Better application logging
6. **Add Metrics**: Monitor application performance

### Deployment

1. **Docker**: Containerize the application
2. **Cloud**: Deploy to DigitalOcean/AWS using the hosting package
3. **CI/CD**: Set up automated deployments
4. **Monitoring**: Add health checks and monitoring

## Conclusion

You've built a complete todo application using TheSkyscape DevTools! This tutorial covered:

- ✅ Project structure and organization
- ✅ Database models with relationships
- ✅ HTTP controllers with proper routing
- ✅ Real-time UI updates with HTMX
- ✅ User authentication and authorization
- ✅ Responsive design with DaisyUI

The patterns you've learned here can be applied to build much more complex applications. The TheSkyscape DevTools provides the foundation for authentication, database management, and deployment - you focus on your business logic.

Happy coding! 🚀