package providers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/The-Skyscape/devtools/pkg/emailing"
)

// ResendProvider implements the Provider interface for Resend
type ResendProvider struct {
	ApiKey   string
	FromAddr string
	FromName string
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

	// Build Resend request
	payload := url.Values{}
	payload.Add("from", fromStr)
	payload.Add("to", e.ToAddr)
	payload.Add("subject", e.Subject)

	// Add HTML content if provided
	if e.Body != "" {
		payload.Add("html", e.Body)
	}

	// Add text content if provided
	if e.PlainText != "" {
		payload.Add("text", e.PlainText)
	}

	// Add reply-to if provided
	if e.ReplyTo != "" {
		payload.Add("reply_to", e.ReplyTo)
	}

	req, err := http.NewRequest("POST", uri, strings.NewReader(payload.Encode()))
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
