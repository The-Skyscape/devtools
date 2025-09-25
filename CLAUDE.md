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
func (c HomeController) Handle(r *http.Request) application.Handler {
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

# With DigitalOcean project assignment
./build/launch-app --name my-server --domain app.example.com --binary ./app \
  --project be9c9d70-33e3-4f12-9639-59973390243e
```

**Features:**
- Automated DigitalOcean droplet creation with Docker
- SSL certificate generation via Let's Encrypt
- Server configuration persistence as JSON files
- Container-based deployment with secure internal networking
- DigitalOcean project organization support (--project flag)
- Automatic resource assignment to specified projects

## Architecture

### Core Packages

- **`pkg/application/`** - Web application framework with MVC pattern, template rendering, and SSL support
- **`pkg/containers/`** - Docker container management with local and remote host abstractions
- **`pkg/hosting/`** - Multi-cloud server deployment using the **Platform-Resource Pattern** (see detailed docs below)
- **`pkg/authentication/`** - User authentication, sessions, and JWT token management
- **`pkg/security/`** - HashiCorp Vault integration with automatic fallback storage for secrets management
- **`pkg/database/`** - Database abstraction layer supporting SQLite3 with dynamic queries
- **`pkg/email/`** - Email provider abstraction with Resend, SendGrid, and Postmark implementations

### Key Design Patterns

#### Platform-Resource Pattern (CRITICAL for new packages)
This is the core pattern used in hosting and should be used for payments, storage, and other packages:

```go
// The client IS the platform
type Platform interface {
    NewResource(resource *Resource) (*Resource, error)
    GetResource(id string) (*Resource, error)
    DestroyResource(id string) error
}

type Resource struct {
    Platform Platform  // Bidirectional reference - THIS IS INTENTIONAL
    ID       string
    // fields...
}

// Resource delegates operations to Platform
func (r *Resource) Destroy() error {
    return r.Platform.DestroyResource(r.ID)
}
```

**Key Points:**
- Platform creates resources and assigns itself as `resource.Platform`
- Resources hold Platform reference to delegate operations
- This is NOT a circular dependency - it's intentional bidirectional collaboration
- Similar to Go's `http.Request`/`http.Client` pattern
- Use concrete types (`*Server`, `*Volume`), not unnecessary interfaces (`ServerRef`)

#### Other Patterns
- **ClientOption pattern**: Configuration through functional options (e.g., `WithProjectID()`, `WithTimeout()`)
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
make # Build both tools
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

### Email Package (pkg/emailing) - CRITICAL
The emailing package provides a **FUNCTION-BASED API**, not map-based:

**TEMPLATE SYNTAX - NEVER USE DOTS:**
- ✅ CORRECT: `{{Name}}`, `{{Product}}`, `{{Email}}`
- ❌ WRONG: `{{.Name}}`, `{{.Product}}`, `{{.Email}}`

```go
// Send an email with options
err := collection.Send(to, subject, opts...)

// Available options:
emailing.WithTemplate("filename.html")  // Use HTML template (lazy parsed)
emailing.WithHTML(htmlContent)          // Direct HTML content
emailing.WithText(textContent)          // Plain text version
emailing.WithRequest(r)                  // Inject HTTP request for templates
emailing.WithData(key, value)           // Register data as FUNCTION returning value (accepts any)
emailing.WithFunc(name, fn)             // Register custom function for lazy evaluation
emailing.WithType(emailType)            // Set email type for tracking
emailing.WithReplyTo(email)             // Set reply-to address
emailing.WithFromOverride(addr, name)   // Override default from address

// Example usage:
models.Emails.Send(
    "user@example.com",
    "Welcome!", 
    emailing.WithTemplate("welcome.html"),
    emailing.WithRequest(req),
    emailing.WithData("username", "Alice"),  // Creates: func() any { return "Alice" }
    emailing.WithFunc("timestamp", func() string { 
        return time.Now().Format(time.RFC3339) // Lazy evaluation
    }),
)
```

**Key Features:**
- **Function-based, not map-based** - No `map[string]any`, everything is type-safe functions
- **Lazy template parsing** - Templates parsed at send time with all functions registered
- **Consistent with application package** - Same patterns as controller/template system
- **WithData vs WithFunc** - Use WithData for static values, WithFunc for dynamic/computed values

**Template Access:**
- Templates can call any registered function: `{{username}}`, `{{timestamp}}`
- Built-in functions: `{{now}}`, `{{Year}}`, `{{baseURL req}}`
- Conditionals work with function results: `{{if username}}Hi {{username}}{{end}}`

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

## Directory Structure Philosophy

**IMPORTANT**: The devtools library provides the framework and utilities. Applications using devtools should follow these patterns:

### services/ Directory (Application-level, not in devtools)
**Purpose**: Docker container management services that applications need
**Pattern**: Wraps containers.Service from devtools to manage Docker containers
**Examples**: VS Code server, AI models, databases, etc.

### internal/ Directory (Application-level, not in devtools)
**Purpose**: Business logic and algorithms specific to the application
**Pattern**: Implements domain logic that's beyond devtools' scope
**Examples**: Financial calculations, AI orchestration, workspace provisioning

The devtools library provides the building blocks (containers, database, auth, etc.) while applications compose these into services and business logic.

## Integration Points

- **Docker Runtime** - All container operations require Docker daemon
- **SSH Keys** - Automatic SSH key generation for cloud server access (stored in `~/.ssh/`)
- **File System** - Template views embedded at build time, external views loaded at runtime
- **HTTP/HTTPS** - Dual-protocol web server with automatic SSL certificate detection
- **HTMX** - Built-in support for dynamic updates with `Controller.Refresh()`
- **DaisyUI** - Theme integration through `WithDaisyTheme()` option

## Security Architecture

### Network Isolation Pattern
All internal services (Vault, databases, etc.) run on an isolated Docker network:
- **Internal Network**: `skyscape-internal` - containers communicate using container names
- **No Host Networking**: Never use `--network host` which exposes all ports
- **Explicit Port Binding**: Only expose necessary ports (80, 443) to the host
- **Vault Security**: HashiCorp Vault runs without external port exposure

### Vault Configuration
```go
// Secure vault configuration pattern - CRITICAL FOR SECURITY
Secrets = security.Manage(
    security.WithVault(
        security.WithContainerName("skyscape-vault"),
        security.WithNetwork("skyscape-internal"),
        security.WithPortBinding(false),  // NEVER set to true - prevents port 8200 exposure
    ),
)

// The vault service configuration ensures isolation
type VaultConfig struct {
    ExposePort bool  // MUST be false for production
    Port       int   // Internal port (8200), not exposed externally
}
```

### Container Security Best Practices
- All containers use the internal Docker network for communication
- Vault accessible only from within the Docker network (port 8200 not exposed)
- Applications access vault via container name: `http://skyscape-vault:8200`
- External access limited to web ports only (80/443)
- Port mapping uses map[int]int for proper type safety

### Security Vulnerability Prevention
- **Never expose Vault port 8200** to the public internet
- **Always use internal Docker networks** for service communication
- **Validate port configurations** before deployment
- **Use WithPortBinding(false)** for all internal services

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
func (c ResourcesController) Handle(req *http.Request) application.Handler {
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

## Platform-Resource Pattern Implementation Guide

When implementing new packages, follow the pattern established in the hosting package:

### 1. Study the Hosting Package
The hosting package (`pkg/hosting/`) is the reference implementation. Review:
- How `Platform` interface is defined
- How `Server`, `Volume`, `Domain` structs include Platform reference
- How platform implementations assign themselves during resource creation
- How resources delegate operations back to their Platform

### 2. Key Implementation Rules
- **Platform creates resources**: Use `NewResource()` methods that return structs
- **Platform assigns itself**: `resource.Platform = p` in creation methods
- **Resources delegate back**: Convenience methods call Platform methods
- **Use concrete types**: Return `*Resource`, not `ResourceRef` interfaces
- **ClientOption pattern**: For flexible platform configuration

### 3. Example Structure
```go
// Define your Platform interface
type Platform interface {
    NewResource(r *Resource) (*Resource, error)
    GetResource(id string) (*Resource, error)
    DestroyResource(id string) error
}

// Resources hold Platform reference
type Resource struct {
    Platform Platform  // Set by Platform.NewResource
    ID       string
    // ... fields
}

// Convenience methods delegate to Platform
func (r *Resource) Destroy() error {
    return r.Platform.DestroyResource(r.ID)
}
```

### 4. Testing with Mock
Create a mock implementation that follows the same pattern as `pkg/hosting/platforms/mock/`
for comprehensive testing without real infrastructure.

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