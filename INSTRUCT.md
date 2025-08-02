# INSTRUCT.md

**Instructions for Claude agents building web applications with TheSkyscape DevTools**

## Overview

TheSkyscape DevTools is a Go-based cloud infrastructure toolkit that provides unified abstractions for building, deploying, and managing web applications across multiple cloud platforms. Use this toolkit when you need to create applications that can manage containers, handle user authentication, work with Git repositories, or deploy to cloud platforms.

## Quick Start

### 1. Basic Web Application Setup

```go
package main

import (
    "github.com/The-Skyscape/devtools/pkg/application"
    "github.com/The-Skyscape/devtools/pkg/authentication"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

func main() {
    // Initialize database
    db := local.Database("myapp.db")
    
    // Setup authentication
    auth := authentication.Manage(db).Controller()
    
    // Create application with authentication
    app := application.New(nil,
        application.WithController("auth", auth),
        application.WithDaisyTheme("corporate"),
    )
    
    // Start the server
    app.Start()
}
```

### 2. Required Environment Variables

Set these environment variables before running your application:

```bash
# Required for authentication
export AUTH_SECRET="your-secret-key-here"

# Required for cloud deployment (if using DigitalOcean)
export DIGITAL_OCEAN_API_KEY="your-do-api-key"

# Optional: Custom SSL certificates
export CONGO_SSL_FULLCHAIN="/path/to/fullchain.pem"
export CONGO_SSL_PRIVKEY="/path/to/privkey.pem"

# Optional: Custom data directory
export INTERNAL_DATA="/custom/data/path"
```

## Common Use Cases

### 1. Container Management Application

Build applications that manage Docker containers across local and remote hosts:

```go
import "github.com/The-Skyscape/devtools/pkg/containers"

// Local container management
localhost := containers.Local()
service := &containers.Service{
    Name:  "web-app",
    Image: "nginx:latest",
    Ports: map[int]int{8080: 80},
    Env:   map[string]string{"NODE_ENV": "production"},
}
localhost.Launch(service)

// Remote container management
remote := containers.Remote(server)
remote.Launch(service)
```

### 2. Cloud Server Management

Deploy and manage servers across cloud platforms:

```go
import "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"

client := digitalocean.Connect("your-api-key")
server := &digitalocean.Server{
    Name:   "my-server",
    Size:   "s-1vcpu-1gb", 
    Region: "sfo2",
    Image:  "docker-20-04",
}

deployedServer, err := client.Launch(server,
    hosting.WithFileUpload("./app", "/usr/local/bin/app"),
    hosting.WithSetupScript("systemctl", "start", "docker"),
)
```

### 3. Git Repository Management

Create applications that manage Git repositories and workspaces:

```go
import "github.com/The-Skyscape/devtools/pkg/coding"

repo := coding.Manage(db)
gitRepo, err := repo.NewRepo("project-id", "My Project")

// Create development workspace
workspace, err := repo.NewWorkspace("developer1", 8080, gitRepo)
workspace.Start(user)
```

### 4. User Authentication & Authorization

Add user management to your applications:

```go
auth := authentication.Manage(db).Controller(
    authentication.WithCookie("myapp-session"),
    authentication.WithSignoutURL("/dashboard"),
)

// Protect routes
http.Handle("/admin/", auth.Protect(adminHandler, true))  // Admin only
http.Handle("/user/", auth.Protect(userHandler, false))   // Any authenticated user

// Use in templates
app := application.New(views, application.WithController("auth", auth))
```

## Application Patterns

### 1. Multi-Tenant SaaS Platform

```go
// Each tenant gets their own database and containers
func createTenant(tenantID string) {
    db := local.Database(fmt.Sprintf("tenant_%s.db", tenantID))
    auth := authentication.Manage(db).Controller()
    
    app := application.New(tenantViews,
        application.WithController("auth", auth),
        application.WithHostPrefix(fmt.Sprintf("/tenant/%s", tenantID)),
    )
}
```

### 2. Development Platform (Like Replit/CodeSandbox)

