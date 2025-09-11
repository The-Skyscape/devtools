package emailing

import (
	"cmp"
	"fmt"
	"time"

	"github.com/The-Skyscape/devtools/pkg/database"
)

// Collection manages email records, metadata, and sending
type Collection struct {
	db       *database.DynamicDB
	Emails   *database.Collection[*Email]
	Metadata *database.Collection[*EmailMetadata]

	// Email sending configuration
	provider  Provider
	fromAddr  string
	fromName  string
	templates *TemplateEngine
}

// Option is a functional option for configuring the collection
type Option func(*Collection) error

// Manage creates a new email collection with the given database and options
func Manage(db *database.DynamicDB, opts ...Option) *Collection {
	c := &Collection{
		db:        db,
		Emails:    database.Manage(db, new(Email)),
		Metadata:  database.Manage(db, new(EmailMetadata)),
		templates: NewTemplateEngine(),
		fromAddr:  "noreply@example.com",
		fromName:  "Application",
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			// Log error but continue - collection can work without provider
			fmt.Printf("Warning: Failed to apply option: %v\n", err)
		}
	}

	// Create indexes for better performance
	db.Query(`
		CREATE INDEX IF NOT EXISTS idx_emails_status ON emails (Status);
		CREATE INDEX IF NOT EXISTS idx_emails_type ON emails (Type);
		CREATE INDEX IF NOT EXISTS idx_emails_messageid ON emails (MessageID);
		CREATE INDEX IF NOT EXISTS idx_email_metadata_emailid ON email_metadata (EmailID);
		CREATE INDEX IF NOT EXISTS idx_email_metadata_key ON email_metadata (Key);
	`).Exec()

	return c
}

// WithProvider sets the email provider
func WithProvider(p Provider) Option {
	return func(c *Collection) error {
		c.provider = p // nil is OK, means no provider
		return nil
	}
}

// WithFrom sets the default from address
func WithFrom(addr, name string) Option {
	return func(c *Collection) error {
		c.fromAddr = addr
		c.fromName = name
		return nil
	}
}

// WithTemplates sets the template engine
func WithTemplates(engine *TemplateEngine) Option {
	return func(c *Collection) error {
		if engine != nil {
			c.templates = engine
		}
		return nil
	}
}

// IsConfigured returns true if the collection has a provider configured
func (c *Collection) IsConfigured() bool {
	return c.provider != nil
}

// Send sends a simple email message
func (c *Collection) Send(to, subject, htmlContent, textContent string) error {
	if c.provider == nil {
		return fmt.Errorf("email provider not configured")
	}

	msg := &Message{
		ToAddr:      to,
		FromAddr:    c.fromAddr,
		FromName:    c.fromName,
		Subject:     subject,
		HTMLContent: htmlContent,
		TextContent: textContent,
	}

	return c.SendMessage(msg)
}

// SendMessage sends a pre-built message
func (c *Collection) SendMessage(msg *Message) error {
	if c.provider == nil {
		return fmt.Errorf("email provider not configured")
	}

	// Set defaults if not specified
	msg.FromAddr = cmp.Or(msg.FromAddr, c.fromAddr)
	msg.FromName = cmp.Or(msg.FromName, c.fromName)

	return c.provider.Send(msg)
}

