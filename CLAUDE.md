# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is the **TheSkyscape DevTools** repository - a Go-based toolkit for cloud infrastructure management and application deployment. The project provides unified abstractions for managing containers, cloud hosting, and web applications across multiple platforms.

## Architecture

### Core Packages

- **`pkg/application/`** - Web application framework with MVC pattern, template rendering, and SSL support
- **`pkg/containers/`** - Docker container management with local and remote host abstractions
- **`pkg/hosting/`** - Multi-cloud server deployment (DigitalOcean, AWS, GCP) with unified Platform interface
- **`pkg/authentication/`** - User authentication, sessions, and JWT token management
- **`pkg/coding/`** - Git repository management and workspace setup utilities
- **`pkg/database/`** - Database abstraction layer supporting SQLite3 with dynamic queries

### Key Design Patterns

- **Interface-based abstractions**: `Host`, `Platform`, and `Service` interfaces allow swapping implementations
- **Option pattern**: Configuration through variadic option functions (e.g., `WithFileUpload()`, `WithSetupScript()`)
- **Embedded file systems**: Views and static assets embedded using Go's `embed` package
- **Plugin architecture**: Cloud platforms implemented as separate packages under `platforms/`

### Commands

- **`cmd/create-app/`** - Application creation utility (currently basic Hello World)
- **`cmd/launch-app/`** - Application launcher utility (currently basic Hello World)

## Common Development Commands

### Building
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
go run ./cmd/create-app
go run ./cmd/launch-app
```

## Key Dependencies

- **Cloud SDKs**: DigitalOcean (`godo`), AWS support planned
- **Database**: SQLite3 driver (`mattn/go-sqlite3`) 
- **Security**: JWT tokens (`golang-jwt/jwt`), bcrypt (`golang.org/x/crypto`)
- **Git**: Repository management (`sosedoff/gitkit`)
- **Migration**: Database migrations (`golang-migrate/migrate`)

## Environment Variables

### Container Services
- `PORT` - Application server port (default: 5000)
- `CONGO_SSL_FULLCHAIN` - SSL certificate path (default: /root/fullchain.pem)
- `CONGO_SSL_PRIVKEY` - SSL private key path (default: /root/privkey.pem)

### Cloud Platforms
- `DIGITAL_OCEAN_API_KEY` - DigitalOcean API token
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` - AWS credentials
- `GCP_PROJECT_ID`, `GCP_SERVICE_ACCOUNT_KEY`, `GCP_ZONE` - GCP credentials

## Integration Points

- **Docker Runtime** - All container operations require Docker daemon
- **SSH Keys** - Automatic SSH key generation for cloud server access (stored in `~/.ssh/`)
- **File System** - Template views embedded at build time, external views loaded at runtime
- **HTTP/HTTPS** - Dual-protocol web server with automatic SSL certificate detection