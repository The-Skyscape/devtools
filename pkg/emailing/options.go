package emailing

import (
	"html/template"
	"io/fs"

	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/security"
)

// Option is a functional option for configuring the collection
type Option func(*Collection)

// Collection configuration options

// WithVault sets the security vault for the collection
func WithVault(vault *security.Collection) Option {
	return func(c *Collection) {
		c.vault = vault
		// Provider should be set separately with WithProvider
	}
}

// WithProvider sets the email provider (overrides vault configuration)
func WithProvider(p Provider) Option {
	return func(c *Collection) {
		c.provider = p // nil is OK, means no provider
	}
}

// WithTemplates sets the templates directly
func WithTemplates(tmpl *template.Template) Option {
	return func(c *Collection) {
		if tmpl != nil {
			c.templates = tmpl.Funcs(c.templateFunc)
		}
	}
}

// WithFunc adds a template function to the email templates
func WithFunc(name string, fn any) Option {
	return func(c *Collection) {
		c.templateFunc[name] = fn
		c.templates = c.templates.Funcs(c.templateFunc)
	}
}

// WithController adds a controller to make it available in email templates
func WithController(name string, controller application.Controller) Option {
	return func(c *Collection) {
		c.controllers[name] = controller
		// Also add to template functions for direct access
		WithFunc(name, func() application.Controller { return controller })(c)
	}
}

// WithTemplateFS loads templates from a filesystem
func WithTemplateFS(fsys fs.FS, patterns ...string) Option {
	return func(c *Collection) {
		// If no patterns specified, use common email patterns
		if len(patterns) == 0 {
			patterns = []string{"emails/*.html", "emails/**/*.html"}
		}

		for _, pattern := range patterns {
			tmpl, err := c.templates.Funcs(c.templateFunc).ParseFS(fsys, pattern)
			if err != nil {
				// Log but don't fail if pattern doesn't match
				continue
			}
			c.templates = tmpl
		}
	}
}

// Send operation options

// SendOption configures an email send operation
type SendOption func(*Email)

// WithSubject sets the email subject
func WithSubject(subject string) SendOption {
	return func(e *Email) {
		e.Subject = subject
	}
}

// WithHTML sets the HTML content
func WithHTML(html string) SendOption {
	return func(e *Email) {
		e.Body = html
	}
}

// WithPlainText sets the plain text content
func WithPlainText(text string) SendOption {
	return func(e *Email) {
		e.PlainText = text
	}
}

// WithType sets the email type for tracking
func WithType(emailType string) SendOption {
	return func(e *Email) {
		e.Type = emailType
	}
}

// WithReplyTo sets the reply-to address
func WithReplyTo(replyTo string) SendOption {
	return func(e *Email) {
		e.ReplyTo = replyTo
	}
}
