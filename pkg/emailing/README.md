# Emailing Package

A flexible email sending package with template support and multiple provider backends.

## Features

- **Multiple Providers**: Support for Resend, SendGrid, and Postmark
- **Embedded Templates**: Use Go's `embed` package for email templates
- **Template Engine**: Built-in template engine with useful functions
- **Message Builder**: Fluent interface for building emails
- **Type Safety**: Strongly typed configuration and messages

## Installation

```go
import "github.com/The-Skyscape/devtools/pkg/emailing"
```

## Quick Start

```go
// Embed your email templates
//go:embed templates/*.html
var emailTemplates embed.FS

// Create service
service, err := emailing.NewService(
    emailing.WithResend("your-api-key"),
    emailing.WithFrom("noreply@example.com", "My App"),
    emailing.WithTemplates(emailTemplates, "templates"),
)

// Send email
err = service.Send(
    "user@example.com",
    "Subject",
    "<h1>HTML Content</h1>",
    "Text Content",
)

// Send with template
data := emailing.NewTemplateData()
data.Subject = "Welcome"
data.UserName = "John"
err = service.SendTemplate("welcome", "user@example.com", data)
```

## Providers

### Resend
```go
emailing.WithResend("re_YOUR_API_KEY")
```

### SendGrid
```go
emailing.WithSendGrid("SG.YOUR_API_KEY")
```

### Postmark
```go
emailing.WithPostmark("YOUR_SERVER_TOKEN")
```

## Templates

Create HTML templates with embedded partials:

```html
<!-- templates/welcome.html -->
<!DOCTYPE html>
<html>
<body>
    {{template "email-header" .}}
    <h1>Welcome {{.UserName}}!</h1>
    {{template "email-footer" .}}
</body>
</html>
```

## Template Functions

Built-in template functions:
- String manipulation: `upper`, `lower`, `title`, `trim`, `replace`
- Safety: `safeHTML`, `safeURL`, `safeJS`, `safeCSS`
- Utilities: `default`, `contains`, `hasPrefix`, `hasSuffix`

## Message Builder

```go
service.BuildMessage().
    To("user@example.com").
    Subject("Hello").
    HTML("<h1>Hello World</h1>").
    ReplyTo("support@example.com").
    Tag("notification").
    Send()
```

## Custom Template Functions

```go
service.AddTemplateFunc("formatDate", func(t time.Time) string {
    return t.Format("Jan 2, 2006")
})
```