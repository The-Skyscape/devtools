# API Reference

Complete reference for TheSkyscape DevTools packages and CLI tools.

## CLI Tools

### create-app

Generates complete TheSkyscape applications with proper MVC structure.

**Installation:**
```bash
curl -L -o create-app https://github.com/The-Skyscape/devtools/releases/download/v1.0.1/create-app
chmod +x create-app
```

**Usage:**
```bash
./create-app [app-name]
```

**Generated Structure:**
- `controllers/` - HTTP handlers with factory functions
- `models/` - Database models with Table() methods  
- `views/` - HTML templates with HTMX integration
- `main.go` - Application entry point

**Example:**
```bash
./create-app my-todo-app
cd my-todo-app
export AUTH_SECRET="your-secret"
go run .
```

### launch-app

Deploys applications to DigitalOcean with automated setup.

**Installation:**
```bash
curl -L -o launch-app https://github.com/The-Skyscape/devtools/releases/download/v1.0.1/launch-app
chmod +x launch-app
```

**Usage:**
```bash
./launch-app [options]
```

**Flags:**
- `--name` - Server name (default: skyscape-app)
- `--region` - Server region (default: sfo3)  
- `--size` - Server size (default: s-2vcpu-2gb)
- `--domain` - Domain for SSL certificates
- `--binary` - Path to application binary

**Example:**
```bash
export DIGITAL_OCEAN_API_KEY="your-token"
./launch-app --name production --domain app.example.com --binary ./app
```

---

## pkg/application

Web application framework with MVC pattern and embedded templates.

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

