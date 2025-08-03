# Changelog

All notable changes to TheSkyscape DevTools will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-01-03

### Fixed
- **GitHub Actions**: Fixed release creation permissions by adding `contents: write` permission
- **CI/CD**: Resolved 403 error when creating releases from GitHub Actions workflow

### Changed
- **Installation**: Updated documentation to use direct GitHub release URLs for CLI tools
- **Workflow**: Improved release job with explicit permissions and token configuration

---

## [1.0.0] - 2025-01-03

### Added

- **CLI Tools**
  - `create-app` - Application scaffolding tool that generates complete todo applications
  - `launch-app` - Cloud deployment tool for DigitalOcean with automated SSL setup
  - Build system with Makefile and GitHub Actions for cross-platform releases

- **Web Application Framework** (`pkg/application/`)
  - MVC architecture with embedded template engine
  - HTMX integration for dynamic updates without JavaScript
  - DaisyUI theme support with corporate default
  - Built-in SSL/TLS support with automatic certificate detection
  - Template helpers and functions (`{{theme}}`, `{{host}}`, etc.)
  - Controller factory pattern with Setup() and Handle() methods

- **Authentication & Authorization** (`pkg/authentication/`)
  - JWT-based session management with secure defaults
  - Bcrypt password hashing with configurable cost
  - Role-based access control (admin/user roles)
  - Cookie-based authentication with CSRF protection
  - Flexible middleware for protecting routes
  - Built-in signin/signup/signout templates

- **Database & ORM** (`pkg/database/`)
  - Dynamic ORM built on SQLite3 with WAL mode
  - Reflection-based table creation and field mapping
  - Type-safe repository pattern with Go generics
  - Built-in migration support via golang-migrate
  - Connection pooling and concurrent access
  - Global database and repository management

- **Container Management** (`pkg/containers/`)
  - Unified Docker abstraction for local and remote hosts
  - Service lifecycle management (create, start, stop, remove)
  - Port mapping, volume mounting, and environment configuration
  - Built-in reverse proxy for container services
  - Image building capabilities with Dockerfile support

- **Multi-Cloud Deployment** (`pkg/hosting/`)
  - Platform abstraction for multiple cloud providers
  - DigitalOcean implementation with droplet management and DNS
  - Automatic SSH key generation and secure server access
  - Launch options for post-deployment configuration
  - Server lifecycle management with JSON persistence

- **Development Workspaces** (`pkg/coding/`)
  - Containerized code-server environments with Git integration
  - Git repository management with access token system
  - Per-user workspace isolation with port management
  - Automatic workspace setup and configuration

- **Application Templates**
  - Complete todo application template with authentication
  - HTMX-powered dynamic UI with DaisyUI styling
  - Proper MVC separation with models, controllers, and views
  - Database integration with todo CRUD operations
  - Responsive design with mobile-first approach

- **Build & Deployment System**
  - Makefile with clean, elegant ASCII art design
  - GitHub Actions for automated testing and releases
  - Cross-platform builds (Linux, macOS, Windows)
  - Automatic release creation on version tags
  - Artifact uploads with 30-day retention

- **Documentation**
  - Comprehensive README with CLI tool usage
  - Complete tutorial building a todo application
  - API reference documentation
  - Contributing guidelines and development setup
  - Example application demonstrating all features

### CLI Usage

**create-app**: Generate new applications
```bash
./build/create-app my-app
cd my-app && go run .
```

**launch-app**: Deploy to DigitalOcean
```bash
export DIGITAL_OCEAN_API_KEY="your-token"
./build/launch-app --name my-server --domain app.example.com --binary ./app
```

### Security Features

- Automatic SSH key management with proper permissions
- Secure session handling with HTTPOnly cookies and SameSite protection
- Environment isolation with separate data directories
- Built-in CSRF protection for forms
- JWT signing with configurable secrets
- SSL certificate automation via Let's Encrypt

### Performance Optimizations

- Embedded assets and templates for fast startup
- SQLite3 WAL mode for high concurrency
- Template caching and pre-compilation
- Intelligent container lifecycle management
- Build artifacts optimized for production deployment

### Dependencies

- Go 1.21+ with generics support
- DigitalOcean SDK (`godo`) for cloud integration
- JWT library (`golang-jwt/jwt`) for authentication
- SQLite3 driver (`mattn/go-sqlite3`) for database
- Migration library (`golang-migrate/migrate`) for schema changes
- Git toolkit (`sosedoff/gitkit`) for repository management
- Crypto library (`golang.org/x/crypto`) for security

### Platforms Supported

- **Container Hosts**: Local (Docker), Remote (SSH)
- **Cloud Providers**: DigitalOcean (full implementation)
- **Databases**: SQLite3 with dynamic schema
- **Operating Systems**: Linux, macOS, Windows

### Environment Variables

**Required:**
- `AUTH_SECRET` - JWT signing secret (required for authentication)
- `DIGITAL_OCEAN_API_KEY` - Cloud provider credentials (for deployment)

**Optional:**
- `PORT` - Application server port (default: 5000)
- `THEME` - DaisyUI theme selection (default: corporate)
- `CONGO_SSL_FULLCHAIN` - SSL certificate path
- `CONGO_SSL_PRIVKEY` - SSL private key path

### GitHub Actions Setup

The project includes automated CI/CD with these workflows:
- **Test**: Cross-platform testing with linting and security scans
- **Build**: Artifact generation for all supported platforms
- **Release**: Automatic releases on version tags

**Required GitHub permissions**: Enable "Read and write permissions" in repository Settings → Actions → General → Workflow permissions.

## [Planned for v1.1.0]

### Planned Features
- AWS platform implementation for `pkg/hosting/`
- Google Cloud Platform integration
- Enhanced CLI tools with project templates
- Multi-template support in create-app
- Environment configuration management
- Database migration commands

### Planned Improvements
- Enhanced error handling and logging
- Metrics and monitoring integration
- Configuration validation and defaults
- Development mode with hot reloading
- Plugin system for extending functionality

## [Planned for v1.2.0]

### Future Features
- Kubernetes deployment support
- Built-in monitoring and logging
- GraphQL API layer
- Advanced authentication providers (OAuth, SAML)
- Multi-database support (PostgreSQL, MySQL)
- Distributed caching with Redis

---

## Release Notes Format

For each release, we document:

- **Added** - New features and capabilities
- **Changed** - Changes to existing functionality
- **Deprecated** - Features that will be removed in future versions
- **Removed** - Features that have been removed
- **Fixed** - Bug fixes and issue resolutions
- **Security** - Security-related changes and fixes

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for information about contributing to this project and how releases are managed.