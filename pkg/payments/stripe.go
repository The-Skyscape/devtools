package payments

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
)

// StripeClient provides a native Go implementation of Stripe API integration
type StripeClient struct {
	secretKey      string
	publishableKey string
	webhookSecret  string
	baseURL        string
	client         *http.Client
}

// StripeConfig holds configuration for Stripe
type StripeConfig struct {
	SecretKey      string
	PublishableKey string
	WebhookSecret  string
	TestMode       bool
}

// NewStripeClient creates a new Stripe client
func NewStripeClient(config *StripeConfig) *StripeClient {
	baseURL := "https://api.stripe.com"

	return &StripeClient{
		secretKey:      config.SecretKey,
		publishableKey: config.PublishableKey,
		webhookSecret:  config.WebhookSecret,
		baseURL:        baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes an authenticated request to the Stripe API
func (s *StripeClient) makeRequest(method, endpoint string, data url.Values) (map[string]interface{}, error) {
	url := s.baseURL + endpoint

	var req *http.Request
	var err error

	if data != nil && (method == "POST" || method == "PUT") {
		req, err = http.NewRequest(method, url, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		if data != nil {
			req.URL.RawQuery = data.Encode()
		}
	}

	// Set authentication header
	req.SetBasicAuth(s.secretKey, "")
	req.Header.Set("Stripe-Version", "2023-10-16")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if resp.StatusCode >= 400 {
		if errMsg, ok := result["error"].(map[string]interface{}); ok {
			return nil, fmt.Errorf("stripe error: %v", errMsg["message"])
		}
		return nil, fmt.Errorf("stripe returned status %d: %s", resp.StatusCode, string(body))
	}

	return result, nil
}

// CreateCheckoutSession creates a new Stripe Checkout session
func (s *StripeClient) CreateCheckoutSession(customerEmail, priceID, successURL, cancelURL string) (string, error) {
	data := url.Values{}
	data.Set("customer_email", customerEmail)
	data.Set("mode", "subscription")
	data.Set("success_url", successURL)
	data.Set("cancel_url", cancelURL)
	data.Set("line_items[0][price]", priceID)
	data.Set("line_items[0][quantity]", "1")

	// Add metadata
	data.Set("metadata[platform]", "skyscape")
	data.Set("metadata[timestamp]", fmt.Sprintf("%d", time.Now().Unix()))

	result, err := s.makeRequest("POST", "/v1/checkout/sessions", data)
	if err != nil {
		return "", err
	}

	if url, ok := result["url"].(string); ok {
		return url, nil
	}

	return "", fmt.Errorf("no URL in checkout session response")
}

// CreateCustomer creates a new customer
func (s *StripeClient) CreateCustomer(email, name, description string) (string, error) {
	data := url.Values{}
	data.Set("email", email)
	if name != "" {
		data.Set("name", name)
	}
	if description != "" {
		data.Set("description", description)
	}

	result, err := s.makeRequest("POST", "/v1/customers", data)
	if err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("no customer ID in response")
}

// GetCustomer retrieves a customer by ID
func (s *StripeClient) GetCustomer(customerID string) (map[string]interface{}, error) {
	return s.makeRequest("GET", "/v1/customers/"+customerID, nil)
}

// CreateSubscription creates a new subscription for a customer
func (s *StripeClient) CreateSubscription(customerID, priceID string) (string, error) {
	data := url.Values{}
	data.Set("customer", customerID)
	data.Set("items[0][price]", priceID)

	result, err := s.makeRequest("POST", "/v1/subscriptions", data)
	if err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("no subscription ID in response")
}

// CancelSubscription cancels a subscription
func (s *StripeClient) CancelSubscription(subscriptionID string, immediately bool) error {
	data := url.Values{}
	if immediately {
		data.Set("invoice_now", "true")
		data.Set("prorate", "true")
	}

	_, err := s.makeRequest("DELETE", "/v1/subscriptions/"+subscriptionID, data)
	return err
}

// GetSubscription retrieves subscription details
func (s *StripeClient) GetSubscription(subscriptionID string) (map[string]interface{}, error) {
	return s.makeRequest("GET", "/v1/subscriptions/"+subscriptionID, nil)
}

// CreatePrice creates a new price object
func (s *StripeClient) CreatePrice(productID string, amount int64, currency, interval string) (string, error) {
	data := url.Values{}
	data.Set("product", productID)
	data.Set("unit_amount", fmt.Sprintf("%d", amount))
	data.Set("currency", currency)
	data.Set("recurring[interval]", interval)

	result, err := s.makeRequest("POST", "/v1/prices", data)
	if err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("no price ID in response")
}

// CreateProduct creates a new product
func (s *StripeClient) CreateProduct(name, description string) (string, error) {
	data := url.Values{}
	data.Set("name", name)
	if description != "" {
		data.Set("description", description)
	}

	result, err := s.makeRequest("POST", "/v1/products", data)
	if err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("no product ID in response")
}

// VerifyWebhookSignature verifies a webhook signature from Stripe
func (s *StripeClient) VerifyWebhookSignature(payload, signature string) bool {
	// Parse the signature header
	elements := strings.Split(signature, ",")
	var timestamp string
	var signatures []string

	for _, element := range elements {
		parts := strings.Split(element, "=")
		if len(parts) != 2 {
			continue
		}

		switch parts[0] {
		case "t":
			timestamp = parts[1]
		case "v1":
			signatures = append(signatures, parts[1])
		}
	}

	if timestamp == "" || len(signatures) == 0 {
		return false
	}

	// Check timestamp is recent (within 5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix()-ts > 300 {
		return false // Too old
	}

	// Compute expected signature
	signedPayload := timestamp + "." + payload
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Check if any signature matches
	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expectedSig)) {
			return true
		}
	}

	return false
}

// ParseWebhookEvent parses a webhook event from Stripe
func (s *StripeClient) ParseWebhookEvent(payload string) (map[string]interface{}, error) {
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}
	return event, nil
}

// CreatePaymentIntent creates a new payment intent
func (s *StripeClient) CreatePaymentIntent(amount int64, currency, customerID string) (string, error) {
	data := url.Values{}
	data.Set("amount", fmt.Sprintf("%d", amount))
	data.Set("currency", currency)
	if customerID != "" {
		data.Set("customer", customerID)
	}

	result, err := s.makeRequest("POST", "/v1/payment_intents", data)
	if err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("no payment intent ID in response")
}

// GetPublishableKey returns the publishable key for client-side use
func (s *StripeClient) GetPublishableKey() string {
	return s.publishableKey
}
