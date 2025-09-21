package stripe

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/The-Skyscape/devtools/pkg/payments"
)

// CreateCustomer creates a new customer in Stripe (implements Backend interface)
func (c *Client) CreateCustomer(email, name string, opts ...payments.CustomerOption) (payments.Customer, error) {
	// Build config from options
	cfg := payments.BuildCustomerConfig(opts...)

	formParams := url.Values{}
	formParams.Set("email", email)
	formParams.Set("name", name)

	if cfg.Phone != "" {
		formParams.Set("phone", cfg.Phone)
	}

	// Add metadata
	formatMetadata(formParams, cfg.Metadata)

	// Make request
	resp, err := c.request("POST", "/customers", formParams)
	if err != nil {
		return nil, err
	}

	// Parse response
	var stripeCustomer struct {
		ID       string            `json:"id"`
		Email    string            `json:"email"`
		Name     string            `json:"name"`
		Metadata map[string]string `json:"metadata"`
	}

	if err := json.Unmarshal(resp, &stripeCustomer); err != nil {
		return nil, fmt.Errorf("failed to parse customer: %w", err)
	}

	return &StripeCustomer{
		id:    stripeCustomer.ID,
		email: stripeCustomer.Email,
		name:  stripeCustomer.Name,
	}, nil
}


// Customer retrieves a customer by ID (implements Backend interface)
func (c *Client) Customer(customerID string) (payments.Customer, error) {
	resp, err := c.request("GET", "/customers/"+customerID, nil)
	if err != nil {
		return nil, err
	}

	// Parse response
	var stripeCustomer struct {
		ID       string            `json:"id"`
		Email    string            `json:"email"`
		Name     string            `json:"name"`
		Metadata map[string]string `json:"metadata"`
	}

	if err := json.Unmarshal(resp, &stripeCustomer); err != nil {
		return nil, fmt.Errorf("failed to parse customer: %w", err)
	}

	return &StripeCustomer{
		id:    stripeCustomer.ID,
		email: stripeCustomer.Email,
		name:  stripeCustomer.Name,
	}, nil
}

// Customers retrieves a list of customers (implements Backend interface)
func (c *Client) Customers(limit int) ([]payments.Customer, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "100") // Default limit
	}

	resp, err := c.request("GET", "/customers", params)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response struct {
		Data []struct {
			ID       string            `json:"id"`
			Email    string            `json:"email"`
			Name     string            `json:"name"`
			Metadata map[string]string `json:"metadata"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse customers list: %w", err)
	}

	customers := make([]payments.Customer, len(response.Data))
	for i, cust := range response.Data {
		customers[i] = &StripeCustomer{
			id:    cust.ID,
			email: cust.Email,
			name:  cust.Name,
		}
	}

	return customers, nil
}

// AttachPaymentMethod attaches a payment method to a customer
func (c *Client) AttachPaymentMethod(customerID, paymentMethodID string) error {
	params := url.Values{}
	params.Set("customer", customerID)
	
	_, err := c.request("POST", "/payment_methods/"+paymentMethodID+"/attach", params)
	return err
}

// DetachPaymentMethod detaches a payment method
func (c *Client) DetachPaymentMethod(paymentMethodID string) error {
	_, err := c.request("POST", "/payment_methods/"+paymentMethodID+"/detach", nil)
	return err
}

// SetDefaultPaymentMethod sets the default payment method for a customer
func (c *Client) SetDefaultPaymentMethod(customerID, paymentMethodID string) error {
	params := url.Values{}
	params.Set("invoice_settings[default_payment_method]", paymentMethodID)
	
	_, err := c.request("POST", "/customers/"+customerID, params)
	return err
}