// SendAndTrack sends an email and tracks it in the database
func (c *Collection) SendAndTrack(to, subject, htmlContent, textContent, emailType string, metadata map[string]string) error {
	// Create tracking record
	email := &Email{
		ToAddr:    to,
		FromAddr:  c.fromAddr,
		Subject:   subject,
		Body:      htmlContent,
		PlainText: textContent,
		Type:      emailType,
		Status:    "pending",
		Provider:  c.GetProviderName(),
	}

	// Save to database
	savedEmail, err := c.Emails.Insert(email)
	if err != nil {
		// Log but don't fail - sending is more important than tracking
		fmt.Printf("Failed to create email tracking record: %v\n", err)
		savedEmail = nil
	}

	// Add metadata if provided
	if savedEmail != nil && metadata != nil {
		for key, value := range metadata {
			c.AddEmailMetadata(savedEmail.ID, key, value, "string")
		}
	}

	// Send the email
	err = c.Send(to, subject, htmlContent, textContent)

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

// SendTemplate sends an email using a template
func (c *Collection) SendTemplate(templateName, to string, data interface{}) error {
	if c.templates == nil {
		return fmt.Errorf("template engine not configured")
	}

	// Prepare template data
	td, ok := data.(*TemplateData)
	if !ok {
		// Wrap raw data in TemplateData
		td = NewTemplateData()
		td.To = to
		td.From = c.fromAddr
		td.FromName = c.fromName
		td.Data["Content"] = data
	}

	// Ensure recipient is set
	td.To = cmp.Or(td.To, to)

	// Render the template
	htmlContent, err := c.templates.Render(templateName, td)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Create and send the message
	msg := &Message{
		ToAddr:      td.To,
		FromAddr:    td.From,
		FromName:    td.FromName,
		Subject:     td.Subject,
		HTMLContent: htmlContent,
		TextContent: "", // Could generate plain text from HTML
	}

	return c.SendMessage(msg)
}

// GetProviderName returns the name of the configured provider
func (c *Collection) GetProviderName() string {
	if c.provider == nil {
		return ""
	}
	return c.provider.GetName()
}

// SetProvider updates the email provider
func (c *Collection) SetProvider(p Provider) {
	c.provider = p
}

// SetFrom updates the default from address
func (c *Collection) SetFrom(email, name string) {
	c.fromAddr = email
	c.fromName = name
}

// GetEmailByMessageID retrieves an email by provider message ID
func (c *Collection) GetEmailByMessageID(messageID string) (*Email, error) {
	emails, err := c.Emails.Search("WHERE MessageID = ?", messageID)
	if err != nil {
		return nil, err
	}
	if len(emails) == 0 {
		return nil, nil
	}
	return emails[0], nil
}

// GetEmailMetadata retrieves all metadata for an email
func (c *Collection) GetEmailMetadata(emailID string) ([]*EmailMetadata, error) {
	return c.Metadata.Search("WHERE EmailID = ?", emailID)
}

// AddEmailMetadata adds metadata to an email
func (c *Collection) AddEmailMetadata(emailID, key, value, valueType string) (*EmailMetadata, error) {
	metadata := &EmailMetadata{
		EmailID:   emailID,
		Key:       key,
		Value:     value,
		ValueType: valueType,
	}
	return c.Metadata.Insert(metadata)
}

// GetEmailsByMetadata retrieves emails with specific metadata
func (c *Collection) GetEmailsByMetadata(key, value string) ([]*Email, error) {
	// First get metadata entries
	metadataEntries, err := c.Metadata.Search("WHERE Key = ? AND Value = ?", key, value)
	if err != nil {
		return nil, err
	}

	// Collect email IDs and fetch them
	var emails []*Email
	for _, m := range metadataEntries {
		email, err := c.Emails.Get(m.EmailID)
		if err == nil {
			emails = append(emails, email)
		}
	}

	return emails, nil
}

// GetRecentEmails retrieves recent emails
func (c *Collection) GetRecentEmails(limit int) ([]*Email, error) {
	return c.Emails.Search("ORDER BY CreatedAt DESC LIMIT ?", limit)
}

// GetFailedEmails retrieves failed emails that can be retried
func (c *Collection) GetFailedEmails(maxRetries int) ([]*Email, error) {
	return c.Emails.Search("WHERE Status = ? AND RetryCount < ?", "failed", maxRetries)
}

// GetPendingEmails retrieves all pending emails
func (c *Collection) GetPendingEmails() ([]*Email, error) {
	return c.Emails.Search("WHERE Status = ?", "pending")
}

// GetEmailStats returns email statistics for a time period
func (c *Collection) GetEmailStats(since time.Time) (map[string]int, error) {
	stats := make(map[string]int)

	// Get counts by status
	statuses := []string{"pending", "sent", "delivered", "opened", "clicked", "bounced", "failed"}
	for _, status := range statuses {
		emails, err := c.Emails.Search("WHERE Status = ? AND CreatedAt >= ?", status, since)
		if err != nil {
			return nil, err
		}
		stats[status] = len(emails)
	}

	// Total emails
	allEmails, err := c.Emails.Search("WHERE CreatedAt >= ?", since)
	if err != nil {
		return nil, err
	}
	stats["total"] = len(allEmails)

	return stats, nil
}
