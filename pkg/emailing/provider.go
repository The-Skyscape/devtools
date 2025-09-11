package emailing

// Provider is the interface for email providers
type Provider interface {
	Send(email *Email) error
	GetName() string
}

// Config holds email service configuration
type Config struct {
	Provider string // "resend", "sendgrid" or "postmark"
	APIKey   string
	From     string
	FromName string
}
