# Hosting Package - Platform Interface Pattern

## Overview

The hosting package provides a unified interface for managing cloud infrastructure across multiple providers (DigitalOcean, AWS, GCP, etc.). It implements a powerful **Platform-Resource Pattern** that creates a clean bidirectional relationship between platforms and their resources.

## Core Architecture Pattern

### The Platform-Resource Pattern

This package implements a pattern where:
1. **The client IS the platform** - Platform implementations (like DigitalOceanClient) directly implement the Platform interface
2. **Resources hold their Platform reference** - Server, Volume, and Domain structs contain a Platform field
3. **Bidirectional collaboration** - Platform creates resources, resources delegate operations back to Platform

```go
// The Platform creates and manages resources
platform := digitalocean.Connect(apiKey)
server := &hosting.Server{Name: "my-server", Size: "s-2vcpu-4gb"}
server, err := platform.NewServer(server)  // Platform creates and assigns itself

// Resources delegate operations back to their Platform
err = server.Destroy()  // Calls server.Platform.DestroyServer(server.ID)
```

### Why This Pattern Works

This is similar to Go's standard library patterns:
- `http.Request` holds reference to `http.Client`
- `sql.Rows` holds reference to `sql.DB`
- Resources know their origin and can perform operations through it

**This is NOT a circular dependency** - it's intentional bidirectional collaboration where each side has clear responsibilities.

## Interface Design

### Platform Interface

```go
type Platform interface {
    // Server operations
    NewServer(server *Server) (*Server, error)
    GetServer(id string) (*Server, error)
    AllServers() ([]*Server, error)
    DestroyServer(id string) error

    // Volume operations
    NewVolume(volume *Volume) (*Volume, error)
    GetVolume(id string) (*Volume, error)
    AllVolumes() ([]*Volume, error)
    MountVolume(volume *Volume, server *Server) error
    UnmountVolume(volume *Volume, server *Server) error
    DestroyVolume(id string) error

    // Domain operations
    LookupDomain(domain *Domain) (*Domain, error)
    AssignDomain(server *Server, domain *Domain) error
    DestroyDomain(id string) error
}
```

### Resource Structs

```go
type Server struct {
    Platform Platform  // Reference to creating platform
    ID       string
    IP       string
    Loc      string    // Location/Region
    Size     string    // Server size
    Name     string
    Status   string
}

type Volume struct {
    Platform Platform  // Reference to creating platform
    ID       string
    Loc      string
    Size     int       // Size in GB
    Name     string
}

type Domain struct {
    Platform Platform  // Reference to creating platform
    ID       string
    Sub      string    // Subdomain
    Name     string    // Base domain
    Type     string    // Record type (A, CNAME, etc.)
    Data     string    // Record data (IP address, etc.)
}
```

### Resource Methods

Resources have convenience methods that delegate to their Platform:

```go
func (s *Server) Destroy() error {
    return s.Platform.DestroyServer(s.ID)
}

func (v *Volume) Mount(server *Server) error {
    return v.Platform.MountVolume(v, server)
}

func (d *Domain) Assign(server *Server) error {
    return d.Platform.AssignDomain(server, d)
}
```

## Implementation Guidelines

### Creating a New Platform

1. **Implement the Platform interface directly**:

```go
type MyPlatformClient struct {
    client   *sdk.Client
    apiKey   string
    // Platform-specific fields
}

// Ensure it implements Platform
var _ Platform = (*MyPlatformClient)(nil)
```

2. **Use ClientOption pattern for configuration**:

```go
func Connect(apiKey string, opts ...ClientOption) *MyPlatformClient {
    client := &MyPlatformClient{
        apiKey: apiKey,
        // defaults
    }

    for _, opt := range opts {
        opt(client)
    }

    return client
}

type ClientOption func(*MyPlatformClient)

func WithRegion(region string) ClientOption {
    return func(c *MyPlatformClient) {
        c.defaultRegion = region
    }
}
```

3. **Always assign Platform reference when creating resources**:

```go
func (p *MyPlatformClient) NewServer(server *Server) (*Server, error) {
    // CRITICAL: Assign platform reference
    server.Platform = p

    // Create the actual resource
    response, err := p.client.CreateServer(server.Name, server.Size)
    if err != nil {
        return nil, err
    }

    // Update fields from response
    server.ID = response.ID
    server.IP = response.IP
    server.Status = response.Status

    return server, nil
}
```

## Usage Examples

### Basic Usage

```go
// Connect to platform
platform := digitalocean.Connect(apiKey,
    digitalocean.WithProjectID("proj-123"),
    digitalocean.WithDefaultImage("ubuntu-22-04"),
)

// Create a server
server := &hosting.Server{
    Name: "web-server",
    Size: "s-2vcpu-4gb",
    Loc:  "nyc3",
}
server, err := platform.NewServer(server)

// Server operations through Platform reference
_, _, err = server.Exec("apt-get update")
err = server.Copy("./app", "/opt/app")

// Create and mount a volume
volume := &hosting.Volume{
    Name: "data",
    Size: 100,  // 100GB
    Loc:  "nyc3",
}
volume, err = platform.NewVolume(volume)
err = volume.Mount(server)  // Delegates to Platform

// DNS management
domain := &hosting.Domain{
    Name: "example.com",
    Sub:  "app",
    Type: "A",
}
err = platform.AssignDomain(server, domain)

// Cleanup
err = server.Destroy()  // Delegates to Platform
```

### Testing with Mock