```go
func createDevEnvironment(userID, repoID string) {
    coding := coding.Manage(db)
    workspace, _ := coding.NewWorkspace(userID, 8000+userCount, repo)
    
    // Workspace automatically starts code-server container
    workspace.Start(user)
    
    // Proxy traffic to the workspace
    http.Handle("/workspace/", workspace.Service().Proxy(8080))
}
```

### 3. Cloud Management Dashboard

```go
func cloudDashboard() {
    // Manage multiple cloud providers
    doClient := digitalocean.Connect(os.Getenv("DO_KEY"))
    
    // List and manage servers
    servers, _ := doClient.GetServer("server-id")
    
    // Deploy applications
    localhost := containers.Local()
    localhost.Launch(&containers.Service{
        Name: "monitoring",
        Image: "grafana/grafana",
        Ports: map[int]int{3000: 3000},
    })
}
```

## Template Integration

### Using Authentication in Views

```html
<!-- Check if user is authenticated -->
{{if auth.CurrentUser}}
    <p>Welcome, {{auth.CurrentUser.Name}}!</p>
    {{template "logout-button.html"}}
{{else}}
    {{template "signin.html"}}
{{end}}

<!-- Admin-only sections -->
{{if auth.CurrentUser.IsAdmin}}
    <a href="/admin">Admin Panel</a>
{{end}}
```

### Path and Theme Helpers

```html
<!-- Use theme helper -->
<html data-theme="{{theme}}">

<!-- Generate paths -->
<a href="{{path "projects" .ID "settings"}}">Settings</a>

<!-- Check current path -->
{{if path_eq "projects" .ID}}
    <span class="active">Current Project</span>
{{end}}
```

## File Structure Recommendations

```
your-app/
├── cmd/
│   ├── server/main.go          # Main application entry
│   └── migrate/main.go         # Database migrations
├── internal/
│   ├── handlers/               # HTTP handlers
│   ├── models/                 # Business logic models
│   └── middleware/             # Custom middleware
├── views/
│   ├── layouts/                # Base templates
│   ├── pages/                  # Page templates
│   └── components/             # Reusable components
├── assets/
│   ├── css/                    # Stylesheets
│   ├── js/                     # JavaScript
│   └── images/                 # Static images
└── deployments/
    ├── docker/                 # Container configurations
    └── terraform/              # Infrastructure as code
```

## Security Best Practices

1. **Always set AUTH_SECRET** - Use a strong, random secret for JWT signing
2. **Use HTTPS in production** - Set SSL certificate paths via environment variables
3. **Validate user input** - Never trust data from forms or URLs
4. **Use admin-only protection** - Wrap sensitive handlers with `auth.Protect(handler, true)`
5. **Secure file permissions** - The toolkit automatically sets secure permissions for SSH keys
6. **Environment isolation** - Each application gets its own data directory

## Available Platforms

### Container Hosts
- **LocalHost** - Run containers on local machine
- **RemoteHost** - Run containers on remote servers via SSH

### Cloud Platforms  
- **DigitalOcean** - Full implementation with DNS management
- **Featherweight** - Placeholder (coming soon)

## Performance Tips

1. **Use embedded assets** - Embed views and static files for faster startup
2. **Enable WAL mode** - SQLite3 uses WAL mode by default for better concurrency
3. **Container reuse** - Check if containers are running before creating new ones
4. **Connection pooling** - Database connections are automatically pooled
5. **Template caching** - Templates are parsed once at startup

## Troubleshooting

### Common Issues

1. **Port conflicts** - Ensure ports aren't already in use before launching containers
2. **SSH key permissions** - Keys are auto-generated with correct permissions
3. **Database locks** - Use WAL mode (enabled by default) for concurrent access
4. **Memory usage** - Monitor container resource usage in production

### Debug Mode

Set environment variables for additional logging:
```bash
export DEBUG=true
export LOG_LEVEL=debug
```

## Integration Examples

The toolkit integrates seamlessly with:
- **HTMX** - Built-in support for dynamic updates
- **DaisyUI** - Theme integration with `WithDaisyTheme()`
- **TailwindCSS** - Works with embedded CSS assets
- **Docker** - Native container management
- **Git** - Full repository management
- **Cloud APIs** - Direct integration with provider SDKs