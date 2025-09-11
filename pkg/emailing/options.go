package emailing

import (
	"html/template"
	"io/fs"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// Option is a functional option for configuring the collection
type Option func(*Collection) error

// Collection configuration options

// WithProvider sets the email provider
func WithProvider(p Provider) Option {
	return func(c *Collection) error {
		c.provider = p // nil is OK, means no provider
		return nil
	}
}

// WithTemplates sets the templates directly
func WithTemplates(tmpl *template.Template) Option {
	return func(c *Collection) error {
		if tmpl != nil {
			c.templates = tmpl.Funcs(c.templateFunc)
		}
		return nil
	}
}

// WithFunc adds a template function to the email templates
func WithFunc(name string, fn any) Option {
	return func(c *Collection) error {
		c.templateFunc[name] = fn
		c.templates = c.templates.Funcs(c.templateFunc)
		return nil
	}
}

// WithController adds a controller to make it available in email templates
func WithController(name string, controller application.Controller) Option {
	return func(c *Collection) error {
		c.controllers[name] = controller
		// Also add to template functions for direct access
		return WithFunc(name, func() application.Controller { return controller })(c)
	}
}

// WithTemplateFS loads templates from a filesystem
func WithTemplateFS(fsys fs.FS, patterns ...string) Option {
	return func(c *Collection) error {
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
		return nil
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