# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is the **TheSkyscape DevTools** repository - a Go-based toolkit for cloud infrastructure management and application deployment. The project provides unified abstractions for managing containers, cloud hosting, and web applications across multiple platforms.

## Design Philosophy

### HTMX/HATEOAS Architecture
We've rejected the complexity of modern JavaScript frameworks in favor of HTMX with HATEOAS principles:
- **HTML as the engine of application state** - The server sends HTML, not JSON
- **No client-side state management** - All state lives on the server
- **Progressive enhancement** - Works without JavaScript, enhanced with HTMX
- **Simplicity over features** - No webpack, no npm, no build pipeline for the frontend

### Value Receiver Pattern for Request Isolation
Our controllers use a unique pattern for request isolation without mutexes:
```go
// Value receiver creates a copy
func (c HomeController) Handle(r *http.Request) application.IController {
    c.Request = r  // Modifies the copy
    return &c      // Returns pointer to the copy
}
```
This gives each request its own controller instance (16-32 bytes overhead) with zero shared state.

### Template Validation with check-views
Templates are validated at build time using our `check-views` tool:
- Parses Go AST to find all controller methods
- Parses templates to find all references
- Validates that every template reference has a corresponding controller method
- Turns runtime template errors into build-time errors

### No Client State Principle
By eliminating client-side state, we've removed entire categories of bugs:
- No state synchronization issues
- No cache invalidation problems
- No version mismatches between API and client
- Debugging happens in one place: the server

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
- **`pkg/security/`** - HashiCorp Vault integration with automatic fallback storage for secrets management
- **`pkg/database/`** - Database abstraction layer supporting SQLite3 with dynamic queries
- **`pkg/email/`** - Email provider abstraction with Resend, SendGrid, and Postmark implementations

### Key Design Patterns

- **Interface-based abstractions**: `Host`, `Platform`, and `Service` interfaces allow swapping implementations
- **Option pattern**: Configuration through variadic option functions (e.g., `WithFileUpload()`, `WithSetupScript()`)
- **Embedded file systems**: Views and static assets embedded using Go's `embed` package
- **Plugin architecture**: Cloud platforms implemented as separate packages under `platforms/`
- **Fallback pattern**: SecretsController provides automatic fallback from Vault → File → Memory storage
- **Controller factory pattern**: Controllers return `(string, *Controller)` for registration with the app

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

### Email Package
The email package provides a unified interface for sending transactional emails:

```go
// Provider interface for email services
type Provider interface {
    Send(msg *Message) error
    GetName() string
}

// Message struct
type Message struct {
    To          string   // Recipient email
    From        string   // Sender email
    FromName    string   // Sender name
    Subject     string
    HTMLContent string   // HTML version
    TextContent string   // Plain text version
    ReplyTo     string   // Optional reply-to
    Tags        []string  // Optional tags for tracking
}

// Available providers
- ResendProvider - Modern email API with great DX
- SendGridProvider - Enterprise-grade delivery
- PostmarkProvider - Transactional email specialist
```

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

### Template Helpers & Type Safety
**IMPORTANT**: DevTools provides formatting helpers that frontend developers expect, but avoids anti-patterns:
- ✅ **DO USE**: Formatting functions (`formatBytes`, `formatDate`, `timeAgo`, etc.) for display
- ✅ **DO USE**: String manipulation (`truncate`, `pluralize`, `join`, `slice`, etc.) for text processing
- ✅ **DO USE**: Math functions - both integer (`add`, `sub`, `mul`, `div`) and float versions (`addf`, `subf`, `mulf`, `divf`)
- ❌ **NOT PROVIDED**: `dict`, `set`, `head` or similar functions that create untyped map[string]any
- **PRINCIPLE**: Go is a type-safe language - use typed structs from controllers, not generic maps
- **BEST PRACTICE**: All data should come from controller methods that return typed values
- **CONTEXT**: Use `request.Context()` directly in handlers, not wrapper methods

### HTMX/HATEOAS Error Handling
**IMPORTANT**: When using HTMX with HATEOAS, error responses must return HTML fragments with 200 OK status:
- DO NOT set HTTP status codes in error responses (blocks HTMX from swapping content)
- Errors should be rendered as HTML partials that can be inserted into the DOM
- Use `c.RenderError(w, r, err)` which returns 200 OK with error HTML
- Form errors should target specific error containers: `hx-target=".error-message"`
- This allows HTMX to display errors inline without full page refreshes

### Security Package (SecretsController)
The security package provides a controller-based approach to secrets management with automatic fallback:

