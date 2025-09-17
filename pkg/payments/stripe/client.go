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
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	event := &payments.WebhookEvent{
		Data: make(map[string]any),
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
	if data, ok := raw["data"].(map[string]any); ok {
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
func parseJSON(data []byte, v any) error {
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

// Product represents a Stripe product
type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Active      bool              `json:"active"`
	Metadata    map[string]string `json:"metadata"`
	Created     int64             `json:"created"`
	Updated     int64             `json:"updated"`
}

// Price represents a Stripe price
type Price struct {
	ID         string            `json:"id"`
	ProductID  string            `json:"product"`
	Active     bool              `json:"active"`
	Currency   string            `json:"currency"`
	UnitAmount int64             `json:"unit_amount"`
	Type       string            `json:"type"`
	Recurring  *PriceRecurring   `json:"recurring"`
	Metadata   map[string]string `json:"metadata"`
	Created    int64             `json:"created"`
}

// PriceRecurring represents recurring price settings
type PriceRecurring struct {
	Interval      string `json:"interval"`       // day, week, month, year
	IntervalCount int    `json:"interval_count"`
}

// CreateProduct creates a new product in Stripe
func (c *Client) CreateProduct(name, description string, metadata map[string]string) (*Product, error) {
	params := url.Values{}
	params.Set("name", name)
	if description != "" {
		params.Set("description", description)
	}
	formatMetadata(params, metadata)

	data, err := c.request("POST", "/products", params)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	var product Product
	if err := parseJSON(data, &product); err != nil {
		return nil, fmt.Errorf("failed to parse product response: %w", err)
	}

	return &product, nil
}

// GetProduct retrieves a product by ID
func (c *Client) GetProduct(productID string) (*Product, error) {
	data, err := c.request("GET", "/products/"+productID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	var product Product
	if err := parseJSON(data, &product); err != nil {
		return nil, fmt.Errorf("failed to parse product response: %w", err)
	}

	return &product, nil
}

// ListProducts lists all products with optional filters
func (c *Client) ListProducts(limit int) ([]*Product, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	data, err := c.request("GET", "/products", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	var response struct {
		Data []*Product `json:"data"`
	}
	if err := parseJSON(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse products response: %w", err)
	}

	return response.Data, nil
}

// UpdateProduct updates an existing product
func (c *Client) UpdateProduct(productID string, name, description string, metadata map[string]string) (*Product, error) {
	params := url.Values{}
	if name != "" {
		params.Set("name", name)
	}
	if description != "" {
		params.Set("description", description)
	}
	formatMetadata(params, metadata)

	data, err := c.request("POST", "/products/"+productID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	var product Product
	if err := parseJSON(data, &product); err != nil {
		return nil, fmt.Errorf("failed to parse product response: %w", err)
	}

	return &product, nil
}

// CreatePrice creates a new price for a product
func (c *Client) CreatePrice(productID string, unitAmount int64, currency string, recurring bool, interval string) (*Price, error) {
	params := url.Values{}
	params.Set("product", productID)
	params.Set("unit_amount", strconv.FormatInt(unitAmount, 10))
	params.Set("currency", currency)

	if recurring {
		params.Set("recurring[interval]", interval)
		params.Set("recurring[interval_count]", "1")
	}

	data, err := c.request("POST", "/prices", params)
	if err != nil {
		return nil, fmt.Errorf("failed to create price: %w", err)
	}

	var price Price
	if err := parseJSON(data, &price); err != nil {
		return nil, fmt.Errorf("failed to parse price response: %w", err)
	}

	return &price, nil
}

// GetPrice retrieves a price by ID
func (c *Client) GetPrice(priceID string) (*Price, error) {
	data, err := c.request("GET", "/prices/"+priceID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	var price Price
	if err := parseJSON(data, &price); err != nil {
		return nil, fmt.Errorf("failed to parse price response: %w", err)
	}

	return &price, nil
}

// ListPrices lists prices for a product
func (c *Client) ListPrices(productID string) ([]*Price, error) {
	params := url.Values{}
	params.Set("product", productID)
	params.Set("active", "true")

	data, err := c.request("GET", "/prices", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list prices: %w", err)
	}

	var response struct {
		Data []*Price `json:"data"`
	}
	if err := parseJSON(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse prices response: %w", err)
	}

	return response.Data, nil
}
