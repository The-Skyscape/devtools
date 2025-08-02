# API Reference

This document provides a comprehensive reference for the TheSkyscape DevTools packages.

## Package Overview

- [`pkg/application`](#application) - Web framework with MVC architecture
- [`pkg/authentication`](#authentication) - User authentication and sessions
- [`pkg/database`](#database) - Database abstraction and ORM
- [`pkg/containers`](#containers) - Docker container management
- [`pkg/hosting`](#hosting) - Multi-cloud server deployment
- [`pkg/coding`](#coding) - Git repositories and development workspaces

---

## application

Web application framework with embedded templates, routing, and middleware.

### Core Types

#### `App`
```go
type App struct {
    // Main application instance
}

func New(views fs.FS, opts ...Option) *App
func Serve(views fs.FS, opts ...Option) // Convenience function
```

#### `Controller`
```go
type Controller interface {
    Setup(*App)
    Handle(*http.Request) Controller
}

type BaseController struct {
    *App
    *http.Request
}
```

### Options

```go
func WithController(name string, ctrl Controller) Option
func WithViews(views fs.FS) Option
func WithHostPrefix(prefix string) Option
func WithDaisyTheme(theme string) Option
```

### Template Helpers

Available in all templates:
- `{{theme}}` - Current DaisyUI theme
- `{{host}}` - Host prefix for URLs  
- `{{req}}` - Current HTTP request
- `{{path "section" "id"}}` - Generate URL paths
- `{{path_eq "section" "id"}}` - Check current path

---

## authentication

JWT-based authentication with user management and session handling.

### Core Types

#### `Controller`
```go
type Controller struct {
    // Authentication controller
}

func (r *Repository) Controller(opts ...Option) *Controller
```

#### `User`
```go
type User struct {
    database.Model
    Avatar   string
    Name     string
    Email    string
    Handle   string
    IsAdmin  bool
    PassHash []byte
}
```

#### `Session`
```go
type Session struct {
    database.Model
    UserID string
}
```

### Repository

```go
func Manage(db *database.DynamicDB) *Repository

func (r *Repository) Signup(name, email, handle, password string, isAdmin bool) (*User, error)
func (r *Repository) Signin(ident string, password string) (*User, error)
func (r *Repository) GetUser(ident string) (*User, error)
```

### Options

```go
func WithCookie(name string) Option
func WithSetupView(view, dest string) Option
func WithSigninView(view, dest string) Option
func WithSignoutURL(url string) Option
func WithSignupHandler(fn func(*Controller, *User) http.HandlerFunc) Option
func WithSigninHandler(fn func(*Controller, *User) http.HandlerFunc) Option
```

### Template Methods

Available as `{{auth.MethodName}}`:
- `{{auth.CurrentUser}}` - Get current authenticated user
- `{{auth.CurrentSession}}` - Get current session

---

## database

Dynamic ORM with SQLite3 backend and type-safe repositories.

### Core Types

#### `Database`
```go
type Database interface {
    Model() Model
    NewModel(id string) Model
    Query(string, ...any) *Iter
}
```

#### `DynamicDB`
```go
type DynamicDB struct {
    Database
    // Dynamic database with reflection-based ORM
}

func Dynamic(engine Database, opts ...DynamicDBOption) *DynamicDB
```

#### `Repository[E Entity]`
```go
type Repository[E Entity] struct {
    DB   *DynamicDB
    Ent  E
    Type reflect.Type
}

func Manage[E Entity](db *DynamicDB, ent E) *Repository[E]
```

#### `Model`
```go
type Model struct {
    DB        Database
    ID        string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Repository Methods

```go
func (r *Repository[E]) Count() int
func (r *Repository[E]) Get(id string) (E, error)
func (r *Repository[E]) Insert(ent E) (E, error)
func (r *Repository[E]) Update(ent E) error
func (r *Repository[E]) Delete(ent E) error
func (r *Repository[E]) Search(query string, args ...any) ([]E, error)
```

### Local Database

```go
import "github.com/The-Skyscape/devtools/pkg/database/local"

func Database(name string) *database.DynamicDB
```

---

## containers

Docker container management for local and remote hosts.

### Core Types

#### `Host`
```go
type Host interface {
    SetStdin(io.Reader)
    SetStdout(io.Writer)
    SetStderr(io.Writer)
    Exec(...string) error
}
```

#### `Service`
```go
type Service struct {
    Host
    ID         string
    Status     string
    Name       string
    Image      string
    Network    string
    Privileged bool
    Entrypoint string
    Command    string
    Ports      map[int]int
    Mounts     map[string]string
    Copied     map[string]string
    Env        map[string]string
}
```

### Host Types

#### LocalHost
```go
func Local() *LocalHost

func (l *LocalHost) Launch(service *Service) error
func (l *LocalHost) Services() ([]*Service, error)
func (l *LocalHost) Service(name string) *Service
func (l *LocalHost) BuildImage(tag, context string) error
```

#### RemoteHost
```go
func Remote(s hosting.Server) *RemoteHost

func (r *RemoteHost) Launch(service *Service) error
func (r *RemoteHost) Services() ([]*Service, error)
func (r *RemoteHost) Service(name string) *Service
func (r *RemoteHost) BuildImage(tag, context string) error
```

### Service Methods

```go
func (s *Service) Start() error
func (s *Service) Stop() error
func (s *Service) Remove() error
func (s *Service) IsRunning() bool
func (s *Service) Copy(srcPath, destPath string) error
func (s *Service) Proxy(port int) http.Handler
```

### Package Functions

```go
func Launch(host Host, s *Service) error
func ListServices(host Host) ([]*Service, error)
func GetService(host Host, name string) (*Service, error)
func BuildImage(host Host, tag, context string) error
```

---

## hosting

Multi-cloud server deployment and management.

### Core Types

#### `Platform`
```go
type Platform interface {
    Launch(s *Server) (*Server, error)
    Server(id string) (*Server, error)
}
```

#### `Server`
```go
type Server interface {
    GetID() string
    GetIP() string
    GetName() string
    Launch(opts ...LaunchOption) error
    Destroy(ctx context.Context) error
    Alias(sub, domain string) error
    Env(string, string) error
    Exec(args ...string) (bytes.Buffer, bytes.Buffer, error)
    Copy(string, string) (bytes.Buffer, bytes.Buffer, error)
    Dump(string, []byte) (bytes.Buffer, bytes.Buffer, error)
    Connect(io.Reader, io.Writer, io.Writer, ...string) error
}
```

### Launch Options

```go
func WithFileUpload(path, dest string) LaunchOption
func WithBinaryData(path string, data []byte) LaunchOption
func WithSetupScript(script string, args ...string) LaunchOption
```

### DigitalOcean Platform

```go
import "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"

func Connect(apiKey string) *DigitalOceanClient

type Server struct {
    ID     int
    Name   string
    Size   string
    Region string
    Image  string
    Status string
    IP     string
}
```

---

## coding

Git repository management and development workspace orchestration.

### Core Types

#### `Repository`
```go
type Repository struct {
    // Git and workspace management
}

func Manage(db *database.DynamicDB) *Repository
```

#### `GitRepo`
```go
type GitRepo struct {
    database.Model
    Name        string
    Description string
}

func (r *Repository) NewRepo(id, name string) (*GitRepo, error)
func (r *Repository) GetRepo(id string) (*GitRepo, error)
```

#### `Workspace`
```go
type Workspace struct {
    database.Model
    Name   string
    Port   int
    Ready  bool
    RepoID string
}

func (r *Repository) NewWorkspace(name string, port int, repo *GitRepo) (*Workspace, error)
func (r *Repository) GetWorkspace(name string) (*Workspace, error)
```

#### `AccessToken`
```go
type AccessToken struct {
    database.Model
    Secret  string
    Expires time.Time
}

func (r *Repository) NewAccessToken(expires time.Time) (*AccessToken, error)
```

### Repository Methods

```go
func (repo *GitRepo) Path() string
func (repo *GitRepo) IsEmpty(branch string) bool
func (repo *GitRepo) Open(branch, path string) (*GitBlob, error)
func (repo *GitRepo) Blobs(branch, path string) ([]*GitBlob, error)
```

### Workspace Methods

```go
func (w *Workspace) Start(u *authentication.User) error
func (w *Workspace) Service() *containers.Service
func (w *Workspace) Run(cmd string) (bytes.Buffer, error)
```

---

## Environment Variables

### Required

- `AUTH_SECRET` - JWT signing secret for authentication

### Optional

- `PORT` - Application server port (default: 5000)
- `THEME` - DaisyUI theme name (default: "retro")
- `HOST_PREFIX` - URL prefix for applications
- `INTERNAL_DATA` - Custom data directory (default: ~/.theskyscape)

### SSL/TLS

- `CONGO_SSL_FULLCHAIN` - Path to SSL certificate (default: /root/fullchain.pem)
- `CONGO_SSL_PRIVKEY` - Path to SSL private key (default: /root/privkey.pem)

### Cloud Providers

- `DIGITAL_OCEAN_API_KEY` - DigitalOcean API token
- `AWS_ACCESS_KEY_ID` - AWS access key ID
- `AWS_SECRET_ACCESS_KEY` - AWS secret access key
- `AWS_REGION` - AWS region (default: us-east-1)
- `GCP_PROJECT_ID` - Google Cloud project ID
- `GCP_SERVICE_ACCOUNT_KEY` - Path to GCP service account JSON
- `GCP_ZONE` - GCP zone (default: us-central1-a)

---

## Error Handling

All packages use the `github.com/pkg/errors` library for error wrapping:

```go
import "github.com/pkg/errors"

func SomeFunction() error {
    if err := someOperation(); err != nil {
        return errors.Wrap(err, "failed to perform operation")
    }
    return nil
}
```

Common error patterns:
- `errors.New("message")` - Simple error
- `errors.Wrap(err, "context")` - Wrap existing error
- `errors.Wrapf(err, "format %s", arg)` - Wrap with formatted message

---

## Security Considerations

### Authentication
- JWT tokens are signed with `AUTH_SECRET`
- Passwords are hashed with bcrypt
- Sessions expire after 7 days by default
- Cookies use `HttpOnly` and `SameSite=Strict`

### File Permissions
- SSH keys: 0600 (private), 0644 (public)
- Data directory: 0755
- Database files: 0644

### Network Security
- HTTPS is enabled if SSL certificates are available
- SSH connections use key-based authentication
- Cloud security groups allow SSH, HTTP, and HTTPS only

---

For more examples and usage patterns, see the [Tutorial](tutorial.md) and [Examples](examples/) directory.