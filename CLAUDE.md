# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is the **TheSkyscape DevTools** repository - a Go-based toolkit for cloud infrastructure management and application deployment. The project provides unified abstractions for managing containers, cloud hosting, and web applications across multiple platforms.

## CLI Tools

### create-app
Application scaffolding tool that generates complete TheSkyscape applications:

```bash
./build/create-app my-app
cd my-app
go run .
```

**Generated structure:**
- `controllers/` - HTTP handlers with factory functions 
- `models/` - Database models with Table() methods
- `views/` - HTML templates with HTMX integration
- `main.go` - Application entry point with embedded views

### launch-app  
Cloud deployment tool for DigitalOcean with automated setup:

```bash
export DIGITAL_OCEAN_API_KEY="your-token"
./build/launch-app --name my-server --domain app.example.com --binary ./app
```

**Features:**
- Automated DigitalOcean droplet creation with Docker
- SSL certificate generation via Let's Encrypt
- Server configuration persistence as JSON files
- Container-based deployment with proper networking

## Architecture

### Core Packages

- **`pkg/application/`** - Web application framework with MVC pattern, template rendering, and SSL support
- **`pkg/containers/`** - Docker container management with local and remote host abstractions
- **`pkg/hosting/`** - Multi-cloud server deployment (DigitalOcean, AWS, GCP) with unified Platform interface
- **`pkg/authentication/`** - User authentication, sessions, and JWT token management
- **`pkg/database/`** - Database abstraction layer supporting SQLite3 with dynamic queries

### Key Design Patterns

- **Interface-based abstractions**: `Host`, `Platform`, and `Service` interfaces allow swapping implementations
- **Option pattern**: Configuration through variadic option functions (e.g., `WithFileUpload()`, `WithSetupScript()`)
- **Embedded file systems**: Views and static assets embedded using Go's `embed` package
- **Plugin architecture**: Cloud platforms implemented as separate packages under `platforms/`

### Commands

- **`cmd/create-app/`** - Application scaffolding tool with todo template generation
- **`cmd/launch-app/`** - Cloud deployment tool with embedded resources for server setup

### Example Applications

- **`example/`** - Simple todo application showing basic usage patterns
  - Uses embedded views with `//go:embed all:views`
  - Demonstrates controller methods accessible in templates
  - Shows HTMX integration with `c.Refresh(w, r)` 
  - Proper MVC separation with models, controllers, and views

- **`workspace/`** - Production-ready GitHub-like developer platform
  - Complete repository management with file browsing and search
  - Containerized VS Code workspaces with Docker integration
  - Issues and pull request management with full CRUD operations
  - Mobile-responsive design with DaisyUI components
  - Advanced permission system with role-based access control
  - **Note**: Uses internal `coding` package (moved from `pkg/coding/`)

## Common Development Commands

### Building CLI Tools
```bash
make build          # Build both tools
make clean          # Clean build artifacts
make install        # Install to system PATH
```

### Building Individual Tools
```bash
go build ./cmd/create-app
go build ./cmd/launch-app
```

### Testing
```bash
go test ./...
go test -v ./pkg/containers
go test -race ./pkg/hosting
```

### Dependencies
```bash
go mod tidy
go mod download
```

### Running Applications
```bash
# Use CLI tools (recommended)
./build/create-app my-app
cd my-app && go run .

# Run example application
go run ./example
```

## Key Dependencies

- **Cloud SDKs**: DigitalOcean (`godo`), AWS support planned
- **Database**: SQLite3 driver (`mattn/go-sqlite3`) 
- **Security**: JWT tokens (`golang-jwt/jwt`), bcrypt (`golang.org/x/crypto`)
- **Migration**: Database migrations (`golang-migrate/migrate`)

## Environment Variables

### Application Settings
- `PORT` - Application server port (default: 5000)
- `AUTH_SECRET` - JWT signing secret (required for authentication)
- `THEME` - DaisyUI theme selection (default: corporate)

### SSL Configuration  
- `CONGO_SSL_FULLCHAIN` - SSL certificate path (default: /root/fullchain.pem)
- `CONGO_SSL_PRIVKEY` - SSL private key path (default: /root/privkey.pem)

### Cloud Platforms
- `DIGITAL_OCEAN_API_KEY` - DigitalOcean API token (required for launch-app)
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` - AWS credentials
- `GCP_PROJECT_ID`, `GCP_SERVICE_ACCOUNT_KEY`, `GCP_ZONE` - GCP credentials

## Application Development Patterns

### Controller Pattern (from create-app templates)
Controllers should implement factory functions and the required methods:
- `func Name() (string, *Controller)` - Factory function returning name and instance
- `Setup(app *application.App)` - Called at application startup to register routes
- `Handle(r *http.Request) application.Controller` - Called per request, returns controller instance
- Public methods accessible in templates (e.g., `AllTodos()` → `{{range todos.AllTodos}}`)

### Model Pattern
Models should embed `application.Model` and implement:
- `Table() string` method returning the database table name
- Use `database.Manage(db, new(Model))` to get a typed repository

### Application Startup Pattern
Use `application.Serve(views, ...options)` for convenience:

```go
//go:embed all:views
var views embed.FS

func main() {
    application.Serve(views,
        application.WithController("auth", models.Auth.Controller()),
        application.WithController(controllers.Home()),
        application.WithController(controllers.Todos()),
        application.WithDaisyTheme("corporate"),
    )
}
```

### Template Integration
- Controllers registered with `WithController("name", controller)` are accessible as `{{name.Method}}`
- Built-in helpers: `{{theme}}`, `{{host}}`, `{{path}}`, `{{req}}`
- HTMX integration: use `c.Refresh(w, r)` to trigger page refresh after form submission
- Templates use unique filenames (no paths) due to Go's global template namespace

## Integration Points

- **Docker Runtime** - All container operations require Docker daemon
- **SSH Keys** - Automatic SSH key generation for cloud server access (stored in `~/.ssh/`)
- **File System** - Template views embedded at build time, external views loaded at runtime
- **HTTP/HTTPS** - Dual-protocol web server with automatic SSL certificate detection
- **HTMX** - Built-in support for dynamic updates with `BaseController.Refresh()`
- **DaisyUI** - Theme integration through `WithDaisyTheme()` option

## Project Structure (Generated by create-app)

```
my-app/
├── controllers/
│   ├── home.go       # Home page controller
│   └── todos.go      # Todo CRUD controller  
├── models/
│   ├── database.go   # Database setup with global variables
│   └── todo.go       # Todo model with Table() method
├── views/
│   ├── home.html     # Home page template
│   ├── todos.html    # Todo list template
│   ├── layout.html   # Shared layout components
│   └── partials/     # Reusable template components
├── main.go           # Application entry point
└── go.mod
```

## CLI Usage Examples

```bash
# Create new application
./build/create-app my-todo-app
cd my-todo-app
export AUTH_SECRET="your-secret-key"
go run .

# Deploy to cloud
go build -o app
export DIGITAL_OCEAN_API_KEY="your-token"
../devtools/build/launch-app --name production --domain app.example.com --binary ./app
```