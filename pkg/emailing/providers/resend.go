package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/The-Skyscape/devtools/pkg/emailing"
)

// ResendProvider implements the Provider interface for Resend
type ResendProvider struct {
	apiKey string
	client *http.Client
}

// NewResendProvider creates a new Resend email provider
func NewResendProvider(apiKey string, client *http.Client) *ResendProvider {
	return &ResendProvider{
		apiKey: apiKey,
		client: client,
	}
}

// Send sends an email via Resend
func (p *ResendProvider) Send(msg *emailing.Message) error {
	url := "https://api.resend.com/emails"

	// Use provider's configured from address
	fromStr := p.fromAddr
	if p.fromName != "" {
		fromStr = fmt.Sprintf("%s <%s>", p.fromName, p.fromAddr)
	}
	
	// Build Resend request
	payload := map[string]any{
		"from":    fromStr,
		"to":      email.ToAddr,
		"subject": email.Subject,
	}

	// Add HTML content if provided
	if email.Body != "" {
		payload["html"] = email.Body
	}

	// Add text content if provided
	if email.PlainText != "" {
		payload["text"] = email.PlainText
	}

	// Add reply-to if provided
	if email.ReplyTo != "" {
		payload["reply_to"] = email.ReplyTo
	}

	// Add tags if provided
	if len(msg.Tags) > 0 {
		payload["tags"] = msg.Tags
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Resend request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create Resend request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendBatch sends multiple emails via Resend's batch endpoint
func (p *ResendProvider) SendBatch(messages []*emailing.Message) error {
	url := "https://api.resend.com/emails/batch"

	// Build batch request
	batch := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		email := map[string]any{
			"from":    fmt.Sprintf("%s <%s>", msg.FromName, msg.FromAddr),
			"to":      msg.ToAddr,
			"subject": msg.Subject,
		}

		if msg.HTMLContent != "" {
			email["html"] = msg.HTMLContent
		}
		if msg.TextContent != "" {
			email["text"] = msg.TextContent
		}
		if msg.ReplyTo != "" {
			email["reply_to"] = msg.ReplyTo
		}
		if len(msg.Tags) > 0 {
			email["tags"] = msg.Tags
		}

		batch = append(batch, email)
	}

	body, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal Resend batch request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create Resend batch request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send batch email via Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend batch returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetName returns the provider name
func (p *ResendProvider) GetName() string {
	return "resend"
}

// ValidateAPIKey checks if the API key has the correct format
func ValidateResendAPIKey(apiKey string) bool {
	// Resend API keys start with "re_"
	return strings.HasPrefix(apiKey, "re_")
}