```go
// Create mock platform
mock := mock.New()

// Control test behavior
mock.FailNext("simulated error")

// Use exactly like real platform
server := &hosting.Server{Name: "test"}
server, err := mock.NewServer(server)
// err will be "simulated error"

mock.Reset()  // Clear failure state
```

## Design Principles

### 1. **Direct Implementation Over Abstraction**
- Don't create unnecessary interfaces (no ServerRef, VolumeRef)
- Use concrete types (*Server, *Volume, *Domain)
- Platform implementations ARE the interface

### 2. **Platform-Centric Operations**
- Platform is responsible for lifecycle (create, get, destroy)
- Resources are data + convenience methods
- All operations flow through Platform

### 3. **Bidirectional by Design**
- Platform → Resource: Creates and manages
- Resource → Platform: Delegates operations
- This is intentional, not a code smell

### 4. **Consistent Patterns**
- All platforms follow the same structure
- All resources have Platform reference
- All operations have predictable signatures

## Migration Guide

### From Old Pattern (with Refs)

```go
// OLD
var server hosting.ServerRef
server, err = platform.Launch(opts...)
id := server.GetID()
ip := server.GetIP()

// NEW
server := &hosting.Server{Name: "my-server"}
server, err = platform.NewServer(server)
id := server.ID
ip := server.IP
```

### Key Changes
1. No more Ref interfaces - use structs directly
2. No more getter methods - use fields directly
3. Platform methods take/return struct pointers
4. Resources always have Platform reference

## Common Patterns

### Resource Creation Flow

```go
// 1. Create resource struct with desired state
resource := &hosting.Server{
    Name: "desired-name",
    Size: "desired-size",
}

// 2. Platform creates actual resource
resource, err = platform.NewServer(resource)
// Platform has:
// - Assigned itself as resource.Platform
// - Created the cloud resource
// - Updated resource fields with actual values (ID, IP, etc.)

// 3. Resource can now perform operations
err = resource.Destroy()  // Works because Platform is set
```

### Platform Initialization

```go
// Simple initialization
platform := provider.Connect(apiKey)

// With options
platform := provider.Connect(apiKey,
    provider.WithProjectID("proj-123"),
    provider.WithDefaultRegion("nyc3"),
    provider.WithTimeout(30 * time.Second),
)
```

## Extending the Pattern to Other Packages

The Platform-Resource pattern demonstrated in the hosting package can be applied
to other devtools packages. The key principles remain the same:

1. **The client IS the platform** - No unnecessary abstraction layers
2. **Resources hold Platform reference** - For delegating operations
3. **Bidirectional collaboration** - Platform creates, resources delegate back
4. **Concrete types over interfaces** - Use structs directly, not Ref interfaces

When implementing new packages, study the hosting package implementation:
- How Platform assigns itself to resources during creation
- How resources delegate operations back to Platform
- How the ClientOption pattern provides flexibility
- How mock implementations facilitate testing

The hosting package serves as the reference implementation for this pattern
across the devtools ecosystem.

## Testing

### Unit Testing

```go
func TestServerCreation(t *testing.T) {
    // Use mock platform
    platform := mock.New()

    // Test server creation
    server := &hosting.Server{Name: "test"}
    created, err := platform.NewServer(server)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if created.Platform != platform {
        t.Error("platform reference not set")
    }

    if created.ID == "" {
        t.Error("server ID not set")
    }
}
```

### Integration Testing

```go
func TestRealPlatform(t *testing.T) {
    if os.Getenv("INTEGRATION_TEST") != "true" {
        t.Skip("skipping integration test")
    }

    platform := digitalocean.Connect(os.Getenv("DO_API_KEY"))

    server := &hosting.Server{
        Name: fmt.Sprintf("test-%d", time.Now().Unix()),
        Size: "s-1vcpu-1gb",
        Loc:  "nyc3",
    }

    server, err := platform.NewServer(server)
    if err != nil {
        t.Fatal(err)
    }

    defer server.Destroy()  // Cleanup

    // Test operations
    _, _, err = server.Exec("echo", "test")
    if err != nil {
        t.Error(err)
    }
}
```

## Troubleshooting

### Common Issues

1. **"Platform is nil" errors**
   - Ensure you're using platform.NewServer(), not creating Server directly
   - Platform must assign itself during creation

2. **"Method not implemented" errors**
   - Check that platform fully implements the Platform interface
   - Use `var _ Platform = (*YourClient)(nil)` to verify at compile time

3. **Resource operations failing**
   - Verify Platform reference is set
   - Check that resource was created through Platform, not manually

## Platform Implementations

### Current
- **DigitalOcean** (`platforms/digitalocean`) - Full implementation
- **Mock** (`platforms/mock`) - For testing

### Planned
- **AWS** (`platforms/aws`) - EC2, EBS, Route53
- **Google Cloud** (`platforms/gcp`) - Compute Engine, Persistent Disks, Cloud DNS
- **Azure** (`platforms/azure`) - Virtual Machines, Managed Disks, DNS Zones

## Contributing

When adding new platforms:

1. Follow the Platform-Resource pattern exactly
2. Implement all Platform interface methods
3. Always set Platform reference in resources
4. Use ClientOption pattern for configuration
5. Provide comprehensive tests using mock
6. Document platform-specific features

## Summary

The Platform-Resource pattern provides:
- **Clean separation** of concerns
- **Strong interfaces** for scalability
- **Flexibility** through platform implementations
- **Consistency** across different providers
- **Testability** through mock implementation

This pattern is intentionally bidirectional, creating a powerful and flexible system for managing cloud resources across any provider.