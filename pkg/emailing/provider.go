package emailing

import "github.com/The-Skyscape/devtools/pkg/security"

// Provider is the interface for email providers
type Provider interface {
	Init(vault *security.Collection) error // Initialize from vault secrets
	Name() string
	Send(email *Email) error
}