func (c *BaseController) Setup(app *App)
func (c *BaseController) Handle(r *http.Request) Controller
func (c *BaseController) Render(w http.ResponseWriter, r *http.Request, template string, data interface{})
func (c *BaseController) Refresh(w http.ResponseWriter, r *http.Request)
```

#### `Model`
```go
type Model struct {
    ID        string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Options

```go
func WithController(name string, controller Controller) Option
func WithDaisyTheme(theme string) Option
func WithPort(port string) Option
func WithHostPrefix(prefix string) Option
```

### Template Helpers

Available in all templates:
- `{{theme}}` - Current DaisyUI theme
- `{{host}}` - Host prefix for URLs  
- `{{req}}` - Current HTTP request
- `{{path "section" "id"}}` - Generate URL paths
- `{{auth.CurrentUser}}` - Current authenticated user

---

## pkg/authentication

JWT-based authentication with user management and session handling.

### Core Types

#### `Manager`
```go
type Manager struct {
    // Authentication manager
}

func Manage(db database.Database) *Manager
func (m *Manager) Controller(options ...Option) Controller
func (m *Manager) Required(app *App, r *http.Request) string
```

#### `User`
```go
type User struct {
    application.Model
    Name     string
    Email    string
    Handle   string
    IsAdmin  bool
    PassHash []byte
}
```

### Options

```go
func WithCookie(name string) Option
func WithSignoutURL(url string) Option
func WithSigninView(view, dest string) Option
func WithSignupView(view, dest string) Option
```

### Template Methods

Available as `{{auth.MethodName}}`:
- `{{auth.CurrentUser}}` - Get current authenticated user
- `{{auth.SigninURL}}` - Sign-in page URL
- `{{auth.SignupURL}}` - Sign-up page URL

---

## pkg/database

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

#### `Repository[T any]`
```go
type Repository[T any] struct {
    DB   Database
    Type reflect.Type
}

func Manage[T any](db Database, model *T) *Repository[T]
```

### Repository Methods

```go
func (r *Repository[T]) Get(id string) (*T, error)
func (r *Repository[T]) Insert(model *T) error
func (r *Repository[T]) Update(model *T) error
func (r *Repository[T]) Delete(model *T) error
func (r *Repository[T]) Search(query string, args ...interface{}) ([]*T, error)
```

### Local Database

```go
import "github.com/The-Skyscape/devtools/pkg/database/local"

func Database(filename string) database.Database
```

---

## pkg/containers

Docker container management for local and remote hosts.

### Core Types

#### `Host`
```go
type Host interface {
    BuildImage(tag, context string) error
    Launch(service *Service) error
    Stop(name string) error
    Remove(name string) error
}
```

#### `Service`
```go
type Service struct {
    Name       string
    Image      string
    Entrypoint string
    Network    string
    Privileged bool
    Ports      map[int]int
    Mounts     map[string]string
    Copied     map[string]string
    Env        map[string]string
}

func (s *Service) Start() error
func (s *Service) Stop() error
```

### Host Types

```go
func Local() Host                           // Local Docker host
func Remote(server hosting.Server) Host     // Remote Docker host via SSH
```

---

## pkg/hosting

Multi-cloud server deployment and management.

### Core Types

#### `Platform`
```go
type Platform interface {
    Launch(server Server, options ...LaunchOption) (Server, error)
}
```

#### `Server`
```go
type Server interface {
    GetID() string
    GetIP() string
    Alias(subdomain, domain string) error
    Run(script string) error
    Exec(command string) (string, string, error)
}
```

### Launch Options

```go
func WithFileUpload(localPath, remotePath string) LaunchOption
func WithBinaryData(remotePath string, data []byte) LaunchOption
func WithSetupScript(script string) LaunchOption
```

### DigitalOcean Platform

```go
import "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"

func Connect(apiKey string) Platform

var ApiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")

type Server struct {
    Name   string
    Size   string
    Region string
    Image  string
    Status string
}
```

---

## pkg/coding

Git repository management and development workspace orchestration.

### Core Types

#### `Manager`
```go
type Manager struct {
    // Repository and workspace manager
}

func Manage(db database.Database) *Manager
func (m *Manager) NewRepo(id, name string) (*Repository, error)
func (m *Manager) NewWorkspace(userID string, port int, repo *Repository) (*Workspace, error)
```

#### `Repository`
```go
type Repository struct {
    application.Model
    Name        string
    Description string
}

func (r *Repository) Path() string
func (r *Repository) IsEmpty(branch string) bool
```

#### `Workspace`
```go
type Workspace struct {
    application.Model
    Name   string
    Port   int
    Ready  bool
    RepoID string
}

func (w *Workspace) Start(user *authentication.User) error
func (w *Workspace) Stop() error
```

---

## Environment Variables

### Required

- `AUTH_SECRET` - JWT signing secret (required for authentication)

### Optional Application

- `PORT` - Server port (default: 5000)
- `THEME` - DaisyUI theme (default: corporate)

### SSL Configuration

- `CONGO_SSL_FULLCHAIN` - SSL certificate path
- `CONGO_SSL_PRIVKEY` - SSL private key path

### Cloud Platforms

- `DIGITAL_OCEAN_API_KEY` - DigitalOcean API token

---

## Error Handling

All packages use consistent error handling patterns:

```go
import "github.com/pkg/errors"

// Simple error
return errors.New("something went wrong")

// Wrapped error
return errors.Wrap(err, "failed to perform operation")

// Formatted error
return errors.Wrapf(err, "failed to process %s", filename)
```

### Error Templates

Use `error-message.html` template for consistent error display:

```go
c.Render(w, r, "error-message.html", errors.New("Something went wrong"))
```

---

## Best Practices

### Controller Patterns

1. **Factory Functions**: Return `(string, *Controller)`
2. **Setup Method**: Register routes in `Setup(app *App)`
3. **Handle Method**: Return controller instance per request
4. **Template Methods**: Public methods accessible in views
5. **HTTP Handlers**: Private methods for form processing

### Model Patterns

1. **Embed application.Model**: Provides ID, timestamps
2. **Implement Table() method**: Return database table name
3. **Use global repositories**: Created in `models/database.go`
4. **Business logic methods**: Keep data operations in models

### Template Patterns

1. **Unique filenames**: Avoid path conflicts in global namespace
2. **Layout components**: Use `{{define "layout/start"}}` patterns
3. **Partials directory**: Reusable components in `views/partials/`
4. **HTMX integration**: Use `hx-*` attributes for dynamic updates

### Security Patterns

1. **Protect routes**: Use `app.ProtectFunc()` for authenticated endpoints
2. **Validate input**: Check form values before processing
3. **Error handling**: Use proper error templates
4. **Environment secrets**: Store sensitive data in environment variables

---

## Security Features

### Automatic Security

- **JWT Authentication** - Secure session management
- **Password Hashing** - Bcrypt with proper salt
- **CSRF Protection** - Built into forms
- **SSH Key Management** - Auto-generated for cloud access
- **Secure Cookies** - HTTPOnly with SameSite protection
- **SSL Certificates** - Automatic generation via Let's Encrypt

### File Permissions

- SSH keys: 0600 (private), 0644 (public)
- Data directory: 0755
- Database files: 0644

---

For more examples and usage patterns, see the [Tutorial](tutorial.md) and [README](../README.md).