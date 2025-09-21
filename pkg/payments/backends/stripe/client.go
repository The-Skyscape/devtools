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

// Client is a Stripe API client implementing the payments.Backend interface
type Client struct {
	secretKey     string
	publishKey    string
	webhookSecret string
	baseURL       string
	client        *http.Client
	apiVersion    string
	products      map[string]*StripeProduct // Product cache by name
}

// NewClient creates a new Stripe client with configured products
func NewClient(secretKey, publishKey, webhookSecret string, opts ...payments.BackendOption) *Client {
	c := &Client{
		secretKey:     secretKey,
		publishKey:    publishKey,
		webhookSecret: webhookSecret,
		baseURL:       "https://api.stripe.com/v1",
		apiVersion:    "2023-10-16",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		products: make(map[string]*StripeProduct),
	}

	// Apply options to build configuration
	cfg := &payments.BackendConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Initialize products from configuration
	if err := c.initializeProducts(cfg); err != nil {
		// Log error but don't fail - products can be synced later
		fmt.Printf("Warning: Failed to initialize products: %v\n", err)
	}

	return c
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


// listProducts lists all products with optional filters (internal use only)
func (c *Client) listProducts(limit int) ([]*Product, error) {
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

// StripeProduct implements the payments.Product interface
type StripeProduct struct {
	id          string
	name        string
	description string
	priceID     string
	amount      int64
	currency    string
	interval    string
	metadata    map[string]string
}

func (p *StripeProduct) ID() string                 { return p.id }
func (p *StripeProduct) Name() string               { return p.name }
func (p *StripeProduct) Description() string        { return p.description }
func (p *StripeProduct) PriceID() string            { return p.priceID }
func (p *StripeProduct) Amount() int64              { return p.amount }
func (p *StripeProduct) Currency() string           { return p.currency }
func (p *StripeProduct) Interval() string           { return p.interval }
func (p *StripeProduct) Metadata() map[string]string { return p.metadata }

// StripeCustomer implements the payments.Customer interface
type StripeCustomer struct {
	id    string
	email string
	name  string
}

func (c *StripeCustomer) ID() string    { return c.id }
func (c *StripeCustomer) Email() string { return c.email }
func (c *StripeCustomer) Name() string  { return c.name }

// StripeCheckout implements the payments.Checkout interface
type StripeCheckout struct {
	id             string
	url            string
	status         string
	customerID     string
	subscriptionID string
	metadata       map[string]string
}

func (c *StripeCheckout) ID() string                 { return c.id }
func (c *StripeCheckout) URL() string                { return c.url }
func (c *StripeCheckout) Status() string             { return c.status }
func (c *StripeCheckout) CustomerID() string         { return c.customerID }
func (c *StripeCheckout) SubscriptionID() string     { return c.subscriptionID }
func (c *StripeCheckout) Metadata() map[string]string { return c.metadata }

// StripeSubscription implements the payments.Subscription interface
type StripeSubscription struct {
	id               string
	customerID       string
	productID        string
	priceID          string
	status           string
	currentPeriodEnd int64
}

func (s *StripeSubscription) ID() string               { return s.id }
func (s *StripeSubscription) CustomerID() string       { return s.customerID }
func (s *StripeSubscription) ProductID() string        { return s.productID }
func (s *StripeSubscription) PriceID() string          { return s.priceID }
func (s *StripeSubscription) Status() string           { return s.status }
func (s *StripeSubscription) CurrentPeriodEnd() int64  { return s.currentPeriodEnd }

// StripeEvent implements the payments.Event interface
type StripeEvent struct {
	id   string
	typ  string
	data map[string]interface{}
}

func (e *StripeEvent) ID() string                   { return e.id }
func (e *StripeEvent) Type() string                 { return e.typ }
func (e *StripeEvent) Data() map[string]interface{} { return e.data }

// initializeProducts ensures products exist in Stripe based on configuration
func (c *Client) initializeProducts(cfg *payments.BackendConfig) error {
	for _, product := range cfg.Products {
		// Check if product exists
		existingProducts, err := c.listProducts(100)
		if err != nil {
			return fmt.Errorf("failed to list products: %w", err)
		}

		var stripeProduct *Product
		for _, p := range existingProducts {
			if p.Name == product.Name {
				stripeProduct = p
				break
			}
		}

		// Create product if it doesn't exist
		if stripeProduct == nil {
			stripeProduct, err = c.CreateProduct(product.Name, product.Description, product.Metadata)
			if err != nil {
				return fmt.Errorf("failed to create product %s: %w", product.Name, err)
			}
		}

		// Check for existing price
		prices, err := c.ListPrices(stripeProduct.ID)
		if err != nil {
			return fmt.Errorf("failed to list prices for %s: %w", product.Name, err)
		}

		var stripePrice *Price
		for _, p := range prices {
			if p.UnitAmount == product.Price &&
				p.Currency == product.Currency &&
				p.Recurring != nil &&
				p.Recurring.Interval == product.Interval {
				stripePrice = p
				break
			}
		}

		// Create price if it doesn't exist
		if stripePrice == nil {
			stripePrice, err = c.CreatePrice(stripeProduct.ID, product.Price, product.Currency, true, product.Interval)
			if err != nil {
				return fmt.Errorf("failed to create price for %s: %w", product.Name, err)
			}
		}

		// Cache the product
		c.products[product.Name] = &StripeProduct{
			id:          stripeProduct.ID,
			name:        stripeProduct.Name,
			description: stripeProduct.Description,
			priceID:     stripePrice.ID,
			amount:      stripePrice.UnitAmount,
			currency:    stripePrice.Currency,
			interval:    product.Interval,
			metadata:    stripeProduct.Metadata,
		}
	}

	return nil
}

// Product retrieves a product by ID (implements Backend interface)
func (c *Client) Product(id string) (payments.Product, error) {
	// Check cache first
	for _, p := range c.products {
		if p.id == id {
			return p, nil
		}
	}

	// Fetch from Stripe
	stripeProduct, err := c.getStripeProduct(id)
	if err != nil {
		return nil, err
	}

	// Get default price
	prices, err := c.ListPrices(id)
	if err != nil {
		return nil, err
	}
	if len(prices) == 0 {
		return nil, fmt.Errorf("product %s has no prices", id)
	}

	return &StripeProduct{
		id:          stripeProduct.ID,
		name:        stripeProduct.Name,
		description: stripeProduct.Description,
		priceID:     prices[0].ID,
		amount:      prices[0].UnitAmount,
		currency:    prices[0].Currency,
		interval:    prices[0].Recurring.Interval,
		metadata:    stripeProduct.Metadata,
	}, nil
}

// ProductByName retrieves a product by name (implements Backend interface)
func (c *Client) ProductByName(name string) (payments.Product, error) {
	if product, ok := c.products[name]; ok {
		return product, nil
	}
	return nil, fmt.Errorf("product %s not found", name)
}

// Products returns all configured products (implements Backend interface)
func (c *Client) Products() ([]payments.Product, error) {
	products := make([]payments.Product, 0, len(c.products))
	for _, p := range c.products {
		products = append(products, p)
	}
	return products, nil
}

// getStripeProduct is a helper to get the raw Stripe product
func (c *Client) getStripeProduct(productID string) (*Product, error) {
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
