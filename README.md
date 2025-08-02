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

### Simple Web Application

```go
package main

import (
    "github.com/The-Skyscape/devtools/pkg/application"
    "github.com/The-Skyscape/devtools/pkg/authentication"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

func main() {
    // Initialize SQLite database
    db := local.Database("myapp.db")
    
    // Setup user authentication
    auth := authentication.Manage(db).Controller()
    
    // Create web application
    app := application.New(nil,
        application.WithController("auth", auth),
        application.WithDaisyTheme("corporate"),
    )
    
    // Start server (HTTP on :5000, HTTPS on :443 if certs available)
    app.Start()
}
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
```

### Application Structure

```
your-project/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/                # HTTP request handlers
│   ├── models/                  # Business logic models  
│   └── middleware/              # Custom middleware
├── views/
│   ├── layouts/                 # Base HTML templates
│   ├── pages/                   # Individual page templates
│   └── components/              # Reusable UI components
├── assets/
│   ├── css/                     # Stylesheets (TailwindCSS/DaisyUI)
│   ├── js/                      # JavaScript files
│   └── images/                  # Static images
└── deployments/
    ├── Dockerfile               # Container configuration
    └── docker-compose.yml       # Multi-service setup
```

## 💡 Use Cases

### 🏗️ **Infrastructure Management Platforms**
Build applications like **Terraform Cloud** or **AWS Console** that manage cloud resources across multiple providers.

### 👨‍💻 **Development Environment Platforms**  
Create services like **Replit**, **CodeSandbox**, or **Gitpod** with containerized development environments.

### 🏢 **Multi-Tenant SaaS Applications**
Develop platforms where each tenant gets isolated containers, databases, and cloud resources.

### 📊 **DevOps & Monitoring Dashboards**
Build comprehensive dashboards that deploy, monitor, and manage containerized applications.

### 🔧 **Internal Developer Tools**
Create custom tooling for your team's deployment pipelines and development workflows.

## 🎯 Examples

### Multi-Tenant Development Platform

```go
func createTenant(tenantID string) *application.App {
    // Isolated database per tenant
    db := local.Database(fmt.Sprintf("tenant_%s.db", tenantID))
    auth := authentication.Manage(db).Controller()
    coding := coding.Manage(db)
    
    return application.New(tenantViews,
        application.WithController("auth", auth),
        application.WithController("coding", coding),
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

### Development Setup

```bash
# Clone the repository
git clone https://github.com/The-Skyscape/devtools.git
cd devtools

# Install dependencies
go mod download

# Run tests
go test ./...

# Build examples
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