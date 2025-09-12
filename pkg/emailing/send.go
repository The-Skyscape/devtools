package emailing

import (
	"cmp"
	"fmt"
)

// Send sends an email with the provided options
func (c *Collection) Send(to string, opts ...SendOption) error {
	// Create email with defaults
	email := &Email{
		ToAddr:   to,
		Type:     "notification",
		Status:   "pending",
		Provider: c.provider.Name(),
	}

	// Apply options
	for _, opt := range opts {
		opt(email)
	}

	// Apply defaults
	email.Subject = cmp.Or(email.Subject, "Notification")

	// Save to database (don't fail if tracking fails)
	savedEmail, _ := c.Emails.Insert(email)

	// Send the email if provider configured
	var err error
	if c.provider != nil {
		err = c.provider.Send(email)
	} else {
		err = fmt.Errorf("email provider not configured")
	}

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

// Convenience methods for common cases

// SendText sends a simple text email
func (c *Collection) SendText(to, subject, text string) error {
	return c.Send(to,
		WithSubject(subject),
		WithPlainText(text),
	)
}

// SendHTML sends an HTML email with optional text fallback
func (c *Collection) SendHTML(to, subject, html, text string) error {
	return c.Send(to,
		WithSubject(subject),
		WithHTML(html),
		WithPlainText(text),
	)
}

