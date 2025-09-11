package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/payments"
)

// Client is a Stripe API client implementing the payments.Provider interface
type Client struct {
	secretKey     string
	publishKey    string
	webhookSecret string
	baseURL       string
	client        *http.Client
	apiVersion    string
}

// NewClient creates a new Stripe client
func NewClient(secretKey, publishKey, webhookSecret string) *Client {
	return &Client{
		secretKey:     secretKey,
		publishKey:    publishKey,
		webhookSecret: webhookSecret,
		baseURL:       "https://api.stripe.com/v1",
		apiVersion:    "2023-10-16",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the provider name
func (c *Client) GetName() string {
	return "stripe"
}

// PublishableKey returns the publishable key for client-side use
func (c *Client) PublishableKey() string {
	return c.publishKey
}

// Request makes an authenticated request to the Stripe API
func (c *Client) request(method, endpoint string, params url.Values) ([]byte, error) {
	url := c.baseURL + endpoint

	var body io.Reader
	if params != nil && len(params) > 0 {
		body = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Use basic auth with secret key
	req.SetBasicAuth(c.secretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Stripe-Version", c.apiVersion)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var stripeErr struct {
			Error struct {
				Type    string `json:"type"`
				Code    string `json:"code"`
				Message string `json:"message"`
				Param   string `json:"param"`
			} `json:"error"`
		}
		
		if err := json.Unmarshal(responseBody, &stripeErr); err == nil {
			return nil, &payments.Error{
				Type:    stripeErr.Error.Type,
				Code:    stripeErr.Error.Code,
				Message: stripeErr.Error.Message,
				Param:   stripeErr.Error.Param,
			}
		}
		
		return nil, fmt.Errorf("stripe API error: status %d", resp.StatusCode)
	}

	return responseBody, nil
}

// VerifyWebhookSignature verifies a Stripe webhook signature
func (c *Client) VerifyWebhookSignature(payload []byte, sigHeader string) bool {
	// Parse the signature header
	// Format: t=timestamp,v1=signature
	parts := strings.Split(sigHeader, ",")
	var timestamp string
	var signature string

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			if kv[0] == "t" {
				timestamp = kv[1]
			} else if kv[0] == "v1" {
				signature = kv[1]
			}
		}
	}

	if timestamp == "" || signature == "" {
		return false
	}

	// Check timestamp is recent (within 5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > 300 {
		return false
	}

	// Compute expected signature
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// ParseWebhookEvent parses a webhook event from Stripe
func (c *Client) ParseWebhookEvent(payload []byte) (*payments.WebhookEvent, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	event := &payments.WebhookEvent{
		Data: make(map[string]interface{}),
	}

	// Extract standard fields
	if id, ok := raw["id"].(string); ok {
		event.ID = id
	}
	if eventType, ok := raw["type"].(string); ok {
		event.Type = c.normalizeEventType(eventType)
	}
	if created, ok := raw["created"].(float64); ok {
		event.Created = time.Unix(int64(created), 0)
	}

	// Extract data object
	if data, ok := raw["data"].(map[string]interface{}); ok {
		event.Data = data
		if object, ok := data["object"]; ok {
			event.Object = object
		}
	}

	return event, nil
}

// normalizeEventType converts Stripe event types to generic types
func (c *Client) normalizeEventType(stripeType string) string {
	// Map Stripe-specific event types to generic ones
	switch stripeType {
	case "checkout.session.completed":
		return payments.EventTypeCheckoutComplete
	case "customer.created":
		return payments.EventTypeCustomerCreated
	case "customer.updated":
		return payments.EventTypeCustomerUpdated
	case "customer.deleted":
		return payments.EventTypeCustomerDeleted
	case "customer.subscription.created":
		return payments.EventTypeSubscriptionCreated
	case "customer.subscription.updated":
		return payments.EventTypeSubscriptionUpdated
	case "customer.subscription.deleted":
		return payments.EventTypeSubscriptionDeleted
	case "payment_intent.succeeded":
		return payments.EventTypePaymentIntentSucceeded
	case "payment_intent.payment_failed":
		return payments.EventTypePaymentIntentFailed
	case "payment_method.attached":
		return payments.EventTypePaymentMethodAttached
	case "payment_method.detached":
		return payments.EventTypePaymentMethodDetached
	default:
		return stripeType // Return as-is if not mapped
	}
}

// parseJSON is a helper to parse JSON responses
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// formatMetadata adds metadata to form parameters
func formatMetadata(params url.Values, metadata map[string]string) {
	for key, value := range metadata {
		params.Set("metadata["+key+"]", value)
	}
}

// parseTime parses a Unix timestamp to time.Time
func parseTime(timestamp float64) time.Time {
	return time.Unix(int64(timestamp), 0)
}

// parseTimePtr parses a Unix timestamp to *time.Time
func parseTimePtr(timestamp float64) *time.Time {
	if timestamp == 0 {
		return nil
	}
	t := parseTime(timestamp)
	return &t
}