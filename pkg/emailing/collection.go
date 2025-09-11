package emailing

import (
	"html/template"
	"io/fs"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/database"
	"github.com/The-Skyscape/devtools/pkg/security"
)

// Collection manages email records, metadata, and sending
type Collection struct {
	db     *database.DynamicDB
	Emails *database.Collection[*Email]

	vault        *security.Collection // Vault for getting email provider credentials
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
		templates:    template.New(""),
		templateFunc: defaultEmailFuncs(),
		controllers:  make(map[string]application.Controller),
	}

	// Apply template functions
	c.templates = c.templates.Funcs(c.templateFunc)

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
func (c *Collection) SetProvider(p Provider) {
	c.provider = p
}

// GetVault returns the vault if configured
func (c *Collection) GetVault() *security.Collection {
	return c.vault
}

// LoadTemplates loads templates from a filesystem
func (c *Collection) LoadTemplates(fsys fs.FS, patterns ...string) {
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
