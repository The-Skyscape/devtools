# INSTRUCT.md

**Project setup and development guide for Claude agents working with TheSkyscape DevTools applications**

## Project Overview

This application is built using **TheSkyscape DevTools**, a Go toolkit for cloud-native applications with built-in authentication, database management, container orchestration, and multi-cloud deployment capabilities.

## Architecture

This project follows the standard TheSkyscape DevTools application structure:

```
project-root/
├── controllers/                 # HTTP handlers and request logic
├── models/                      # Database models and business logic
├── views/                       # Page templates (embedded at build time)
├── assets/                      # Static files (optional)
├── deployments/                 # Infrastructure configs (optional)
├── main.go                      # Application entry point
├── go.mod                       # Go dependencies
├── go.sum
└── README.md                    # Project-specific documentation
```

## Development Patterns

### 1. Database Models (`models/`)

Models embed `application.Model` and implement the `Table()` method:

```go
package models

import (
    "github.com/The-Skyscape/devtools/pkg/application"
    "github.com/The-Skyscape/devtools/pkg/authentication"
    "github.com/The-Skyscape/devtools/pkg/database"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

// Database and repository setup
var (
    DB = local.Database("app.db")
    Auth = authentication.Manage(DB)
    Users = database.Manage(DB, new(User))
)

// Example model
type User struct {
    application.Model
    Name  string
    Email string
    Role  string
}

func (*User) Table() string { return "users" }

// Business logic methods
func GetUserByEmail(email string) (*User, error) {
    return Users.Search("email = ?", email)
}
```

### 2. Controllers (`controllers/`)

Controllers embed `application.BaseController` and expose methods to templates:

```go
package controllers

import (
    "net/http"
    "your-app/models"
    "github.com/The-Skyscape/devtools/pkg/application"
)

// Factory function returning controller name and instance
func Users() (string, *UserController) {
    return "users", &UserController{}
}

type UserController struct {
    application.BaseController
}

// Setup registers routes when application starts
func (c *UserController) Setup(app *application.App) {
    c.BaseController.Setup(app)
    http.Handle("GET /users", app.Serve("users.html", nil))
    http.Handle("POST /users", app.ProtectFunc(c.createUser, false))
}

// Handle returns controller instance for each request
func (c UserController) Handle(r *http.Request) application.Controller {
    c.Request = r
    return &c
}

// Template method: accessible as {{users.AllUsers}} in views
func (c *UserController) AllUsers() ([]*models.User, error) {
    return models.Users.Search("")
}

// HTTP handler methods
func (c *UserController) createUser(w http.ResponseWriter, r *http.Request) {
    user := &models.User{
        Name:  r.FormValue("name"),
        Email: r.FormValue("email"),
        Role:  r.FormValue("role"),
    }
    
    if _, err := models.Users.Insert(user); err != nil {
        c.Render(w, r, "error-message", err)
        return
    }
    
    c.Refresh(w, r) // HTMX page refresh
}
```

### 3. Main Application (`main.go`)

```go
package main

import (
    "embed"
    "os"
    "your-app/controllers"
    "your-app/models"
    "github.com/The-Skyscape/devtools/pkg/application"
    "github.com/The-Skyscape/devtools/pkg/authentication"
)

//go:embed all:views
var views embed.FS

func main() {
    // Setup authentication
    auth := models.Auth.Controller(
        authentication.WithCookie(os.Getenv("SESSION_COOKIE")),
        authentication.WithSignoutURL("/"),
    )
    
    // Start application with controllers
    application.Serve(views,
        application.WithController("auth", auth),
        application.WithController(controllers.Users()),
        application.WithDaisyTheme(os.Getenv("THEME")),
        application.WithHostPrefix(os.Getenv("HOST_PREFIX")),
    )
}
```

### 4. Templates (`views/*.html`)

Templates can access controller methods and built-in helpers:

```html
<html data-theme="{{theme}}">
<head>
    <title>My App</title>
    <script src="https://unpkg.com/htmx.org"></script>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@3.9.4/dist/full.css" rel="stylesheet">
</head>
<body>
    <!-- Authentication check -->
    {{if auth.CurrentUser}}
        <h1>Welcome, {{auth.CurrentUser.Name}}!</h1>
        
        <!-- Controller method access -->
        <div class="overflow-x-auto">
            <table class="table table-zebra">
                {{range users.AllUsers}}
                <tr>
                    <td>{{.Name}}</td>
                    <td>{{.Email}}</td>
                    <td><span class="badge">{{.Role}}</span></td>
                </tr>
                {{end}}
            </table>
        </div>
        
        <!-- HTMX form -->
        <form hx-post="{{host}}/users" class="card bg-base-100 shadow-xl">
            <div class="card-body">
                <input name="name" class="input input-bordered" placeholder="Name">
                <input name="email" class="input input-bordered" placeholder="Email">
                <select name="role" class="select select-bordered">
                    <option value="user">User</option>
                    <option value="admin">Admin</option>
                </select>
                <button class="btn btn-primary">Add User</button>
            </div>
        </form>
    {{else}}
        <!-- Show signin form -->
        {{template "signin.html"}}
    {{end}}
</body>
</html>
```

## Environment Configuration

Set these environment variables before running:

