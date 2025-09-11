package emailing

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/database"
)

// Collection manages email records, metadata, and sending
type Collection struct {
	db       *database.DynamicDB
	Emails   *database.Collection[*Email]
	Metadata *database.Collection[*EmailMetadata]

	// Email sending configuration
	provider     Provider
	templates    *template.Template
	templateFunc template.FuncMap
	controllers  map[string]application.Controller // Controllers available to templates
}

// Manage creates a new email collection with the given database and options
func Manage(db *database.DynamicDB, opts ...Option) *Collection {
	c := &Collection{
		db:           db,
		Emails:       database.Manage(db, new(Email)),
		Metadata:     database.Manage(db, new(EmailMetadata)),
		templates:    template.New(""),
		templateFunc: defaultEmailFuncs(),
		controllers:  make(map[string]application.Controller),
	}

	// Apply template functions
	c.templates = c.templates.Funcs(c.templateFunc)

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

// IsConfigured returns true if the collection has a provider configured
func (c *Collection) IsConfigured() bool {
	return c.provider != nil
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

// defaultEmailFuncs returns the default template functions for emails
func defaultEmailFuncs() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"title":   strings.Title,
		"trim":    strings.TrimSpace,
		"replace": strings.ReplaceAll,

		// Utility functions
		"default": func(def, val any) any {
			if val == nil || val == "" {
				return def
			}
			return val
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
	}
}
