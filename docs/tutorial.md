# Tutorial: Building a Todo Application

This tutorial walks you through building a complete todo list application with TheSkyscape DevTools, covering authentication, database operations, and real-time updates with HTMX.

## What We'll Build

- ✅ User authentication (signup/signin/logout)
- ✅ Create, read, update, delete todos
- ✅ Mark todos as complete/incomplete
- ✅ Real-time updates without page refreshes
- ✅ Responsive UI with DaisyUI

## Prerequisites

- Go 1.21 or later
- Basic understanding of Go and HTML
- Docker (optional, for deployment)

## Step 1: Project Setup

### Option 1: Using create-app CLI (Recommended)

```bash
# Download the CLI tool
curl -L -o create-app https://github.com/The-Skyscape/devtools/releases/download/v1.0.1/create-app
chmod +x create-app

# Create a new todo application
./create-app my-todo-app
cd my-todo-app
export AUTH_SECRET="your-secret"
go run .
```

That's it! The `create-app` tool generates a fully working todo application with all the patterns described in this tutorial.

### Option 2: Manual Setup

Create a new project manually:

```bash
mkdir todo-app
cd todo-app
go mod init todo-app
go get github.com/The-Skyscape/devtools
```

Create the project structure:

```bash
mkdir -p models controllers views/partials
```

Your structure should look like:

```
todo-app/
├── controllers/
│   ├── home.go
│   └── todos.go
├── models/
│   ├── database.go
│   └── todo.go
├── views/
│   ├── home.html
│   ├── todos.html
│   ├── layout.html
│   └── partials/
│       ├── layout.html
│       ├── todos-item.html
│       └── error-message.html
├── main.go
└── go.mod
```

## Step 2: Database Models

Create the database setup and Todo model:

**`models/database.go`**:
```go
package models

import (
	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/authentication"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/database/local"
)

var (
	DB    = local.Database("app.db")
	Auth  = authentication.Manage(DB)
	Todos = database.Manage(DB, new(Todo))
)
```

**`models/todo.go`**:
```go
package models

import "github.com/The-Skyscape/devtools/pkg/application"

type Todo struct {
	application.Model
	Title     string
	Completed bool
}

func (*Todo) Table() string { return "todos" }
```

## Step 3: Controllers

Create controllers to handle HTTP requests:

**`controllers/home.go`**:
```go
package controllers

import (
	"net/http"
	"github.com/The-Skyscape/devtools/pkg/application"
)

func Home() (string, *HomeController) {
	return "home", &HomeController{}
}

type HomeController struct {
	application.BaseController
}

func (c *HomeController) Setup(app *application.App) {
	c.BaseController.Setup(app)
	app.Serve("GET /", "home.html", nil)
}

func (c *HomeController) Handle(r *http.Request) application.Controller {
	c.Request = r
	return c
}
```

**`controllers/todos.go`**:
```go
package controllers

import (
	"errors"
	"net/http"
	"your-app/models"
	"github.com/The-Skyscape/devtools/pkg/application"
)

func Todos() (string, *TodosController) {
	return "todos", &TodosController{}
}

type TodosController struct {
	application.BaseController
}

func (c *TodosController) Setup(app *application.App) {
	c.BaseController.Setup(app)
	app.Serve("GET /todos", "todos.html", models.Auth.Required)
	app.ProtectFunc("POST /todos", c.createTodo, models.Auth.Required)
	app.ProtectFunc("PUT /todos/{id}/toggle", c.toggleTodo, models.Auth.Required)
	app.ProtectFunc("DELETE /todos/{id}", c.deleteTodo, models.Auth.Required)
}

func (c *TodosController) Handle(r *http.Request) application.Controller {
	c.Request = r
	return c
}

// Template method accessible in views
func (c *TodosController) AllTodos() ([]*models.Todo, error) {
	return models.Todos.Search("")
}

func (c *TodosController) createTodo(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		c.Render(w, r, "error-message.html", errors.New("title is required"))
		return
	}

	todo := &models.Todo{Title: title}
	if err := models.Todos.Insert(todo); err != nil {
		c.Render(w, r, "error-message.html", err)
		return
	}

	c.Refresh(w, r)
}

func (c *TodosController) toggleTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	todo, err := models.Todos.Get(id)
	if err != nil {
		c.Render(w, r, "error-message.html", err)
		return
	}

	todo.Completed = !todo.Completed
	if err := models.Todos.Update(todo); err != nil {
		c.Render(w, r, "error-message.html", err)
		return
	}

	c.Refresh(w, r)
}

func (c *TodosController) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	todo, err := models.Todos.Get(id)
	if err != nil {
		c.Render(w, r, "error-message.html", err)
		return
	}

	if err := models.Todos.Delete(todo); err != nil {
		c.Render(w, r, "error-message.html", err)
		return
	}

	c.Refresh(w, r)
}
```

## Step 4: HTML Templates

Create the user interface with HTMX for dynamic updates:

**`views/layout.html`**:
```html
{{define "layout/start"}}
<!DOCTYPE html>
<html data-theme="{{theme}}" lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Todo App - TheSkyscape</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@4.4.19/dist/full.css" rel="stylesheet">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-base-200 min-h-screen">
    <div class="container mx-auto px-4 py-8">
{{end}}

{{define "layout/end"}}
    </div>
</body>
</html>
{{end}}
```