```bash
# Required
export AUTH_SECRET="your-jwt-secret-key"

# Optional application settings
export SESSION_COOKIE="myapp-session"     # Cookie name
export THEME="corporate"                  # DaisyUI theme
export HOST_PREFIX="/api/v1"              # URL prefix
export PORT="8080"                        # Server port (default: 5000)

# Optional: Custom data directory
export INTERNAL_DATA="/path/to/data"      # Default: ~/.theskyscape

# Optional: SSL certificates
export CONGO_SSL_FULLCHAIN="/path/to/fullchain.pem"
export CONGO_SSL_PRIVKEY="/path/to/privkey.pem"

# Cloud provider credentials (if using)
export DIGITAL_OCEAN_API_KEY="your-api-key"
```

## Common Development Tasks

### Running the Application

```bash
# Development
go run .

# Production build
go build -o app .
./app
```

### Database Operations

```bash
# Models are auto-created on first run
# Data stored in ~/.theskyscape/app.db by default

# Check tables
sqlite3 ~/.theskyscape/app.db ".tables"

# View data
sqlite3 ~/.theskyscape/app.db "SELECT * FROM users;"
```

### Adding New Features

1. **Create Model** (`models/feature.go`):
   ```go
   type Feature struct {
       application.Model
       Name string
   }
   func (*Feature) Table() string { return "features" }
   ```

2. **Add Repository** (`models/database.go`):
   ```go
   var Features = database.Manage(DB, new(Feature))
   ```

3. **Create Controller** (`controllers/features.go`):
   ```go
   func Features() (string, *FeatureController) {
       return "features", &FeatureController{}
   }
   ```

4. **Register Controller** (`main.go`):
   ```go
   application.WithController(controllers.Features()),
   ```

5. **Create Templates** (`views/features.html`)

## Container Management

For applications that need to manage Docker containers:

```go
import "github.com/The-Skyscape/devtools/pkg/containers"

// Local containers
localhost := containers.Local()
service := &containers.Service{
    Name:  "redis",
    Image: "redis:alpine",
    Ports: map[int]int{6379: 6379},
}
localhost.Launch(service)

// Remote containers
remote := containers.Remote(server)
remote.Launch(service)
```

## Cloud Deployment

For applications that deploy to cloud platforms:

```go
import "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"

client := digitalocean.Connect(os.Getenv("DIGITAL_OCEAN_API_KEY"))
server := &digitalocean.Server{
    Name:   "production-server",
    Size:   "s-1vcpu-1gb",
    Region: "nyc1",
    Image:  "docker-20-04",
}

deployedServer, err := client.Launch(server,
    hosting.WithFileUpload("./app", "/usr/local/bin/app"),
    hosting.WithSetupScript("systemctl", "enable", "--now", "myapp"),
)
```

## Development Workspace

For applications that provide development environments:

```go
import "github.com/The-Skyscape/devtools/pkg/coding"

repo := coding.Manage(models.DB)
gitRepo, err := repo.NewRepo("project-id", "My Project")
workspace, err := repo.NewWorkspace("user-id", 8080, gitRepo)
workspace.Start(user)
```

## Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./controllers

# Verbose testing
go test -v ./...

# Race condition testing
go test -race ./...
```

## Built-in Template Helpers

- `{{theme}}` - Current DaisyUI theme
- `{{host}}` - Host prefix for URLs
- `{{path "section" "id"}}` - Generate URL paths
- `{{path_eq "section" "id"}}` - Check current path
- `{{req.URL.Query.Get "param"}}` - Access request parameters
- `{{auth.CurrentUser}}` - Current authenticated user
- `{{controllerName.MethodName}}` - Call controller methods

## Security Features

The toolkit provides these security features automatically:

- **JWT Authentication** - Secure session management
- **Password Hashing** - Bcrypt with proper salt
- **CSRF Protection** - Built into forms
- **SSH Key Management** - Auto-generated for cloud access
- **Secure Cookies** - HTTPOnly with SameSite protection
- **File Permissions** - Proper chmod for sensitive files

## Performance Features

- **Embedded Assets** - Views compiled into binary
- **Connection Pooling** - Automatic database connection management
- **WAL Mode** - SQLite3 optimized for concurrency
- **Template Caching** - Pre-compiled templates
- **Container Reuse** - Intelligent lifecycle management

## Integration Libraries

The toolkit integrates seamlessly with:

- **HTMX** - Dynamic page updates without JavaScript
- **DaisyUI** - Beautiful UI components with Tailwind CSS
- **Docker** - Container management and deployment
- **Cloud APIs** - DigitalOcean, AWS, GCP integration
- **Git** - Repository and workspace management

## Troubleshooting

### Common Issues

1. **Port in use**: Change `PORT` environment variable
2. **Database locked**: Ensure no other instances are running
3. **Template not found**: Check file paths and embedding
4. **Authentication failing**: Verify `AUTH_SECRET` is set
5. **Container issues**: Ensure Docker daemon is running

### Debug Mode

```bash
export DEBUG=true
export LOG_LEVEL=debug
go run ./cmd/server
```

## Need Help?

- **DevTools Documentation**: See TheSkyscape DevTools repository
- **Example Application**: Check the `example/` directory in DevTools
- **Template Examples**: Look at existing `views/*.html` files
- **API Reference**: Check pkg documentation in DevTools repo