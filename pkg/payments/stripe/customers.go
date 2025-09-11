package stripe

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/The-Skyscape/devtools/pkg/payments"
)

// CreateCustomer creates a new customer in Stripe
func (c *Client) CreateCustomer(params *payments.CustomerParams) (*payments.Customer, error) {
	formParams := url.Values{}
	
	if params.Email != "" {
		formParams.Set("email", params.Email)
	}
	if params.Name != "" {
		formParams.Set("name", params.Name)
	}
	if params.Phone != "" {
		formParams.Set("phone", params.Phone)
	}
	if params.Description != "" {
		formParams.Set("description", params.Description)
	}
	
	// Add address if provided
	if params.Address != nil {
		if params.Address.Line1 != "" {
			formParams.Set("address[line1]", params.Address.Line1)
		}
		if params.Address.Line2 != "" {
			formParams.Set("address[line2]", params.Address.Line2)
		}
		if params.Address.City != "" {
			formParams.Set("address[city]", params.Address.City)
		}
		if params.Address.State != "" {
			formParams.Set("address[state]", params.Address.State)
		}
		if params.Address.PostalCode != "" {
			formParams.Set("address[postal_code]", params.Address.PostalCode)
		}
		if params.Address.Country != "" {
			formParams.Set("address[country]", params.Address.Country)
		}
	}
	
	// Add tax ID if provided
	if params.TaxID != "" {
		formParams.Set("tax_id_data[0][type]", "us_ein") // Default to US EIN, could be parameterized
		formParams.Set("tax_id_data[0][value]", params.TaxID)
	}
	
	// Add metadata
	formatMetadata(formParams, params.Metadata)
	
	// Make request
	resp, err := c.request("POST", "/customers", formParams)
	if err != nil {
		return nil, err
	}
	
	return c.parseCustomer(resp)
}

// GetCustomer retrieves a customer by ID
func (c *Client) GetCustomer(customerID string) (*payments.Customer, error) {
	resp, err := c.request("GET", "/customers/"+customerID, nil)
	if err != nil {
		return nil, err
	}
	
	return c.parseCustomer(resp)
}

// UpdateCustomer updates a customer's information
func (c *Client) UpdateCustomer(customerID string, params *payments.CustomerParams) (*payments.Customer, error) {
	formParams := url.Values{}
	
	if params.Email != "" {
		formParams.Set("email", params.Email)
	}
	if params.Name != "" {
		formParams.Set("name", params.Name)
	}
	if params.Phone != "" {
		formParams.Set("phone", params.Phone)
	}
	if params.Description != "" {
		formParams.Set("description", params.Description)
	}
	
	// Update address if provided
	if params.Address != nil {
		if params.Address.Line1 != "" {
			formParams.Set("address[line1]", params.Address.Line1)
		}
		if params.Address.Line2 != "" {
			formParams.Set("address[line2]", params.Address.Line2)
		}
		if params.Address.City != "" {
			formParams.Set("address[city]", params.Address.City)
		}
		if params.Address.State != "" {
			formParams.Set("address[state]", params.Address.State)
		}
		if params.Address.PostalCode != "" {
			formParams.Set("address[postal_code]", params.Address.PostalCode)
		}
		if params.Address.Country != "" {
			formParams.Set("address[country]", params.Address.Country)
		}
	}
	
	// Update metadata
	if len(params.Metadata) > 0 {
		formatMetadata(formParams, params.Metadata)
	}
	
	// Make request
	resp, err := c.request("POST", "/customers/"+customerID, formParams)
	if err != nil {
		return nil, err
	}
	
	return c.parseCustomer(resp)
}

// DeleteCustomer deletes a customer
func (c *Client) DeleteCustomer(customerID string) error {
	_, err := c.request("DELETE", "/customers/"+customerID, nil)
	return err
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

// parseCustomer parses a Stripe customer response
func (c *Client) parseCustomer(data []byte) (*payments.Customer, error) {
	var stripeCustomer struct {
		ID          string            `json:"id"`
		Email       string            `json:"email"`
		Name        string            `json:"name"`
		Phone       string            `json:"phone"`
		Description string            `json:"description"`
		Metadata    map[string]string `json:"metadata"`
		Created     int64             `json:"created"`
		Balance     int64             `json:"balance"`
		Currency    string            `json:"currency"`
		Delinquent  bool              `json:"delinquent"`
		InvoiceSettings struct {
			DefaultPaymentMethod string `json:"default_payment_method"`
		} `json:"invoice_settings"`
	}
	
	if err := json.Unmarshal(data, &stripeCustomer); err != nil {
		return nil, fmt.Errorf("failed to parse customer: %w", err)
	}
	
	return &payments.Customer{
		ID:                   stripeCustomer.ID,
		Email:                stripeCustomer.Email,
		Name:                 stripeCustomer.Name,
		Phone:                stripeCustomer.Phone,
		Description:          stripeCustomer.Description,
		DefaultPaymentMethod: stripeCustomer.InvoiceSettings.DefaultPaymentMethod,
		Metadata:             stripeCustomer.Metadata,
		Created:              parseTime(float64(stripeCustomer.Created)),
		Updated:              parseTime(float64(stripeCustomer.Created)), // Stripe doesn't have updated field
		Balance:              stripeCustomer.Balance,
		Currency:             stripeCustomer.Currency,
		Delinquent:           stripeCustomer.Delinquent,
	}, nil
}