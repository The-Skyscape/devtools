package emailing

import (
	"github.com/The-Skyscape/devtools/pkg/security"
)

// Option is a functional option for configuring the collection
type Option func(*Collection)

// Collection configuration options

// WithVault sets the security vault for the collection
func WithVault(v *security.Collection) Option {
	return func(c *Collection) {
		c.SetVault(v)
	}
}

// WithProvider sets the email provider (overrides vault configuration)
func WithProvider(p Provider) Option {
	return func(c *Collection) {
		c.SetProvider(p)
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
