package emailing

// Provider is the interface for email providers
type Provider interface {
	Name() string
	Send(email *Email) error
}
