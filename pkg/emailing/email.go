package emailing

import (
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// Email represents a tracked email in the system
type Email struct {
	application.Model
	
	// Email details
	ToAddr      string    // Recipient email address
	FromAddr    string    // Sender email address
	Subject     string
	Body        string    // HTML body
	PlainText   string    // Plain text version
	
	// Email metadata
	Type        string    // welcome, password_reset, notification, etc.
	Status      string    // pending, sent, delivered, bounced, failed
	Provider    string    // resend, sendgrid, etc.
	
	// Tracking information
	MessageID   string    // Provider's message ID
	SentAt      *time.Time
	DeliveredAt *time.Time
	OpenedAt    *time.Time
	ClickedAt   *time.Time
	BouncedAt   *time.Time
	
	// Error tracking
	ErrorMessage string
	RetryCount   int
}

func (e *Email) Table() string {
	return "emails"
}

// EmailMetadata represents app-specific metadata for an email
// Applications can use this to link emails to their domain objects
type EmailMetadata struct {
	application.Model
	
	EmailID    string    // Reference to the Email
	Key        string    // Metadata key (e.g., "UserID", "WorkspaceID", "OrderID")
	Value      string    // Metadata value
	ValueType  string    // Type hint (e.g., "string", "int", "uuid")
}

func (m *EmailMetadata) Table() string {
	return "email_metadata"
}

// Metadata retrieves all metadata for this email
// This method requires passing the collection to access the database
func (e *Email) Metadata(c *Collection) ([]*EmailMetadata, error) {
	return c.GetEmailMetadata(e.ID)
}

// GetMetadataValue retrieves a specific metadata value by key
func (e *Email) GetMetadataValue(c *Collection, key string) (string, error) {
	metadata, err := c.GetEmailMetadata(e.ID)
	if err != nil {
		return "", err
	}
	
	for _, m := range metadata {
		if m.Key == key {
			return m.Value, nil
		}
	}
	
	return "", nil
}

// AddMetadata adds a metadata entry for an email
func (e *Email) AddMetadata(key, value, valueType string) *EmailMetadata {
	return &EmailMetadata{
		EmailID:   e.ID,
		Key:       key,
		Value:     value,
		ValueType: valueType,
	}
}

// MarkAsSent updates the email status to sent
func (e *Email) MarkAsSent(messageID string) error {
	now := time.Now()
	e.Status = "sent"
	e.MessageID = messageID
	e.SentAt = &now
	return nil // Update will be handled by the collection
}

// MarkAsDelivered updates the email status to delivered
func (e *Email) MarkAsDelivered() error {
	now := time.Now()
	e.Status = "delivered"
	e.DeliveredAt = &now
	return nil // Update will be handled by the collection
}

// MarkAsBounced updates the email status to bounced
func (e *Email) MarkAsBounced(reason string) error {
	now := time.Now()
	e.Status = "bounced"
	e.BouncedAt = &now
	e.ErrorMessage = reason
	return nil // Update will be handled by the collection
}

// MarkAsFailed updates the email status to failed
func (e *Email) MarkAsFailed(reason string) error {
	e.Status = "failed"
	e.ErrorMessage = reason
	e.RetryCount++
	return nil // Update will be handled by the collection
}

// MarkAsOpened updates the email status to opened
func (e *Email) MarkAsOpened() error {
	now := time.Now()
	if e.OpenedAt == nil {
		e.OpenedAt = &now
	}
	// Only update status if it's not already clicked (clicked is higher priority)
	if e.Status != "clicked" {
		e.Status = "opened"
	}
	return nil // Update will be handled by the collection
}

// MarkAsClicked updates the email status to clicked
func (e *Email) MarkAsClicked() error {
	now := time.Now()
	if e.ClickedAt == nil {
		e.ClickedAt = &now
	}
	// Also mark as opened if not already
	if e.OpenedAt == nil {
		e.OpenedAt = &now
	}
	e.Status = "clicked"
	return nil // Update will be handled by the collection
}