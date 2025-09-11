package emailing

// Provider is the interface for email providers
type Provider interface {
	Send(msg *Message) error
	GetName() string
}

// Message represents an email message
type Message struct {
	ToAddr      string
	FromAddr    string
	FromName    string
	Subject     string
	HTMLContent string
	TextContent string
	ReplyTo     string
	Tags        []string
}

// Config holds email service configuration
type Config struct {
	Provider string // "resend", "sendgrid" or "postmark"
	APIKey   string
	From     string
	FromName string
}
