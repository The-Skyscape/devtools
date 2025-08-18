package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Provider is the interface for email providers
type Provider interface {
	Send(msg *Message) error
	GetName() string
}

// Message represents an email message
type Message struct {
	To          string
	From        string
	FromName    string
	Subject     string
	HTMLContent string
	TextContent string
	ReplyTo     string
	Tags        []string
}

// Config holds email service configuration
type Config struct {
	Provider string // "resend", "sendgrid" or "postmark"
	APIKey   string
	From     string
	FromName string
}

// Service handles email sending
type Service struct {
	config   *Config
	provider Provider
	client   *http.Client
}

// NewService creates a new email service
func NewService(config *Config) (*Service, error) {
	s := &Service{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	
	// Initialize the provider
	switch config.Provider {
	case "resend":
		s.provider = NewResendProvider(config.APIKey, s.client)
	case "sendgrid":
		s.provider = &SendGridProvider{
			apiKey: config.APIKey,
			client: s.client,
		}
	case "postmark":
		s.provider = &PostmarkProvider{
			apiKey: config.APIKey,
			client: s.client,
		}
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", config.Provider)
	}
	
	return s, nil
}

// Send sends an email message
func (s *Service) Send(msg *Message) error {
	// Set default from if not specified
	if msg.From == "" {
		msg.From = s.config.From
	}
	if msg.FromName == "" {
		msg.FromName = s.config.FromName
	}
	
	return s.provider.Send(msg)
}

// SendGridProvider implements the Provider interface for SendGrid
type SendGridProvider struct {
	apiKey string
	client *http.Client
}

// Send sends an email via SendGrid
func (p *SendGridProvider) Send(msg *Message) error {
	url := "https://api.sendgrid.com/v3/mail/send"
	
	// Build SendGrid request
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": []map[string]string{
					{"email": msg.To},
				},
			},
		},
		"from": map[string]string{
			"email": msg.From,
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

// PostmarkProvider implements the Provider interface for Postmark
type PostmarkProvider struct {
	apiKey string
	client *http.Client
}

// Send sends an email via Postmark
func (p *PostmarkProvider) Send(msg *Message) error {
	url := "https://api.postmarkapp.com/email"
	
	// Build Postmark request
	payload := map[string]interface{}{
		"From":     fmt.Sprintf("%s <%s>", msg.FromName, msg.From),
		"To":       msg.To,
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