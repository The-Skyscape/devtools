package stripe

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/The-Skyscape/devtools/pkg/payments"
)

// CreateCheckoutSession creates a new Stripe checkout session
func (c *Client) CreateCheckoutSession(params *payments.CheckoutParams) (*payments.CheckoutSession, error) {
	formParams := url.Values{}
	formParams.Set("mode", params.Mode)
	formParams.Set("success_url", params.SuccessURL)
	formParams.Set("cancel_url", params.CancelURL)

	// Add line items
	for i, item := range params.LineItems {
		prefix := fmt.Sprintf("line_items[%d]", i)
		
		if item.PriceID != "" {
			// Using existing price
			formParams.Set(prefix+"[price]", item.PriceID)
			formParams.Set(prefix+"[quantity]", strconv.Itoa(item.Quantity))
		} else {
			// Creating price data inline (for one-time payments)
			formParams.Set(prefix+"[price_data][currency]", item.Currency)
			formParams.Set(prefix+"[price_data][unit_amount]", strconv.FormatInt(item.Amount, 10))
			formParams.Set(prefix+"[price_data][product_data][name]", item.Description)
			formParams.Set(prefix+"[quantity]", strconv.Itoa(item.Quantity))
		}
	}

	// Add metadata
	formatMetadata(formParams, params.Metadata)

	// Add customer if provided
	if params.CustomerID != "" {
		formParams.Set("customer", params.CustomerID)
	} else if params.CustomerEmail != "" {
		formParams.Set("customer_email", params.CustomerEmail)
	}

	// Add trial period for subscriptions
	if params.Mode == "subscription" && params.TrialPeriodDays > 0 {
		formParams.Set("subscription_data[trial_period_days]", strconv.Itoa(params.TrialPeriodDays))
	}

	// Add discounts
	for i, discount := range params.Discounts {
		if discount.CouponID != "" {
			prefix := fmt.Sprintf("discounts[%d]", i)
			formParams.Set(prefix+"[coupon]", discount.CouponID)
		} else if discount.PromotionCode != "" {
			prefix := fmt.Sprintf("discounts[%d]", i)
			formParams.Set(prefix+"[promotion_code]", discount.PromotionCode)
		}
	}

	// Allow promotion codes
	if params.AllowPromoCodes {
		formParams.Set("allow_promotion_codes", "true")
	}

	// Payment method types
	if len(params.PaymentMethodTypes) > 0 {
		for i, pmType := range params.PaymentMethodTypes {
			formParams.Set(fmt.Sprintf("payment_method_types[%d]", i), pmType)
		}
	}

	// Payment method collection
	// "if_required" means Stripe won't ask for payment method during free trial
	formParams.Set("payment_method_collection", "if_required")

	// Make request
	resp, err := c.request("POST", "/checkout/sessions", formParams)
	if err != nil {
		return nil, err
	}

	// Parse response
	var stripeSession struct {
		ID             string                 `json:"id"`
		URL            string                 `json:"url"`
		Status         string                 `json:"status"`
		Customer       string                 `json:"customer"`
		CustomerEmail  string                 `json:"customer_email"`
		Subscription   string                 `json:"subscription"`
		PaymentIntent  string                 `json:"payment_intent"`
		AmountTotal    int64                  `json:"amount_total"`
		Currency       string                 `json:"currency"`
		Metadata       map[string]string      `json:"metadata"`
		ExpiresAt      int64                  `json:"expires_at"`
		CustomerDetails struct {
			Email string `json:"email"`
		} `json:"customer_details"`
	}

	if err := json.Unmarshal(resp, &stripeSession); err != nil {
		return nil, fmt.Errorf("failed to parse checkout session: %w", err)
	}

	session := &payments.CheckoutSession{
		ID:              stripeSession.ID,
		URL:             stripeSession.URL,
		Status:          stripeSession.Status,
		CustomerID:      stripeSession.Customer,
		CustomerEmail:   stripeSession.CustomerEmail,
		SubscriptionID:  stripeSession.Subscription,
		PaymentIntentID: stripeSession.PaymentIntent,
		AmountTotal:     stripeSession.AmountTotal,
		Currency:        stripeSession.Currency,
		Metadata:        stripeSession.Metadata,
		ExpiresAt:       parseTime(float64(stripeSession.ExpiresAt)),
	}

	// Use customer details email if available
	if session.CustomerEmail == "" && stripeSession.CustomerDetails.Email != "" {
		session.CustomerEmail = stripeSession.CustomerDetails.Email
	}

	return session, nil
}

// GetCheckoutSession retrieves a checkout session by ID
func (c *Client) GetCheckoutSession(sessionID string) (*payments.CheckoutSession, error) {
	// Make API request
	resp, err := c.request("GET", "/checkout/sessions/"+sessionID, nil)
	if err != nil {
		return nil, err
	}

	// Parse response
	var stripeSession struct {
		ID             string                 `json:"id"`
		URL            string                 `json:"url"`
		Status         string                 `json:"status"`
		Customer       string                 `json:"customer"`
		CustomerEmail  string                 `json:"customer_email"`
		Subscription   string                 `json:"subscription"`
		PaymentIntent  string                 `json:"payment_intent"`
		AmountTotal    int64                  `json:"amount_total"`
		Currency       string                 `json:"currency"`
		Metadata       map[string]string      `json:"metadata"`
		ExpiresAt      int64                  `json:"expires_at"`
		CustomerDetails struct {
			Email string `json:"email"`
		} `json:"customer_details"`
	}

	if err := json.Unmarshal(resp, &stripeSession); err != nil {
		return nil, fmt.Errorf("failed to parse checkout session: %w", err)
	}

	session := &payments.CheckoutSession{
		ID:              stripeSession.ID,
		URL:             stripeSession.URL,
		Status:          stripeSession.Status,
		CustomerID:      stripeSession.Customer,
		CustomerEmail:   stripeSession.CustomerEmail,
		SubscriptionID:  stripeSession.Subscription,
		PaymentIntentID: stripeSession.PaymentIntent,
		AmountTotal:     stripeSession.AmountTotal,
		Currency:        stripeSession.Currency,
		Metadata:        stripeSession.Metadata,
		ExpiresAt:       parseTime(float64(stripeSession.ExpiresAt)),
	}

	// Use customer details email if available
	if session.CustomerEmail == "" && stripeSession.CustomerDetails.Email != "" {
		session.CustomerEmail = stripeSession.CustomerDetails.Email
	}

	return session, nil
}

// CreatePortalSession creates a billing portal session for customer self-service
func (c *Client) CreatePortalSession(customerID, returnURL string) (*payments.PortalSession, error) {
	params := url.Values{}
	params.Set("customer", customerID)
	params.Set("return_url", returnURL)

	resp, err := c.request("POST", "/billing_portal/sessions", params)
	if err != nil {
		return nil, err
	}

	var stripeSession struct {
		ID        string `json:"id"`
		URL       string `json:"url"`
		Customer  string `json:"customer"`
		ReturnURL string `json:"return_url"`
		Created   int64  `json:"created"`
	}

	if err := json.Unmarshal(resp, &stripeSession); err != nil {
		return nil, fmt.Errorf("failed to parse portal session: %w", err)
	}

	return &payments.PortalSession{
		ID:         stripeSession.ID,
		URL:        stripeSession.URL,
		CustomerID: stripeSession.Customer,
		ReturnURL:  stripeSession.ReturnURL,
		Created:    parseTime(float64(stripeSession.Created)),
	}, nil
}