package stripe

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/The-Skyscape/devtools/pkg/payments"
)

// CreateCheckout creates a new checkout session (implements Backend interface)
func (c *Client) CreateCheckout(productName string, opts ...payments.CheckoutOption) (payments.Checkout, error) {
	// Build config from options
	cfg := payments.BuildCheckoutConfig(opts...)

	// Look up product by name
	product, ok := c.products[productName]
	if !ok {
		return nil, fmt.Errorf("product %s not configured", productName)
	}

	formParams := url.Values{}
	formParams.Set("mode", "subscription") // Default to subscription mode

	// Set URLs
	if cfg.SuccessURL != "" {
		formParams.Set("success_url", cfg.SuccessURL)
	}
	if cfg.CancelURL != "" {
		formParams.Set("cancel_url", cfg.CancelURL)
	}

	// Add line item for the product
	formParams.Set("line_items[0][price]", product.priceID)
	formParams.Set("line_items[0][quantity]", strconv.Itoa(cfg.Quantity))

	// Add customer
	if cfg.CustomerID != "" {
		formParams.Set("customer", cfg.CustomerID)
	} else if cfg.CustomerEmail != "" {
		formParams.Set("customer_email", cfg.CustomerEmail)
	}

	// Add trial period
	if cfg.TrialDays > 0 {
		formParams.Set("subscription_data[trial_period_days]", strconv.Itoa(cfg.TrialDays))
	}

	// Allow promo codes
	if cfg.AllowPromoCodes {
		formParams.Set("allow_promotion_codes", "true")
	}

	// Add metadata
	formatMetadata(formParams, cfg.Metadata)

	// Payment method collection
	formParams.Set("payment_method_collection", "if_required")

	// Make request
	resp, err := c.request("POST", "/checkout/sessions", formParams)
	if err != nil {
		return nil, err
	}

	// Parse response
	var stripeSession struct {
		ID            string            `json:"id"`
		URL           string            `json:"url"`
		Status        string            `json:"status"`
		Customer      string            `json:"customer"`
		Subscription  string            `json:"subscription"`
		Metadata      map[string]string `json:"metadata"`
	}

	if err := json.Unmarshal(resp, &stripeSession); err != nil {
		return nil, fmt.Errorf("failed to parse checkout session: %w", err)
	}

	return &StripeCheckout{
		id:             stripeSession.ID,
		url:            stripeSession.URL,
		status:         stripeSession.Status,
		customerID:     stripeSession.Customer,
		subscriptionID: stripeSession.Subscription,
		metadata:       stripeSession.Metadata,
	}, nil
}

// Checkout retrieves a checkout session (implements Backend interface)
func (c *Client) Checkout(id string) (payments.Checkout, error) {
	resp, err := c.request("GET", "/checkout/sessions/"+id, nil)
	if err != nil {
		return nil, err
	}

	// Parse response
	var stripeSession struct {
		ID            string            `json:"id"`
		URL           string            `json:"url"`
		Status        string            `json:"status"`
		Customer      string            `json:"customer"`
		Subscription  string            `json:"subscription"`
		Metadata      map[string]string `json:"metadata"`
	}

	if err := json.Unmarshal(resp, &stripeSession); err != nil {
		return nil, fmt.Errorf("failed to parse checkout session: %w", err)
	}

	return &StripeCheckout{
		id:             stripeSession.ID,
		url:            stripeSession.URL,
		status:         stripeSession.Status,
		customerID:     stripeSession.Customer,
		subscriptionID: stripeSession.Subscription,
		metadata:       stripeSession.Metadata,
	}, nil
}