package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/emailing"
	"github.com/The-Skyscape/devtools/pkg/security"
)

// ResendProvider implements the Provider interface for Resend
type ResendProvider struct {
	ApiKey   string
	FromAddr string
	FromName string
}

func (p *ResendProvider) Init(vault *security.Collection) error {
	if vault == nil {
		return nil
	}
	
	secret, err := vault.GetSecret("integrations/email/resend")
	if err != nil {
		return err
	}
	
	apiKey, ok := secret["api_key"].(string)
	if !ok || apiKey == "" {
		return nil
	}
	
	// Update provider configuration
	p.ApiKey = apiKey
	
	// Get from address from vault or keep existing
	if addr, ok := secret["from_address"].(string); ok && addr != "" {
		p.FromAddr = addr
	}
	if name, ok := secret["from_name"].(string); ok && name != "" {
		p.FromName = name
	}
	
	return nil
}

func (p *ResendProvider) Name() string {
	return "resend"
}

// NewResendProvider creates a new Resend email provider
func NewResendProvider(apiKey, fromAddr, fromName string) *ResendProvider {
	return &ResendProvider{
		ApiKey:   apiKey,
		FromAddr: fromAddr,
		FromName: fromName,
	}
}

// Send sends an email via Resend
func (p *ResendProvider) Send(e *emailing.Email) error {
	uri := "https://api.resend.com/emails"

	// Use provider's configured from address
	fromStr := p.FromAddr
	if p.FromName != "" {
		fromStr = fmt.Sprintf("%s <%s>", p.FromName, p.FromAddr)
	}

	// Build Resend request as JSON
	payload := map[string]any{
		"from":    fromStr,
		"to":      e.ToAddr,
		"subject": e.Subject,
	}

	// Add HTML content if provided
	if e.Body != "" {
		payload["html"] = e.Body
	}

	// Add text content if provided
	if e.PlainText != "" {
		payload["text"] = e.PlainText
	}

	// Add reply-to if provided
	if e.ReplyTo != "" {
		payload["reply_to"] = e.ReplyTo
	}

	// Encode to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Resend request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