**`views/home.html`**:
```html
{{template "layout/start"}}

<div class="hero bg-base-100 rounded-box shadow-lg">
    <div class="hero-content text-center">
        <div class="max-w-md">
            <h1 class="text-5xl font-bold">📝 Todo App</h1>
            <p class="py-6">A simple todo application built with TheSkyscape DevTools.</p>
            
            {{if auth.CurrentUser}}
                <a href="/todos" class="btn btn-primary">View My Todos</a>
                <form method="POST" action="/_auth/signout" class="mt-4">
                    <button type="submit" class="btn btn-outline">Sign Out</button>
                </form>
            {{else}}
                <a href="/_auth/signin" class="btn btn-primary">Sign In</a>
                <a href="/_auth/signup" class="btn btn-outline ml-2">Sign Up</a>
            {{end}}
        </div>
    </div>
</div>

{{template "layout/end"}}
```

**`views/todos.html`**:
```html
{{template "layout/start"}}

<div class="navbar bg-base-100 rounded-box shadow-lg mb-8">
    <div class="flex-1">
        <h1 class="text-xl font-bold">📝 My Todos</h1>
    </div>
    <div class="flex-none">
        <a href="/" class="btn btn-ghost">Home</a>
        <form method="POST" action="/_auth/signout" class="ml-2">
            <button type="submit" class="btn btn-outline">Sign Out</button>
        </form>
    </div>
</div>

<!-- Add Todo Form -->
<div class="card bg-base-100 shadow-lg mb-8">
    <div class="card-body">
        <h2 class="card-title">Add New Todo</h2>
        <form hx-post="/todos" class="space-y-4">
            <div class="form-control">
                <input name="title" type="text" placeholder="What do you need to do?" 
                       class="input input-bordered w-full" required>
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
<div class="card bg-base-100 shadow-lg">
    <div class="card-body">
        <h2 class="card-title">
            Todo List
            <div class="badge badge-secondary">{{len (todos.AllTodos)}}</div>
        </h2>
        
        <div class="space-y-3">
            {{range todos.AllTodos}}
                {{template "partials/todos-item.html" .}}
            {{else}}
                <div class="text-center py-8 text-base-content/50">
                    <p>No todos yet! 🎉</p>
                    <p class="text-sm">Add one above to get started.</p>
                </div>
            {{end}}
        </div>
    </div>
</div>

{{template "layout/end"}}
```

**`views/partials/todos-item.html`**:
```html
<div class="flex items-center justify-between p-4 border border-base-300 rounded-lg {{if .Completed}}opacity-60{{end}}">
    <div class="flex items-center space-x-3">
        <button hx-put="/todos/{{.ID}}/toggle" 
                class="checkbox {{if .Completed}}checkbox-success{{end}}" 
                {{if .Completed}}checked{{end}}>
        </button>
        <span class="{{if .Completed}}line-through{{end}}">{{.Title}}</span>
    </div>
    <button hx-delete="/todos/{{.ID}}" 
            hx-confirm="Are you sure?"
            class="btn btn-ghost btn-sm text-error">
        🗑️
    </button>
</div>
```

**`views/partials/error-message.html`**:
```html
<div class="alert alert-error">
    <span>{{.}}</span>
</div>
```

## Step 5: Main Application

**`main.go`**:
```go
package main

import (
	"embed"
	"os"
	"cmp"
	"your-app/controllers"
	"your-app/models"
	"github.com/The-Skyscape/devtools/pkg/application"
)

//go:embed all:views
var views embed.FS

func main() {
	port := cmp.Or(os.Getenv("PORT"), "5000")
	
	application.Serve(views,
		application.WithController("auth", models.Auth.Controller()),
		application.WithController(controllers.Home()),
		application.WithController(controllers.Todos()),
		application.WithDaisyTheme("corporate"),
		application.WithPort(port),
	)
}
```

## Step 6: Environment Setup

Set environment variables:

```bash
# Required for authentication
export AUTH_SECRET="your-super-secret-jwt-key-change-this-in-production"

# Optional customization
export PORT="5000"
export THEME="corporate"
```

## Step 7: Running the Application

1. **Run the application**:
   ```bash
   go run .
   ```

2. **Visit** `http://localhost:5000`

3. **Sign up** for a new account

4. **Start adding todos!**

## Step 8: Deployment

Deploy your application to DigitalOcean:

```bash
# Download launch-app tool
curl -L -o launch-app https://github.com/The-Skyscape/devtools/releases/download/v1.0.1/launch-app
chmod +x launch-app

# Build your application
go build -o app

# Deploy using launch-app
export DIGITAL_OCEAN_API_KEY="your-token"
./launch-app --name my-todo-app --domain todos.example.com --binary ./app
```

The deployment tool will:
- Create a DigitalOcean droplet
- Install Docker and dependencies
- Containerize your application
- Generate SSL certificates
- Start your application

## Next Steps

### Features to Add

1. **Categories/Tags**: Add todo categorization
2. **Due Dates**: Add deadline tracking
3. **File Attachments**: Attach files to todos
4. **Search/Filter**: Add search functionality
5. **Bulk Operations**: Select multiple todos
6. **Sharing**: Share todos with other users

### Technical Improvements

1. **Add Tests**: Write unit and integration tests
2. **Add Validation**: Better form validation
3. **Add Caching**: Cache frequently accessed data
4. **Add Logging**: Better application logging
5. **Add Monitoring**: Health checks and metrics

## Conclusion

You've built a complete todo application using TheSkyscape DevTools! This tutorial covered:

- ✅ Project structure and MVC organization
- ✅ Database models with proper Table() methods
- ✅ HTTP controllers with factory functions
- ✅ Real-time UI updates with HTMX
- ✅ User authentication and authorization
- ✅ Responsive design with DaisyUI
- ✅ Cloud deployment with SSL certificates

The `create-app` CLI tool generates this exact application structure, so you can get started immediately and focus on your unique business logic.

Happy coding! 🚀