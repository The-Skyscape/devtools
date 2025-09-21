package stripe

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/The-Skyscape/devtools/pkg/payments"
)


// Subscription retrieves a subscription by ID (implements Backend interface)
func (c *Client) Subscription(subscriptionID string) (payments.Subscription, error) {
	resp, err := c.request("GET", "/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	// Parse response
	var stripeSubscription struct {
		ID                 string `json:"id"`
		Customer           string `json:"customer"`
		Status             string `json:"status"`
		CurrentPeriodEnd   int64  `json:"current_period_end"`
		Items              struct {
			Data []struct {
				Price struct {
					ID        string `json:"id"`
					ProductID string `json:"product"`
				} `json:"price"`
			} `json:"data"`
		} `json:"items"`
	}

	if err := json.Unmarshal(resp, &stripeSubscription); err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}

	// Get product ID from first item
	productID := ""
	priceID := ""
	if len(stripeSubscription.Items.Data) > 0 {
		priceID = stripeSubscription.Items.Data[0].Price.ID
		productID = stripeSubscription.Items.Data[0].Price.ProductID
	}

	return &StripeSubscription{
		id:               stripeSubscription.ID,
		customerID:       stripeSubscription.Customer,
		productID:        productID,
		priceID:          priceID,
		status:           stripeSubscription.Status,
		currentPeriodEnd: stripeSubscription.CurrentPeriodEnd,
	}, nil
}

// Subscriptions retrieves a list of subscriptions (implements Backend interface)
func (c *Client) Subscriptions(limit int) ([]payments.Subscription, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	} else {
		params.Set("limit", "100") // Default limit
	}

	resp, err := c.request("GET", "/subscriptions", params)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response struct {
		Data []struct {
			ID                 string `json:"id"`
			Customer           string `json:"customer"`
			Status             string `json:"status"`
			CurrentPeriodEnd   int64  `json:"current_period_end"`
			Items              struct {
				Data []struct {
					Price struct {
						ID        string `json:"id"`
						ProductID string `json:"product"`
					} `json:"price"`
				} `json:"data"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions list: %w", err)
	}

	subscriptions := make([]payments.Subscription, len(response.Data))
	for i, sub := range response.Data {
		// Get product ID from first item
		productID := ""
		priceID := ""
		if len(sub.Items.Data) > 0 {
			priceID = sub.Items.Data[0].Price.ID
			productID = sub.Items.Data[0].Price.ProductID
		}

		subscriptions[i] = &StripeSubscription{
			id:               sub.ID,
			customerID:       sub.Customer,
			productID:        productID,
			priceID:          priceID,
			status:           sub.Status,
			currentPeriodEnd: sub.CurrentPeriodEnd,
		}
	}

	return subscriptions, nil
}

// PauseSubscription pauses a subscription (implements Backend interface)
func (c *Client) PauseSubscription(id string) error {
	params := url.Values{}
	params.Set("pause_collection[behavior]", "mark_uncollectible")

	_, err := c.request("POST", "/subscriptions/"+id, params)
	return err
}

// ResumeSubscription resumes a subscription (implements Backend interface)
func (c *Client) ResumeSubscription(id string) error {
	params := url.Values{}
	params.Set("pause_collection", "") // Empty string removes pause

	_, err := c.request("POST", "/subscriptions/"+id, params)
	return err
}

// CreatePortalSession creates a customer portal session (implements Backend interface)
func (c *Client) CreatePortalSession(customerID, returnURL string) (string, error) {
	params := url.Values{}
	params.Set("customer", customerID)
	params.Set("return_url", returnURL)

	resp, err := c.request("POST", "/billing_portal/sessions", params)
	if err != nil {
		return "", err
	}

	var session struct {
		URL string `json:"url"`
	}

	if err := json.Unmarshal(resp, &session); err != nil {
		return "", fmt.Errorf("failed to parse portal session: %w", err)
	}

	return session.URL, nil
}

// ConstructWebhookEvent verifies signature and constructs the webhook event (implements Backend interface)
func (c *Client) ConstructWebhookEvent(payload []byte, signature string) (payments.Event, error) {
	// Verify signature
	if !c.VerifyWebhookSignature(payload, signature) {
		return nil, fmt.Errorf("invalid webhook signature")
	}

	// Parse event
	event, err := c.ParseWebhookEvent(payload)
	if err != nil {
		return nil, err
	}

	return &StripeEvent{
		id:   event.ID,
		typ:  event.Type,
		data: event.Data,
	}, nil
}

// parseSubscriptionStatus converts Stripe status to generic status
func (c *Client) parseSubscriptionStatus(stripeStatus string) payments.SubscriptionStatus {
	switch stripeStatus {
	case "active":
		return payments.SubscriptionStatusActive
	case "trialing":
		return payments.SubscriptionStatusTrialing
	case "past_due":
		return payments.SubscriptionStatusPastDue
	case "canceled":
		return payments.SubscriptionStatusCanceled
	case "unpaid":
		return payments.SubscriptionStatusUnpaid
	case "paused":
		return payments.SubscriptionStatusPaused
	default:
		// Return as-is if not mapped
		return payments.SubscriptionStatus(stripeStatus)
	}
}