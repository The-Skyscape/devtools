package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/emailing"
)

// SendGridProvider implements the Provider interface for SendGrid
type SendGridProvider struct {
	apiKey string
	client *http.Client
}

// NewSendGridProvider creates a new SendGrid email provider
func NewSendGridProvider(apiKey string, client *http.Client) *SendGridProvider {
	return &SendGridProvider{
		apiKey: apiKey,
		client: client,
	}
}

// Send sends an email via SendGrid
func (p *SendGridProvider) Send(msg *emailing.Message) error {
	url := "https://api.sendgrid.com/v3/mail/send"

	// Build SendGrid request
	payload := map[string]any{
		"personalizations": []map[string]any{
			{
				"to": []map[string]string{
					{"email": msg.ToAddr},
				},
			},
		},
		"from": map[string]string{
			"email": msg.FromAddr,
			"name":  msg.FromName,
		},
		"subject": msg.Subject,
		"content": []map[string]string{
			{
				"type":  "text/html",
				"value": msg.HTMLContent,
			},
			{
				"type":  "text/plain",
				"value": msg.TextContent,
			},
		},
	}

	if msg.ReplyTo != "" {
		payload["reply_to"] = map[string]string{"email": msg.ReplyTo}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SendGrid request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create SendGrid request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email via SendGrid: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SendGrid returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetName returns the provider name
func (p *SendGridProvider) GetName() string {
	return "sendgrid"
}
