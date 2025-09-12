# Hosting Package

The hosting package provides a unified interface for managing cloud servers across multiple platforms. It abstracts the complexity of different cloud providers behind a simple, consistent API.

## Features

- **Multi-Cloud Support** - DigitalOcean, AWS, GCP (extensible)
- **Unified Interface** - Same API regardless of provider
- **Server Management** - Launch, configure, destroy servers
- **SSH Integration** - Execute commands, copy files, interactive sessions
- **Mock Implementation** - Comprehensive mocking for testing
- **Domain Management** - DNS aliasing support

## Installation

```go
import "github.com/The-Skyscape/devtools/pkg/hosting"
```

## Quick Start

```go
import (
    "github.com/The-Skyscape/devtools/pkg/hosting"
    "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"
)

// Create a platform
platform := digitalocean.New(apiToken)

// Launch a server
server, err := platform.Launch(&hosting.Server{
    Name:   "my-app-server",
    Size:   "s-2vcpu-4gb",
    Region: "nyc3",
    Image:  "ubuntu-22-04-x64",
})

// Configure the server
server.Env("APP_NAME", "myapp")
server.Exec("apt-get", "update")
server.Copy("./app", "/opt/app")

// Create a domain alias
server.Alias("app", "example.com")
```

## Core Interfaces

### Platform Interface

```go
type Platform interface {
    Launch(s *Server) (*Server, error)
    Server(id string) (*Server, error)
}
```

### Server Interface

```go
type Server interface {
    GetID() string
    GetIP() string
    GetName() string
    
    Launch(opts ...LaunchOption) error
    Destroy(ctx context.Context) error
    Alias(sub, domain string) error
    
    Env(key, value string) error
    Exec(args ...string) (stdout, stderr bytes.Buffer, error)
    Copy(src, dst string) (stdout, stderr bytes.Buffer, error)
    Dump(path string, data []byte) (stdout, stderr bytes.Buffer, error)
    Connect(stdin io.Reader, stdout, stderr io.Writer, args ...string) error
}
```

## Platform Implementations

### DigitalOcean

```go
import "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"

platform := digitalocean.New(apiToken,
    digitalocean.WithRegion("nyc3"),
    digitalocean.WithProject(projectID),
)
```

### Mock Platform (Testing)

```go
// Create mock platform for testing
mock := hosting.NewMockPlatform()

// Control test behavior
mock.FailNext("simulated error")
server, err := mock.Launch(nil) // Returns error

mock.Reset()
server, err := mock.Launch(nil) // Succeeds

// Verify operations
assert.Equal(t, 1, mock.GetLaunchCount())
servers := mock.GetServers()
assert.Equal(t, "mock-server-1", servers[0].GetID())
```

## Server Operations

### Environment Variables

```go
// Set environment variables
server.Env("DATABASE_URL", "postgres://...")
server.Env("REDIS_URL", "redis://...")
```

### Command Execution

```go
// Execute commands
stdout, stderr, err := server.Exec("docker", "ps")
if err != nil {
    log.Printf("Error: %s", stderr.String())
}
fmt.Println(stdout.String())
```

### File Operations

```go
// Copy files to server
stdout, stderr, err := server.Copy("./config.yaml", "/etc/app/config.yaml")

// Write data directly
data := []byte("Hello, World!")
stdout, stderr, err := server.Dump("/tmp/message.txt", data)
```

### Interactive Connection

```go
// Open interactive SSH session
err := server.Connect(os.Stdin, os.Stdout, os.Stderr, "bash")
```

## Launch Options

```go
// Launch with options
err := server.Launch(
    hosting.WithSSHKeys(sshKeyIDs...),
    hosting.WithUserData(cloudInitScript),
    hosting.WithTags("env:production", "app:myapp"),
    hosting.WithVolume(volumeID),
)
```

## Testing

### Using Mock Platform