```go
// Initialize with Vault service
vault := security.NewVaultService(
    security.WithContainerName("my-vault"),
    security.WithPort(8200),
)

// Add to application
application.Serve(views,
    application.WithController(security.NewController(vault)),
    // other controllers...
)
```

**Fallback Chain:**
1. **Vault Container** - Attempts to start HashiCorp Vault in Docker
2. **File Storage** - Falls back to encrypted file storage in `~/.skyscape/secrets/`
3. **Memory Storage** - Last resort, non-persistent in-memory storage

**Controller Methods for Templates:**
- `IsVaultAvailable()` - Check if real Vault is running
- `IsFallbackMode()` - Check if using fallback storage
- `GetStorageMode()` - Return current storage mode (vault/file/memory)
- `IsStripeConfigured()` - Check Stripe configuration
- `IsDigitalOceanConfigured()` - Check DigitalOcean configuration
- `GetVaultURL()` - Return Vault UI URL if available

**Storage Backends:**
- `MemoryBackend` - In-memory storage (non-persistent)
- `FileBackend` - Encrypted file storage with AES-GCM
- `EnvBackend` - Read-only environment variable storage
- `HybridBackend` - Combines multiple backends with fallback

## Integration Points

- **Docker Runtime** - All container operations require Docker daemon
- **SSH Keys** - Automatic SSH key generation for cloud server access (stored in `~/.ssh/`)
- **File System** - Template views embedded at build time, external views loaded at runtime
- **HTTP/HTTPS** - Dual-protocol web server with automatic SSL certificate detection
- **HTMX** - Built-in support for dynamic updates with `Controller.Refresh()`
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

## Concrete Examples for AI Assistants

### Complete Controller Example
```go
// controllers/resources.go
func Resources() (string, *ResourcesController) {
    return "resources", &ResourcesController{}
}

type ResourcesController struct {
    application.Controller  // ALWAYS embed Controller
}

func (c *ResourcesController) Setup(app *application.App) {
    c.Controller.Setup(app)  // ALWAYS call parent first
    
    // Routes
    http.Handle("GET /resources", app.Serve("resources.html", nil))
    http.Handle("POST /resources", app.ProtectFunc(c.create, nil))
}

// VALUE receiver for request isolation
func (c ResourcesController) Handle(req *http.Request) application.IController {
    c.Request = req
    return &c
}

// Template method - uppercase, uses c.PathValue()
func (c *ResourcesController) GetResource() (*models.Resource, error) {
    id := c.PathValue("id")  // Request from Handle()
    return models.Resources.Get(id)
}

// Handler method - lowercase, uses r.PathValue()
func (c *ResourcesController) create(w http.ResponseWriter, r *http.Request) {
    c.SetRequest(r)
    
    // Validation
    validator := c.Validator()
    validator.CheckRequired("name", r.FormValue("name"))
    
    if err := validator.Result(); err != nil {
        c.RenderError(w, r, err)  // ALWAYS render errors
        return
    }
    
    // Create
    resource := &models.Resource{
        Name: r.FormValue("name"),
    }
    
    if _, err := models.Resources.Insert(resource); err != nil {
        c.RenderError(w, r, err)
        return
    }
    
    c.Refresh(w, r)  // HTMX refresh
}
```

### Key Patterns to Remember

1. **IDs are ALWAYS strings** - Never convert to int
2. **SQL uses PascalCase** - `WHERE UserID = ?` not `WHERE user_id = ?`
3. **PathValue usage**:
   - In handlers: `r.PathValue("id")` (have request param)
   - In template methods: `c.PathValue("id")` (request from Handle)
4. **Error handling**: Always use `c.RenderError(w, r, err)`
5. **HTMX responses**: Use `c.Refresh()` or `c.Redirect()`

## Recent Improvements (Go Best Practices)

### Core Framework Improvements
1. **Error Handling**: Added sentinel errors (ErrNotFound, ErrUnauthorized, etc.) in application package
2. **Type Safety**: Removed anti-pattern functions (dict, set, head) that encourage untyped data
3. **Code Cleanup**: Removed unused featherweight platform
4. **Template Helpers**: Kept all formatting and math functions developers expect, removed only type-unsafe helpers
5. **Context Philosophy**: Handlers already have request.Context(), no need for wrapper methods

### Go Philosophy Applied
- **Simplicity**: Removed unnecessary abstraction layers and unused code
- **Clarity**: Explicit error types instead of string errors
- **Composition**: Improvements made within existing structures, not new packages
- **Type Safety**: Enforce typed data flow from controllers to templates
- **Idiomatic Go**: Following standard library patterns for context and errors

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