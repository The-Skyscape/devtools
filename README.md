# TheSkyscape DevTools

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/The-Skyscape/devtools)](https://github.com/The-Skyscape/devtools/releases)

**A unified Go toolkit for building cloud-native applications with built-in authentication, database management, container orchestration, and multi-cloud deployment.**

Perfect for building infrastructure management platforms, development environments, SaaS applications, and internal tools.

## 🛠️ CLI Tools

TheSkyscape DevTools provides two main CLI tools for rapid development:

### create-app
Scaffold new TheSkyscape applications with proper MVC structure:

```bash
# Download latest release
curl -L -o create-app https://github.com/The-Skyscape/devtools/releases/download/v1.0.1/create-app
chmod +x create-app

# Create a new application
./create-app my-app
cd my-app
export AUTH_SECRET="your-secret"
go run .
```

**Generated structure:**
```
my-app/
├── controllers/
│   ├── home.go       # Home page controller
│   └── todos.go      # Todo management controller
├── models/
│   ├── database.go   # Database setup and globals
│   └── todo.go       # Todo model with Table() method
├── views/
│   ├── home.html     # Home page template
│   ├── todos.html    # Todo list template
│   ├── layout.html   # Base layout
│   └── partials/     # Reusable components
├── main.go           # Application entry point
└── go.mod
```

### launch-app
Deploy applications to DigitalOcean with automated setup:

```bash
# Download latest release
curl -L -o launch-app https://github.com/The-Skyscape/devtools/releases/download/v1.0.1/launch-app
chmod +x launch-app

# Build your app first
go build -o app

# Deploy to cloud
export DIGITAL_OCEAN_API_KEY="your-token"
./launch-app --name my-server --domain app.example.com --binary ./app
```

**Features:**
- Automated DigitalOcean droplet creation
- Docker containerization with proper networking
- SSL certificate generation via Let's Encrypt
- Server configuration persistence in `servers/` directory
- Custom binary deployment with embedded Dockerfile

## 🚀 Quick Start

### Option 1: Use CLI Tools (Recommended)
```bash
# Download CLI tools
curl -L -o create-app https://github.com/The-Skyscape/devtools/releases/download/v1.0.1/create-app
curl -L -o launch-app https://github.com/The-Skyscape/devtools/releases/download/v1.0.1/launch-app
chmod +x create-app launch-app

# Create and run new app
./create-app my-todo-app
cd my-todo-app
export AUTH_SECRET="your-secret"
go run .
```

### Option 2: As Go Library
```bash
go mod init your-app
go get github.com/The-Skyscape/devtools
```

## 📦 Core Framework Features

- **🌐 Web Framework** - MVC with embedded templates, HTMX, and DaisyUI
- **🔐 Authentication** - JWT sessions, bcrypt hashing, role-based access
- **🗄️ Database** - Dynamic ORM with SQLite3, migrations, type-safe repositories  
- **🐳 Containers** - Docker management for local and remote hosts
- **☁️ Cloud Deployment** - DigitalOcean, AWS, GCP with SSH key management
- **💻 Dev Workspaces** - Containerized code-server environments with Git

## 🏗️ Application Architecture

### Database & Models
```go
// models/database.go
package models

import (
    "github.com/The-Skyscape/devtools/pkg/authentication"
    "github.com/The-Skyscape/devtools/pkg/database"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

var (
    DB   = local.Database("app.db")
    Auth = authentication.Manage(DB)
    Todos = database.Manage(DB, new(Todo))
)

// models/todo.go
type Todo struct {
    application.Model
    Title     string
    Completed bool
}

func (*Todo) Table() string { return "todos" }
```

### Controllers
```go
// controllers/todos.go
package controllers

import (
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
    app.Serve("GET /", "todos.html", models.Auth.Required)
    app.ProtectFunc("POST /todos", c.createTodo, models.Auth.Required)
}

func (c *TodosController) Handle(r *http.Request) application.Controller {
    c.Request = r
    return c
}

// Template method: {{range todos.AllTodos}}
func (c *TodosController) AllTodos() ([]*models.Todo, error) {
    return models.Todos.Search("")
}

func (c *TodosController) createTodo(w http.ResponseWriter, r *http.Request) {
    todo := &models.Todo{Title: r.FormValue("title")}
    if err := models.Todos.Insert(todo); err != nil {
        c.Render(w, r, "error-message.html", err)
        return
    }
    c.Refresh(w, r)
}
```

### Main Application
```go
// main.go
package main

import (
    "embed"
    "os"
    "your-app/controllers" 
    "your-app/models"
    "github.com/The-Skyscape/devtools/pkg/application"
    "cmp"
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

### Templates with HTMX
```html
<!-- views/todos.html -->
{{define "layout/start"}}
<html data-theme="{{theme}}">
<head>
    <script src="https://unpkg.com/htmx.org"></script>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@3.9.4/dist/full.css" rel="stylesheet">
    <title>Todos</title>
</head>
<body class="p-4">
{{end}}

<h1 class="text-2xl font-bold mb-4">Todo List</h1>

<form hx-post="/todos" class="mb-4">
    <div class="flex gap-2">
        <input name="title" class="input input-bordered flex-1" placeholder="New todo" required>
        <button class="btn btn-primary">Add</button>
    </div>
</form>

<div class="space-y-2">
    {{range todos.AllTodos}}
        {{template "partials/todos-item.html" .}}
    {{end}}
</div>

{{define "layout/end"}}
</body>
</html>
{{end}}
```

## 🔧 Environment Variables

```bash
# Application
export PORT="5000"                    # Server port
export THEME="corporate"              # DaisyUI theme

# Security  
export AUTH_SECRET="your-jwt-secret"  # Required for authentication

# Cloud deployment
export DIGITAL_OCEAN_API_KEY="token" # For launch-app deployments
```

## 🐳 Container Management

```go
import "github.com/The-Skyscape/devtools/pkg/containers"

localhost := containers.Local()
service := &containers.Service{
    Name:  "redis",
    Image: "redis:alpine", 
    Ports: map[int]int{6379: 6379},
}
localhost.Launch(service)
```

## ☁️ Cloud Deployment

```go
import "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"

client := digitalocean.Connect(os.Getenv("DIGITAL_OCEAN_API_KEY"))
server := &digitalocean.Server{
    Name: "app-server", Size: "s-1vcpu-1gb", Region: "nyc1", Image: "docker-20-04",
}
client.Launch(server, hosting.WithFileUpload("./app", "/root/app"))
```

## 🔨 Development

```bash
# Build CLI tools
make build

# Install to system PATH
make install

# Clean build artifacts
make clean

# Run tests
go test ./...
```

## 📚 Documentation

- **[Tutorial: Todo App](docs/tutorial.md)** - Complete walkthrough
- **[API Reference](docs/api.md)** - Package documentation  
- **[Deployment Guide](docs/deployment.md)** - Production setup

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## 📝 License

MIT License - see [LICENSE](LICENSE) file.

---

**Built for TheSkyscape** - Simplifying cloud-native application development