```go
func TestServerDeployment(t *testing.T) {
    // Create mock platform
    platform := hosting.NewMockPlatform()
    
    // Test server launch
    server, err := platform.Launch(nil)
    assert.NoError(t, err)
    assert.NotNil(t, server)
    
    // Verify server properties
    assert.Equal(t, "10.0.0.1", (*server).GetIP())
    assert.Equal(t, "mock-server-1", (*server).GetID())
    
    // Test server operations
    mockServer := (*server).(*hosting.MockServer)
    
    err = mockServer.Env("KEY", "value")
    assert.NoError(t, err)
    assert.Equal(t, "value", mockServer.GetEnv()["KEY"])
    
    stdout, _, err := mockServer.Exec("echo", "test")
    assert.NoError(t, err)
    assert.Contains(t, stdout.String(), "test")
    
    // Verify exec calls
    calls := mockServer.GetExecCalls()
    assert.Equal(t, 1, len(calls))
    assert.Equal(t, []string{"echo", "test"}, calls[0])
    
    // Test destruction
    err = mockServer.Destroy(context.Background())
    assert.NoError(t, err)
    assert.True(t, mockServer.IsDestroyed())
}
```

### Simulating Failures

```go
func TestErrorHandling(t *testing.T) {
    platform := hosting.NewMockPlatform()
    
    // Simulate launch failure
    platform.FailNext("network error")
    _, err := platform.Launch(nil)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "network error")
    
    // Reset and succeed
    platform.Reset()
    server, err := platform.Launch(nil)
    assert.NoError(t, err)
    assert.NotNil(t, server)
}
```

## Best Practices

1. **Use interfaces** - Depend on Platform/Server interfaces, not implementations
2. **Handle errors** - Cloud operations can fail; implement retry logic
3. **Clean up resources** - Always destroy servers in defer blocks
4. **Use contexts** - Pass contexts for cancellation support
5. **Mock in tests** - Use MockPlatform for unit tests
6. **Secure credentials** - Never hardcode API tokens

## Example: Complete Server Setup

```go
func DeployApplication(platform hosting.Platform) error {
    // Launch server
    server, err := platform.Launch(&hosting.Server{
        Name:   "app-server",
        Size:   "s-2vcpu-4gb",
        Region: "nyc3",
    })
    if err != nil {
        return err
    }
    
    // Ensure cleanup
    defer func() {
        if err != nil {
            server.Destroy(context.Background())
        }
    }()
    
    // Configure server
    server.Env("APP_ENV", "production")
    
    // Install dependencies
    cmds := [][]string{
        {"apt-get", "update"},
        {"apt-get", "install", "-y", "docker.io"},
        {"systemctl", "start", "docker"},
    }
    
    for _, cmd := range cmds {
        stdout, stderr, err := server.Exec(cmd...)
        if err != nil {
            return fmt.Errorf("failed to run %v: %s", cmd, stderr.String())
        }
        log.Printf("Executed %v: %s", cmd, stdout.String())
    }
    
    // Deploy application
    server.Copy("./app", "/opt/app")
    server.Exec("docker", "build", "-t", "myapp", "/opt/app")
    server.Exec("docker", "run", "-d", "-p", "80:8080", "myapp")
    
    // Setup domain
    return server.Alias("app", "example.com")
}
```

## Platform-Specific Features

### DigitalOcean

- Droplet management
- Floating IPs
- Block storage volumes
- DNS management
- Project organization

### AWS (planned)

- EC2 instances
- EBS volumes
- Route53 DNS
- VPC networking

### GCP (planned)

- Compute Engine instances
- Persistent disks
- Cloud DNS

## Error Handling

```go
// Check specific error types
if err := server.Launch(); err != nil {
    switch {
    case strings.Contains(err.Error(), "insufficient funds"):
        // Handle billing issue
    case strings.Contains(err.Error(), "limit exceeded"):
        // Handle quota issue
    default:
        // Generic error handling
    }
}
```

## License

Part of TheSkyscape DevTools - see root LICENSE file