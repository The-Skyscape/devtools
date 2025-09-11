package emailing

import (
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// Email represents a tracked email in the system
type Email struct {
	application.Model

	// Email details
	ToAddr    string // Recipient email address
	FromAddr  string // Sender email address
	FromName  string // Sender name
	Subject   string
	Body      string // HTML body
	PlainText string // Plain text version
	ReplyTo   string // Reply-to address (optional)

	// Email metadata
	Type     string // welcome, password_reset, notification, etc.
	Status   string // pending, sent, delivered, bounced, failed
	Provider string // resend, sendgrid, etc.

	// Tracking information
	MessageID   string
	SentAt      time.Time
	DeliveredAt time.Time
	OpenedAt    time.Time
	ClickedAt   time.Time
	BouncedAt   time.Time

	// Error tracking
	ErrorMessage string
	RetryCount   int
}

func (e *Email) Table() string {
	return "emails"
}

// MarkAsSent updates the email status to sent
func (e *Email) MarkAsSent(messageID string) {
	e.Status = "sent"
	e.MessageID = messageID
	e.SentAt = time.Now()
}

// MarkAsDelivered updates the email status to delivered
func (e *Email) MarkAsDelivered() {
	e.Status = "delivered"
	e.DeliveredAt = time.Now()
}

// MarkAsBounced updates the email status to bounced
func (e *Email) MarkAsBounced(reason string) {
	e.Status = "bounced"
	e.BouncedAt = time.Now()
	e.ErrorMessage = reason
}

// MarkAsFailed updates the email status to failed
func (e *Email) MarkAsFailed(reason string) {
	e.Status = "failed"
	e.ErrorMessage = reason
	e.RetryCount++
}

// MarkAsOpened updates the email status to opened
func (e *Email) MarkAsOpened() {
	if e.OpenedAt.IsZero() {
		e.OpenedAt = time.Now()
	}

	// Only update status if it's not already clicked (clicked is higher priority)
	if e.Status != "clicked" {
		e.Status = "opened"
	}
}

// MarkAsClicked updates the email status to clicked
func (e *Email) MarkAsClicked() {
	if e.ClickedAt.IsZero() {
		e.ClickedAt = time.Now()
	}

	// Also mark as opened if not already
	if e.OpenedAt.IsZero() {
		e.OpenedAt = time.Now()
	}

	e.Status = "clicked"
}
