package emailing

import (
	"cmp"
	"fmt"
	"html/template"
	"net/http"
)

// emailContext holds the context for building an email
type emailContext struct {
	email    *Email
	request  *http.Request
	template string
	funcs    template.FuncMap
}

// SendOption configures an email before sending
type SendOption func(*emailContext)

// Send sends an email with the provided options
func (c *Collection) Send(to, subject string, opts ...SendOption) error {
	// Create context with defaults
	ctx := &emailContext{
		email: &Email{
			ToAddr:  to,
			Subject: subject,
			Type:    "notification",
			Status:  "pending",
		},
		funcs: make(template.FuncMap),
	}
	
	// Set provider name if available
	if c.provider != nil {
		ctx.email.Provider = c.provider.Name()
	}
	
	// Apply default from address if configured
	if c.fromAddr != "" {
		ctx.email.FromAddr = c.fromAddr
		ctx.email.FromName = c.fromName
	}

	// Apply options
	for _, opt := range opts {
		opt(ctx)
	}
	
	// If template specified, render it
	if ctx.template != "" {
		html, text, err := c.renderTemplate(ctx)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", ctx.template, err)
		}
		ctx.email.Body = html
		ctx.email.PlainText = cmp.Or(ctx.email.PlainText, text)
	}

	// Apply final defaults
	ctx.email.Subject = cmp.Or(ctx.email.Subject, "Notification")

	// Save to database (don't fail if tracking fails)
	savedEmail, _ := c.Emails.Insert(ctx.email)

	// Send the email if provider configured
	var err error
	if c.provider != nil {
		err = c.provider.Send(ctx.email)
	} else {
		// Log but don't error if no provider
		// This allows development without email configuration
		err = nil
	}

	// Update tracking record
	if savedEmail != nil {
		if err != nil {
			savedEmail.MarkAsFailed(err.Error())
		} else {
			savedEmail.MarkAsSent("")
		}
		c.Emails.Update(savedEmail)
	}

	return err
}

// WithTemplate specifies an email template to use (by filename with extension)
func WithTemplate(filename string) SendOption {
	return func(ctx *emailContext) {
		ctx.template = filename
	}
}

// WithRequest injects the HTTP request for template access
func WithRequest(r *http.Request) SendOption {
	return func(ctx *emailContext) {
		ctx.request = r
		// Register request as a function like application package does
		ctx.funcs["req"] = func() *http.Request { return r }
	}
}

// WithData registers a function that returns the given value
// This follows the application package pattern of using functions instead of maps
func WithData(key string, value any) SendOption {
	return func(ctx *emailContext) {
		// Create a closure that captures the value
		ctx.funcs[key] = func() any { return value }
	}
}

// WithText sets the plain text content
func WithText(text string) SendOption {
	return func(ctx *emailContext) {
		ctx.email.PlainText = text
	}
}

// WithHTML sets the HTML content directly (no template)
func WithHTML(html string) SendOption {
	return func(ctx *emailContext) {
		ctx.email.Body = html
	}
}

// WithFunc registers a custom template function for dynamic or lazy evaluation
// The function will be called when the template references it
func WithFunc(name string, fn any) SendOption {
	return func(ctx *emailContext) {
		ctx.funcs[name] = fn
	}
}

// WithType sets the email type for tracking
func WithType(emailType string) SendOption {
	return func(ctx *emailContext) {
		ctx.email.Type = emailType
	}
}

// WithReplyTo sets the reply-to address
func WithReplyTo(replyTo string) SendOption {
	return func(ctx *emailContext) {
		ctx.email.ReplyTo = replyTo
	}
}

// WithFromOverride overrides the default from address for this email
func WithFromOverride(addr, name string) SendOption {
	return func(ctx *emailContext) {
		ctx.email.FromAddr = addr
		ctx.email.FromName = name
	}
}

// Convenience methods for common cases

// SendText sends a simple text email
func (c *Collection) SendText(to, subject, text string) error {
	return c.Send(to, subject,
		WithText(text),
	)
}

// SendHTML sends an HTML email with optional text fallback
func (c *Collection) SendHTML(to, subject, html, text string) error {
	return c.Send(to, subject,
		WithHTML(html),
		WithText(text),
	)
}

