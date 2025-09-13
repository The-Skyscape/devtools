package emailing

import (
	"embed"
	"log"
	
	"github.com/The-Skyscape/devtools/pkg/security"
)

// Option is a functional option for configuring the collection
type Option func(*Collection)

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

// WithTemplates sets the email templates from an embedded filesystem
// DEPRECATED: Use LoadTemplates() method instead for lazy parsing
func WithTemplates(emailFS embed.FS) Option {
	return func(c *Collection) {
		c.emailFS = emailFS
		c.emailFSDir = "emails"
		log.Printf("Email: Templates configured for lazy loading")
	}
}

// WithFrom sets the default from address and name
func WithFrom(addr, name string) Option {
	return func(c *Collection) {
		c.fromAddr = addr
		c.fromName = name
	}
}
