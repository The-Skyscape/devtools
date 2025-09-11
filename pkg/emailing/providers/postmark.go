package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/emailing"
)

// PostmarkProvider implements the Provider interface for Postmark
type PostmarkProvider struct {
	apiKey string
	client *http.Client
}

// NewPostmarkProvider creates a new Postmark email provider
func NewPostmarkProvider(apiKey string, client *http.Client) *PostmarkProvider {
	return &PostmarkProvider{
		apiKey: apiKey,
		client: client,
	}
}

// Send sends an email via Postmark
func (p *PostmarkProvider) Send(msg *emailing.Message) error {
	url := "https://api.postmarkapp.com/email"

	// Build Postmark request
	payload := map[string]any{
		"From":     fmt.Sprintf("%s <%s>", msg.FromName, msg.FromAddr),
		"To":       msg.ToAddr,
		"Subject":  msg.Subject,
		"HtmlBody": msg.HTMLContent,
		"TextBody": msg.TextContent,
	}

	if msg.ReplyTo != "" {
		payload["ReplyTo"] = msg.ReplyTo
	}

	if len(msg.Tags) > 0 {
		payload["Tag"] = msg.Tags[0] // Postmark only supports one tag
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Postmark request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create Postmark request: %w", err)
	}

	req.Header.Set("X-Postmark-Server-Token", p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email via Postmark: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Postmark returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetName returns the provider name
func (p *PostmarkProvider) GetName() string {
	return "postmark"
}
