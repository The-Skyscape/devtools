# Emailing Package

A simple, Go-idiomatic email package that provides sending, tracking, and templating in one unified interface.

## Philosophy

- **Simplicity over features** - One `Send` method that handles all cases
- **Always track** - Every email is recorded in the database for audit trails
- **Flexible data** - Templates accept any type, no forced structures
- **Graceful degradation** - Sending succeeds even if tracking fails

## Quick Start

```go
import (
    "github.com/The-Skyscape/devtools/pkg/emailing"
    "github.com/The-Skyscape/devtools/pkg/emailing/providers"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

// Initialize
db := local.Database("emails.db")
emails := emailing.Manage(db,
    emailing.WithFrom("noreply@example.com", "My App"),
    emailing.WithProvider(providers.NewResendProvider(apiKey, nil)),
    emailing.WithTemplateFS(templatesFS),
)

// Send a simple text email
err := emails.Send("user@example.com",
    emailing.WithSubject("Hello!"),
    emailing.WithPlainText("Hello, world!"),
)

// Send HTML email with text fallback
err := emails.Send("user@example.com",
    emailing.WithSubject("Welcome!"),
    emailing.WithHTML("<h1>Welcome to our app</h1>"),
    emailing.WithPlainText("Welcome to our app"),
)

// Send with template
err := emails.Send("user@example.com",
    emailing.WithTemplate("welcome", map[string]any{
        "UserName": "John",
        "Subject":  "Welcome to My App",
    }),
)
```

## The Send Method

The `Send(to string, opts ...SendOption)` method uses functional options for configuration:

### Simple Text Email
```go
// Send a plain text email
err := emails.Send("user@example.com",
    emailing.WithSubject("Your order has shipped!"),
    emailing.WithPlainText("Your order #1234 is on its way."),
)
```

### HTML Email with Fallback
```go
// Send HTML email with plain text fallback
err := emails.Send("user@example.com",
    emailing.WithSubject("Order Confirmation"),
    emailing.WithHTML("<p>Your order #1234 is confirmed</p>"),
    emailing.WithPlainText("Your order #1234 is confirmed"),
    emailing.WithType("order_confirmation"),
)
```

### Template Email
```go
// Send using a template
err := emails.Send("user@example.com",
    emailing.WithTemplate("welcome", map[string]any{
        "UserName": "Alice",
        "SignupDate": time.Now(),
        "Subject": "Welcome to Our App!",
    }),
    emailing.WithType("welcome"),
)
```

### Email with Metadata Tracking
```go
// Send with metadata for analytics
err := emails.Send("user@example.com",
    emailing.WithSubject("Payment Receipt"),
    emailing.WithTemplate("receipt", receiptData),
    emailing.WithType("payment_receipt"),
    emailing.WithMetadata(map[string]string{
        "OrderID": "1234",
        "UserID":  "5678",
        "Amount":  "99.99",
    }),
)
```

## Templates

Templates use Go's standard `html/template` package. Load them from an embedded filesystem:

```go
//go:embed emails/*.html
var emailTemplates embed.FS

emails := emailing.Manage(db,
    emailing.WithTemplateFS(emailTemplates),
)
```

Templates receive whatever data you pass, no wrapping or modification:

```html
<!-- emails/welcome.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{.Subject}}</title>
</head>
<body>
    <h1>Welcome {{.UserName}}!</h1>
    <p>Thanks for signing up on {{.SignupDate.Format "Jan 2, 2006"}}.</p>
</body>
</html>
```

## Providers

The package includes several email providers:

```go
// Resend (recommended)
provider := providers.NewResendProvider(apiKey, nil)

// SendGrid
provider := providers.NewSendGridProvider(apiKey, nil)

// Postmark
provider := providers.NewPostmarkProvider(serverToken, nil)

// Mock (for testing)
provider := providers.NewMockProvider()
```

## Options

Configure the collection with functional options:

```go
emails := emailing.Manage(db,
    // Set sender
    emailing.WithFrom("noreply@example.com", "App Name"),
    
    // Set provider
    emailing.WithProvider(provider),
    
    // Load templates
    emailing.WithTemplateFS(templatesFS, "emails/*.html"),
    
    // Add template functions
    emailing.WithFunc("formatPrice", formatPrice),
    
    // Add controllers for templates (available in all templates)
    emailing.WithController("user", userController),
)

// You can also add functions for a specific send operation:
err := emails.Send("user@example.com",
    emailing.WithTemplate("invoice", invoiceData),
    emailing.WithSendFunc("calculateTax", calculateTax), // Just for this email
)
```

## Database Tracking

