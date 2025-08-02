# TheSkyscape DevTools

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=flat&logo=docker)](https://docker.com/)

**A unified Go toolkit for building cloud-native applications with built-in container management, multi-cloud deployment, and development workspace orchestration.**

TheSkyscape DevTools provides everything you need to build modern web applications that can manage infrastructure, deploy to multiple cloud providers, and provide development environments as a service.

## 🚀 Features

### 🌐 **Web Application Framework**
- Clean MVC architecture with embedded template engine
- HTMX integration for dynamic, real-time updates
- Built-in SSL/TLS support with automatic certificate detection
- DaisyUI theme integration for beautiful interfaces

### 🔐 **Authentication & Authorization**
- JWT-based session management with secure defaults
- Bcrypt password hashing with admin role support
- Flexible middleware for protecting routes and resources
- Cookie-based authentication with CSRF protection

### 🐳 **Container Management**
- Unified Docker abstraction for local and remote hosts
- Service lifecycle management (create, start, stop, remove)
- Port mapping, volume mounting, and environment configuration
- Built-in reverse proxy for container services

### ☁️ **Multi-Cloud Deployment**
- **DigitalOcean**: Full implementation with droplet management and DNS
- **AWS**: Coming soon with EC2 and security group support
- **GCP**: Planned with Compute Engine integration
- Automatic SSH key generation and secure server access

### 💻 **Development Workspaces**
- Containerized code environments using code-server
- Git repository integration with access token management
- Per-user workspace isolation with port management
- Real-time collaboration and workspace sharing

### 🗄️ **Database & ORM**
- Dynamic ORM built on SQLite3 with WAL mode
- Reflection-based table creation and field mapping
- Type-safe repository pattern with Go generics
- Built-in migration support and database versioning

## 📦 Installation

```bash
go get github.com/The-Skyscape/devtools
```

### Prerequisites
- **Go 1.24+** - Latest Go version with generics support
- **Docker** - Required for container management features
- **Git** - For repository management functionality

## 🏃‍♂️ Quick Start

### 1. Install the Toolkit

```bash
go mod init your-app
go get github.com/The-Skyscape/devtools
```

### 2. Create Project Structure

```
your-app/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── controllers/             # HTTP request handlers
│   └── models/                  # Database models and business logic
├── views/
│   ├── dashboard.html           # Page templates
│   └── includes.html            # Shared components
├── go.mod
└── go.sum
```

### 3. Set Up Your Models (`internal/models/database.go`)

```go
package models

import (
    "github.com/The-Skyscape/devtools/pkg/authentication"
    "github.com/The-Skyscape/devtools/pkg/database"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

var (
    // DB is your application's database
    DB = local.Database("myapp.db")
    
    // Auth manages user authentication
    Auth = authentication.Manage(DB)
    
    // Products manages your business models
    Products = database.Manage(DB, new(Product))
)
```

### 4. Define Your Models (`internal/models/product.go`)

```go
package models

import "github.com/The-Skyscape/devtools/pkg/application"

type Product struct {
    application.Model
    Name        string
    Description string
    Price       float64
}

func (*Product) Table() string { return "products" }

// Business logic methods
func GetProductByID(id string) (*Product, error) {
    return Products.Get(id)
}

func SearchProducts(query string) ([]*Product, error) {
    return Products.Search("name LIKE ?", "%"+query+"%")
}
```

### 5. Create Controllers (`internal/controllers/products.go`)

```go
package controllers

import (
    "net/http"
    "your-app/internal/models"
    "github.com/The-Skyscape/devtools/pkg/application"
)

func Products() (string, *ProductController) {
    return "products", &ProductController{}
}

type ProductController struct {
    application.BaseController
}

func (c *ProductController) Setup(app *application.App) {
    c.BaseController.Setup(app)
    http.Handle("GET /products", app.Serve("products.html", nil))
    http.Handle("POST /products", app.ProtectFunc(c.createProduct, nil))
}

func (c ProductController) Handle(r *http.Request) application.Controller {
    c.Request = r
    return &c
}

// Template method: {{range products.AllProducts}}
func (c *ProductController) AllProducts() ([]*models.Product, error) {
    return models.Products.Search("")
}

func (c *ProductController) createProduct(w http.ResponseWriter, r *http.Request) {
    product := &models.Product{
        Name:        r.FormValue("name"),
        Description: r.FormValue("description"),
    }
    if _, err := models.Products.Insert(product); err != nil {
        c.Render(w, r, "error-message", err)
        return
    }
    c.Refresh(w, r) // HTMX refresh
}
```

### 6. Main Application (`cmd/server/main.go`)

```go
package main

import (
    "embed"
    "os"
    "your-app/internal/controllers"
    "your-app/internal/models"
    "github.com/The-Skyscape/devtools/pkg/application"
    "github.com/The-Skyscape/devtools/pkg/authentication"
)

//go:embed all:views
var views embed.FS

func main() {
    // Create authentication controller
    auth := models.Auth.Controller(
        authentication.WithCookie("myapp-session"),
        authentication.WithSignoutURL("/dashboard"),
    )
    
    // Start application
    application.Serve(views,
        application.WithController("auth", auth),
        application.WithController(controllers.Products()),
        application.WithDaisyTheme(os.Getenv("THEME")),
        application.WithHostPrefix(os.Getenv("PREFIX")),
    )
}
```

### 7. Run Your Application

```bash
# Set required environment variables
export AUTH_SECRET="your-secret-key"
export THEME="corporate"

# Run the application
go run ./cmd/server

# Visit http://localhost:5000
```

### Container Management

```go
import "github.com/The-Skyscape/devtools/pkg/containers"

// Launch local container
localhost := containers.Local()
service := &containers.Service{
    Name:  "web-app",
    Image: "nginx:latest",
    Ports: map[int]int{8080: 80},
    Env:   map[string]string{"NODE_ENV": "production"},
}
localhost.Launch(service)

// Deploy to remote server
remote := containers.Remote(server)
remote.Launch(service)
```

### Cloud Server Deployment

```go
import "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"

client := digitalocean.Connect("your-api-key")
server := &digitalocean.Server{
    Name:   "production-server",
    Size:   "s-2vcpu-4gb",
    Region: "nyc1",
    Image:  "docker-20-04",
}

deployedServer, err := client.Launch(server,
    hosting.WithFileUpload("./app", "/usr/local/bin/app"),
    hosting.WithSetupScript("systemctl", "enable", "--now", "docker"),
)
```

## 🔧 Configuration

### Environment Variables

```bash
# Required for authentication
export AUTH_SECRET="your-256-bit-secret-key"

# Cloud provider credentials
export DIGITAL_OCEAN_API_KEY="your-digitalocean-token"

# Optional: Custom SSL certificates
export CONGO_SSL_FULLCHAIN="/path/to/fullchain.pem"
export CONGO_SSL_PRIVKEY="/path/to/privkey.pem"

# Optional: Custom data directory (default: ~/.theskyscape)
export INTERNAL_DATA="/custom/data/directory"

# Optional: Custom application port (default: 5000)
export PORT="8080"

# Optional: Application customization
export PREFIX="/api/v1"        # Host prefix for URLs
export THEME="corporate"       # DaisyUI theme
export TOKEN="my-app-cookie"   # Custom cookie name
```

### Recommended Project Structure

```
your-application/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── controllers/             # HTTP handlers and business logic
│   ├── models/                  # Database models and repositories
│   └── middleware/              # Custom middleware (optional)
├── views/
│   ├── *.html                   # Page templates
│   └── components/              # Reusable UI components (optional)
├── assets/                      # Static files (optional)
│   ├── css/
│   ├── js/
│   └── images/
├── deployments/                 # Infrastructure (optional)
│   ├── Dockerfile
│   └── docker-compose.yml
├── scripts/                     # Build and deployment scripts (optional)
├── docs/                        # Project documentation (optional)
├── go.mod
├── go.sum
├── README.md
└── INSTRUCT.md                  # Copy from TheSkyscape DevTools
```

See the [`example/`](./example/) directory in this repository for a complete working application.

## 💡 What You Can Build

### 🏗️ **Infrastructure Management Platforms**
Applications like **Terraform Cloud** or **AWS Console** that manage cloud resources across multiple providers.

### 👨‍💻 **Development Environment Platforms**  
Services like **Replit**, **CodeSandbox**, or **Gitpod** with containerized development environments.

### 🏢 **Multi-Tenant SaaS Applications**
Platforms where each tenant gets isolated containers, databases, and cloud resources.

### 📊 **DevOps & Monitoring Dashboards**
Comprehensive dashboards that deploy, monitor, and manage containerized applications.

### 🔧 **Internal Developer Tools**
Custom tooling for your team's deployment pipelines and development workflows.

### 🌐 **Cloud-Native Web Applications**
Any web application that needs authentication, database management, and cloud deployment capabilities.

## 🎯 Implementation Examples

### Multi-Tenant Development Platform

```go
type TenantController struct {
    application.BaseController
    workspaces *database.Repository[*coding.Workspace]
}

func (c *TenantController) Setup(app *application.App) {
    c.BaseController.Setup(app)
    http.Handle("GET /workspace", app.Serve("workspace.html", c.authCheck))
    http.Handle("POST /workspace/create", app.ProtectFunc(c.createWorkspace, false))
}

func (c TenantController) Handle(r *http.Request) application.Controller {
    c.Request = r
    return &c
}

func (c *TenantController) AllWorkspaces() ([]*coding.Workspace, error) {
    return c.workspaces.Search("")
}

func createTenant(tenantID string) {
    // Isolated database per tenant
    db := local.Database(fmt.Sprintf("tenant_%s.db", tenantID))
    auth := authentication.Manage(db)
    coding := coding.Manage(db)
    
    authController := auth.Controller(
        authentication.WithCookie(fmt.Sprintf("tenant-%s", tenantID)),
    )
    
    application.Serve(tenantViews,
        application.WithController("auth", authController),
        application.WithController("workspaces", &TenantController{workspaces: coding.Workspaces}),
        application.WithHostPrefix(fmt.Sprintf("/tenant/%s", tenantID)),
    )
}
```

### Cloud Resource Manager

```go
func deployApplication(appConfig *AppConfig) error {
    // Deploy to multiple clouds simultaneously
    doClient := digitalocean.Connect(os.Getenv("DO_KEY"))
    
    // Create servers
    for _, region := range appConfig.Regions {
        server := &digitalocean.Server{
            Name:   fmt.Sprintf("%s-%s", appConfig.Name, region),
            Region: region,
            Size:   appConfig.ServerSize,
            Image:  "docker-20-04",
        }
        
        deployed, err := doClient.Launch(server,
            hosting.WithFileUpload(appConfig.Binary, "/app/server"),
            hosting.WithSetupScript("systemctl", "start", "docker"),
        )
        
        // Setup DNS
        deployed.Alias(appConfig.Subdomain, appConfig.Domain)
        
        // Deploy containers
        remote := containers.Remote(deployed)
        remote.Launch(&containers.Service{
            Name:  appConfig.Name,
            Image: appConfig.DockerImage,
            Ports: map[int]int{80: appConfig.Port},
        })
    }
    
    return nil
}
```

## 🏗️ Architecture

TheSkyscape DevTools follows a **clean architecture** pattern with clear separation of concerns:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │    │  Authentication │    │     Coding      │
│   (Web Framework)│    │   (Auth & Users)│    │ (Git & Workspace)│
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴───────────┐
                    │      Database           │
                    │   (SQLite3 + ORM)       │
                    └─────────────┬───────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
┌─────────┴───────┐    ┌─────────┴───────┐    ┌─────────┴───────┐
│   Containers    │    │     Hosting     │    │   External      │
│ (Docker Management)│  │ (Cloud Platforms)│   │   Services      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Interfaces

- **`Host`** - Abstraction for command execution (local/remote)
- **`Platform`** - Cloud provider interface for server management  
- **`Server`** - Individual server instance with deployment capabilities
- **`Entity`** - Database model interface for ORM operations
- **`Controller`** - Web application component interface

### Template Integration

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
        <p>Welcome, {{auth.CurrentUser.Name}}!</p>
        
        <!-- Controller method calls -->
        <ul>
            {{range products.AllProducts}}
            <li class="flex justify-between">
                <span>{{.Name}}</span>
                <span class="badge badge-primary">${{.Price}}</span>
            </li>
            {{end}}
        </ul>
        
        <!-- HTMX form with auto-refresh -->
        <form hx-post="{{host}}/products" hx-target="body">
            <input name="name" class="input input-bordered" placeholder="Product name">
            <button class="btn btn-primary">Add Product</button>
        </form>
    {{else}}
        <a href="/signin" class="btn btn-outline">Sign In</a>
    {{end}}
</body>
</html>
```

## 🔒 Security

TheSkyscape DevTools includes security best practices by default:

- **🔐 Automatic SSH Key Management** - Generates and manages SSH keys with proper permissions
- **🍪 Secure Session Handling** - HTTPOnly cookies with SameSite protection
- **🔑 Password Security** - Bcrypt hashing with configurable cost
- **🛡️ CSRF Protection** - Built-in CSRF token validation
- **📁 File Permissions** - Automatic secure file and directory permissions
- **🔒 Environment Isolation** - Separate data directories per application

## 🚀 Performance

- **⚡ Fast Startup** - Embedded assets and templates for quick boot times
- **🔄 Concurrent Database** - SQLite3 WAL mode for high concurrency
- **📦 Container Reuse** - Intelligent container lifecycle management
- **🎯 Connection Pooling** - Automatic database connection management
- **🗂️ Template Caching** - Pre-compiled templates for fast rendering

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Contributing to DevTools

```bash
# Clone the repository
git clone https://github.com/The-Skyscape/devtools.git
cd devtools

# Install dependencies
go mod download

# Run tests
go test ./...

# Run the example application
go run ./example

# Build command line tools
go build ./cmd/create-app
go build ./cmd/launch-app
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙋‍♂️ Support

- **📖 Documentation**: [INSTRUCT.md](INSTRUCT.md) for detailed usage instructions
- **🐛 Bug Reports**: [GitHub Issues](https://github.com/The-Skyscape/devtools/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/The-Skyscape/devtools/discussions)
- **📧 Email**: support@theskyscape.dev

## 🗺️ Roadmap

### v1.0 (Current)
- ✅ Core web framework with authentication
- ✅ Container management (local + remote)
- ✅ DigitalOcean cloud platform integration
- ✅ Git repository and workspace management
- ✅ SQLite3 database with dynamic ORM

### v1.1 (Next)
- 🔄 AWS platform implementation
- 🔄 Comprehensive test suite
- 🔄 CLI tools for project scaffolding
- 🔄 Performance optimizations
- 🔄 Enhanced documentation

### v1.2 (Future)
- 🔄 Google Cloud Platform integration
- 🔄 Kubernetes deployment support
- 🔄 Built-in monitoring and logging
- 🔄 Plugin system for extensions
- 🔄 GraphQL API layer

---

**Built with ❤️ by the TheSkyscape team**

*TheSkyscape DevTools - Simplifying cloud-native application development*