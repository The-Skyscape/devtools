# Authentication Package

## Overview

The authentication package provides a complete authentication system built on the Collection pattern. It wraps database operations with authentication-specific functionality like user management, sessions, email verification, and password reset flows.

## Architecture

### Collection Pattern

The core of the package is the `Collection` struct that wraps a database connection:

```go
// Create authentication collection
Auth := authentication.Manage(db)

// The Collection provides repositories for users, sessions, and tokens
users := Auth.Users        // User repository
sessions := Auth.Sessions   // Session repository
```

### Optional Controller

For web applications, an optional `Controller` provides HTTP handlers and template methods:

```go
// Get the controller (optional)
controller := Auth.Controller()

// Add to application
application.Serve(views,
    application.WithController("auth", controller),
)
```

## Usage

### Basic Setup

```go
import (
    "github.com/The-Skyscape/devtools/pkg/authentication"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

// Initialize database
db := local.Database("app.db", authentication.Migrations)

// Create authentication collection
Auth := authentication.Manage(db)
```

### User Management

```go
// Sign up new user
user, err := Auth.Signup(name, email, handle, password, isAdmin)

// Sign in user
user, err := Auth.Signin(emailOrHandle, password)

// Get user by ID
user, err := Auth.Users.Get(userID)

// Search users
users, err := Auth.Users.Search("WHERE IsAdmin = ?", true)

// Count users
count := Auth.Users.Count("")
```

### Session Management

```go
// Create session
session, err := Auth.Sessions.Insert(&authentication.Session{
    UserID: user.ID,
})

// Get session
session, err := Auth.Sessions.Get(sessionID)

// Delete expired sessions
Auth.Sessions.DeleteExpired()
```

### Email Verification

```go
// Send verification email
token, err := Auth.VerifyEmail(user)

// In email template, create link like:
// https://example.com/verify?token={{token}}

// Verify the token
err := Auth.VerifyToken(token)
```

### Password Reset

```go
// Request password reset
token, err := Auth.ResetPassword(user)

// In email template, create link like:
// https://example.com/reset?token={{token}}

// Reset password with token
err := Auth.ConsumeToken(token, newPassword)
```

### Authentication Middleware

```go
// Authenticate request (returns user, session, error)
user, session, err := Auth.Authenticate(r)

// Use as middleware
app.Serve("GET /admin", "admin.html", Auth.Required)
app.Serve("GET /", "home.html", Auth.Optional)
```

## Controller Usage (Web Applications)

### Setup

```go
func main() {
    // Get controller from collection
    authController := models.Auth.Controller()

    // Add to application
    application.Serve(views,
        application.WithController("auth", authController),
        application.WithController(controllers.Home()),
    )
}
```

### In Templates

```html
{{if auth.CurrentUser}}
    <p>Welcome, {{auth.CurrentUser.Name}}!</p>
    <a href="/signout">Sign Out</a>
{{else}}
    <a href="/signin">Sign In</a>
{{end}}

{{if auth.IsAdmin}}
    <a href="/admin">Admin Panel</a>
{{end}}
```

### HTTP Handlers

The Controller provides these built-in handlers:

- `HandleSignup(w, r)` - Process signup form
- `HandleSignin(w, r)` - Process signin form
- `HandleSignout(w, r)` - Sign out user
- `HandleVerify(w, r)` - Verify email token
- `HandleForgotPassword(w, r)` - Send password reset email
- `HandleResetPassword(w, r)` - Reset password with token

### Access Control

```go
// In controller setup
func (c *MyController) Setup(app *application.App) {
    auth := app.Use("auth").(*authentication.Controller)

    // Public route
    http.Handle("GET /", app.Serve("home.html", auth.Optional))

    // Protected route
    http.Handle("GET /admin", app.Serve("admin.html", auth.Required))

    // Admin only
    http.Handle("GET /settings", app.Serve("settings.html", auth.AdminOnly))
}
```

## Data Models

### User

```go
type User struct {
    ID           string
    Name         string
    Email        string
    Handle       string
    Password     string // Hashed
    IsAdmin      bool
    IsVerified   bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Session

```go
type Session struct {
    ID        string
    UserID    string
    CreatedAt time.Time
    ExpiresAt time.Time
}
```

## Security Features

- **Password Hashing**: Uses bcrypt with default cost
- **Session Tokens**: JWT tokens with configurable expiration
- **Email Verification**: Secure token generation and validation
- **Password Reset**: Time-limited tokens with one-time use
- **Input Validation**: Email, username, and password validation
- **SQL Injection Protection**: Parameterized queries throughout

## Configuration Options

```go
// Create collection with options
Auth := authentication.Manage(db,
    authentication.WithSessionDuration(24 * time.Hour),
    authentication.WithTokenExpiry(1 * time.Hour),
)
```

## Best Practices

1. **Always use HTTPS** in production for secure cookie transmission
2. **Set secure cookie flags** when using the Controller
3. **Implement rate limiting** on authentication endpoints
4. **Log authentication events** for security auditing
5. **Use strong JWT secrets** (set via AUTH_SECRET environment variable)
6. **Regularly clean expired sessions** with a background job

## Testing

```go
import (
    "testing"
    "github.com/The-Skyscape/devtools/pkg/authentication"
    "github.com/The-Skyscape/devtools/pkg/testutils"
)

func TestAuthentication(t *testing.T) {
    // Setup test database
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(t, db)

    // Create auth collection
    auth := authentication.Manage(db)

    // Test signup
    user, err := auth.Signup("Test", "test@example.com", "test", "password", false)
    testutils.AssertNoError(t, err)
    testutils.AssertNotNil(t, user)

    // Test signin
    user, err = auth.Signin("test", "password")
    testutils.AssertNoError(t, err)
    testutils.AssertEqual(t, "test@example.com", user.Email)
}
```

## Migration from Direct Database Access

If you're currently using direct database queries for authentication:

```go
// OLD: Direct database queries
user := &User{}
db.QueryRow("SELECT * FROM users WHERE email = ?", email).Scan(&user)
bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

// NEW: Use authentication collection
user, err := Auth.Signin(email, password)
```

## Related Packages

- `pkg/database` - Database abstraction layer
- `pkg/application` - Web application framework
- `pkg/emailing` - Email sending for verification/reset emails
- `pkg/testutils` - Testing utilities