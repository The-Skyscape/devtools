# TheSkyscape DevTools

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=flat&logo=docker)](https://docker.com/)

**A unified Go toolkit for building cloud-native applications with built-in authentication, database management, container orchestration, and multi-cloud deployment.**

Perfect for building infrastructure management platforms, development environments, SaaS applications, and internal tools.

## 🚀 Quick Start

```bash
go mod init your-app
go get github.com/The-Skyscape/devtools
```

## 📦 Core Features

- **🌐 Web Framework** - MVC with embedded templates, HTMX, and DaisyUI
- **🔐 Authentication** - JWT sessions, bcrypt hashing, role-based access
- **🗄️ Database** - Dynamic ORM with SQLite3, migrations, type-safe repositories  
- **🐳 Containers** - Docker management for local and remote hosts
- **☁️ Cloud Deployment** - DigitalOcean, AWS, GCP with SSH key management
- **💻 Dev Workspaces** - Containerized code-server environments with Git

## 🏗️ Basic Application Structure

```
your-app/
├── controllers/     # HTTP handlers
├── models/          # Database models  
├── views/           # HTML templates
├── main.go          # Application entry
└── go.mod
```

## ⚡ Copy-Paste Patterns

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
    DB    = local.Database("app.db")
    Auth  = authentication.Manage(DB)
    Tasks = database.Manage(DB, new(Task))
)

// models/task.go
type Task struct {
    application.Model
    Title     string
    Completed bool
}

func (*Task) Table() string { return "tasks" }
```

### Controllers

```go
// controllers/tasks.go
package controllers

import (
    "net/http"
    "your-app/models"
    "github.com/The-Skyscape/devtools/pkg/application"
)

func Tasks() (string, *TaskController) {
    return "tasks", &TaskController{}
}

type TaskController struct {
    application.BaseController
}

func (c *TaskController) Setup(app *application.App) {
    c.BaseController.Setup(app)
    http.Handle("GET /", app.Serve("tasks.html", nil))
    http.Handle("POST /tasks", app.ProtectFunc(c.createTask, nil))
}

func (c TaskController) Handle(r *http.Request) application.Controller {
    c.Request = r
    return &c
}

// Template method: {{range tasks.AllTasks}}
func (c *TaskController) AllTasks() ([]*models.Task, error) {
    return models.Tasks.Search("")
}

func (c *TaskController) createTask(w http.ResponseWriter, r *http.Request) {
    task := &models.Task{Title: r.FormValue("title")}
    models.Tasks.Insert(task)
    c.Refresh(w, r) // HTMX refresh
}
```

### Main Application

```go
// main.go
package main

import (
    "embed"
    "your-app/controllers" 
    "your-app/models"
    "github.com/The-Skyscape/devtools/pkg/application"
)

//go:embed all:views
var views embed.FS

func main() {
    auth := models.Auth.Controller()
    
    application.Serve(views,
        application.WithController("auth", auth),
        application.WithController(controllers.Tasks()),
        application.WithDaisyTheme("corporate"),
    )
}
```

### Templates with HTMX

```html
<!-- views/tasks.html -->
<html data-theme="{{theme}}">
<head>
    <script src="https://unpkg.com/htmx.org"></script>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@3.9.4/dist/full.css" rel="stylesheet">
</head>
<body>
    <form hx-post="/tasks" class="mb-4">
        <input name="title" class="input input-bordered" placeholder="New task">
        <button class="btn btn-primary">Add</button>
    </form>
    
    <ul>
        {{range tasks.AllTasks}}
        <li class="flex items-center gap-2">
            <input type="checkbox" {{if .Completed}}checked{{end}}>
            <span>{{.Title}}</span>
        </li>
        {{end}}
    </ul>
</body>
</html>
```

## 🔧 Environment Setup

```bash
export AUTH_SECRET="your-jwt-secret"
export THEME="corporate"              # DaisyUI theme
export PORT="8080"                    # Default: 5000

# Optional: Cloud credentials
export DIGITAL_OCEAN_API_KEY="token"
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
client.Launch(server, hosting.WithFileUpload("./app", "/usr/local/bin/app"))
```

## 💻 Development Workspaces

```go
import "github.com/The-Skyscape/devtools/pkg/coding"

repo := coding.Manage(db)
workspace, _ := repo.NewWorkspace("user-123", 8080, gitRepo)
workspace.Start(user) // Launches code-server container
```

## 📚 Documentation

- **[Tutorial: Todo App](docs/tutorial.md)** - Complete walkthrough
- **[API Reference](docs/api.md)** - Package documentation
- **[Examples](docs/examples/)** - Real-world usage patterns
- **[Deployment Guide](docs/deployment.md)** - Production setup

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## 📝 License

MIT License - see [LICENSE](LICENSE) file.

---

**Built for TheSkyscape** - Simplifying cloud-native application development