Every email is automatically tracked with:
- Recipient, sender, subject
- HTML and plain text content
- Type (welcome, password_reset, etc.)
- Status (pending, sent, delivered, failed)
- Provider used
- Timestamps for sent, delivered, opened, clicked
- Retry count and error messages

Access tracking data:

```go
// Get recent emails
recent, _ := emails.GetRecentEmails(10)

// Get failed emails for retry
failed, _ := emails.GetFailedEmails(maxRetries)

// Get emails by metadata
userEmails, _ := emails.GetEmailsByMetadata("UserID", "12345")

// Get email statistics
stats, _ := emails.GetEmailStats(time.Now().Add(-24*time.Hour))
// Returns map with counts: pending, sent, delivered, failed, total
```

## Testing

Use the mock provider for testing:

```go
mock := providers.NewMockProvider()
emails := emailing.Manage(db,
    emailing.WithProvider(mock),
)

// Send email
emails.Send("test@example.com", "Test message")

// Check it was "sent"
sent := mock.GetSentEmails()
assert.Equal(t, 1, len(sent))
assert.Equal(t, "test@example.com", sent[0].ToAddr)

// Simulate failures
mock.SetFailure(true, errors.New("network error"))
```

## Available Send Options

- `WithSubject(subject string)` - Set the email subject
- `WithHTML(html string)` - Set HTML content
- `WithPlainText(text string)` - Set plain text content
- `WithTemplate(name string, data any)` - Use a template
- `WithType(emailType string)` - Set email type for tracking
- `WithMetadata(metadata map[string]string)` - Add metadata for analytics
- `WithSendFunc(name string, fn any)` - Add a template function for this send only

## Convenience Methods

For common cases, there are also convenience methods:

```go
// SendText - simple text email
err := emails.SendText("user@example.com", "Subject", "Body text")

// SendHTML - HTML with optional text fallback
err := emails.SendHTML("user@example.com", "Subject", htmlContent, textContent)

// SendWithTemplate - template-based email
err := emails.SendWithTemplate("user@example.com", "welcome", data)

// SendMessage - send pre-built message (for compatibility)
err := emails.SendMessage(&emailing.Message{
    ToAddr:      "user@example.com",
    Subject:     "Test",
    HTMLContent: "<p>Test</p>",
    TextContent: "Test",
})
```

## Design Decisions

1. **Functional Options Pattern** - Clean, extensible API using SendOption functions
2. **One Send method** - Simplifies the API and makes the common case easy
3. **Always track** - Audit trail by default, no separate "tracked" version
4. **Any data type** - Templates work with maps, structs, or any type
5. **Filesystem-based templates** - Uses Go's embed for zero-deployment complexity
6. **No template wrapper** - Data is passed as-is to templates
7. **Provider abstraction** - Easy to switch between email services
8. **Graceful failures** - Email sending continues even if tracking fails
9. **Type-safe controllers** - Uses application.Controller interface for template integration

## Package Structure

```
emailing/
├── collection.go   # Core collection type and management
├── send.go        # Send method and convenience helpers
├── options.go     # Functional options for configuration
├── models.go      # Email and metadata models
├── provider.go    # Provider interface definition
└── providers/     # Email provider implementations
    ├── mock.go
    ├── resend.go
    ├── sendgrid.go
    └── postmark.go
```

## Complete Example

```go
package main

import (
    "embed"
    "log"
    
    "github.com/The-Skyscape/devtools/pkg/emailing"
    "github.com/The-Skyscape/devtools/pkg/emailing/providers"
    "github.com/The-Skyscape/devtools/pkg/database/local"
)

//go:embed emails/*.html
var emailTemplates embed.FS

func main() {
    // Setup database
    db := local.Database("app.db")
    
    // Initialize email collection
    emails := emailing.Manage(db,
        emailing.WithFrom("noreply@myapp.com", "My App"),
        emailing.WithProvider(
            providers.NewResendProvider("re_YOUR_KEY", nil),
        ),
        emailing.WithTemplateFS(emailTemplates),
        emailing.WithFunc("formatMoney", formatMoney),
    )
    
    // Send welcome email
    err := emails.Send("user@example.com",
        emailing.WithTemplate("welcome", map[string]any{
            "UserName": "Alice",
            "Subject":  "Welcome to My App!",
        }),
        emailing.WithType("welcome"),
        emailing.WithMetadata(map[string]string{
            "UserID": "12345",
            "Source": "signup",
        }),
    )
    
    if err != nil {
        log.Printf("Failed to send email: %v", err)
    }
}

func formatMoney(cents int) string {
    return fmt.Sprintf("$%.2f", float64(cents)/100)
}
```