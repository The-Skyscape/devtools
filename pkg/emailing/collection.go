package emailing

import (
	"embed"
	"time"

	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/security"
)

// Collection manages email records, metadata, and sending
type Collection struct {
	db     *database.DynamicDB
	Emails *database.Collection[*Email]

	vault      *security.Collection // Vault for getting email provider credentials
	provider   Provider
	emailFS    embed.FS  // Embedded filesystem with email templates
	emailFSDir string    // Root directory in the filesystem
	fromAddr   string    // Default from address
	fromName   string    // Default from name
}

// Manage creates a new email collection with the given database and options
func Manage(db *database.DynamicDB, opts ...Option) *Collection {
	c := &Collection{
		db:     db,
		Emails: database.Manage(db, new(Email)),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Create indexes for better performance
	db.Query(`
		CREATE INDEX IF NOT EXISTS idx_emails_status ON emails (Status);
		CREATE INDEX IF NOT EXISTS idx_emails_type ON emails (Type);
		CREATE INDEX IF NOT EXISTS idx_emails_messageid ON emails (MessageID);
		CREATE INDEX IF NOT EXISTS idx_emails_to ON emails (ToAddr);
		CREATE INDEX IF NOT EXISTS idx_emails_created ON emails (CreatedAt);
	`).Exec()

	return c
}

// IsConfigured returns true if the collection has a provider configured
func (c *Collection) IsConfigured() bool {
	return c.provider != nil
}

// SetProvider updates the email provider
func (c *Collection) SetProvider(p Provider) error {
	c.provider = p
	if c.vault != nil {
		if err := c.provider.Init(c.vault); err != nil {
			return err
		}
	}
	return nil
}

// SetVault updates the vault
func (c *Collection) SetVault(vault *security.Collection) error {
	c.vault = vault
	if c.provider != nil {
		if err := c.provider.Init(c.vault); err != nil {
			return err
		}
	}
	return nil
}

// LoadTemplates stores the embedded filesystem for lazy template parsing
func (c *Collection) LoadTemplates(emailFS embed.FS) error {
	c.emailFS = emailFS
	c.emailFSDir = "emails"
	return nil
}


// GetEmailByMessageID retrieves an email by provider message ID
func (c *Collection) GetEmailByMessageID(messageID string) (*Email, error) {
	email, err := c.Emails.Find("WHERE MessageID = ?", messageID)
	if err != nil {
		return nil, err
	}
	return email, nil
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
		stats[status] = c.Emails.Count("WHERE Status = ? AND CreatedAt >= ?", status, since)
	}

	// Total emails
	stats["total"] = c.Emails.Count("WHERE CreatedAt >= ?", since)

	return stats, nil
}

