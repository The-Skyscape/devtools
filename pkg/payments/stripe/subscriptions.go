package stripe

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/The-Skyscape/devtools/pkg/payments"
)

// CreateSubscription creates a new subscription
func (c *Client) CreateSubscription(params *payments.SubscriptionParams) (*payments.Subscription, error) {
	formParams := url.Values{}
	formParams.Set("customer", params.CustomerID)
	
	// Add price IDs
	for i, priceID := range params.PriceIDs {
		formParams.Set(fmt.Sprintf("items[%d][price]", i), priceID)
	}
	
	// Add trial period
	if params.TrialPeriodDays > 0 {
		formParams.Set("trial_period_days", strconv.Itoa(params.TrialPeriodDays))
	}
	
	// Add cancel at period end
	if params.CancelAtPeriodEnd {
		formParams.Set("cancel_at_period_end", "true")
	}
	
	// Add proration behavior
	if params.ProrationBehavior != "" {
		formParams.Set("proration_behavior", params.ProrationBehavior)
	}
	
	// Add payment behavior
	if params.PaymentBehavior != "" {
		formParams.Set("payment_behavior", params.PaymentBehavior)
	}
	
	// Add metadata
	formatMetadata(formParams, params.Metadata)
	
	// Make request
	resp, err := c.request("POST", "/subscriptions", formParams)
	if err != nil {
		return nil, err
	}
	
	return c.parseSubscription(resp)
}

// GetSubscription retrieves a subscription by ID
func (c *Client) GetSubscription(subscriptionID string) (*payments.Subscription, error) {
	resp, err := c.request("GET", "/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return nil, err
	}
	
	return c.parseSubscription(resp)
}

// UpdateSubscription updates a subscription
func (c *Client) UpdateSubscription(subscriptionID string, params *payments.SubscriptionParams) (*payments.Subscription, error) {
	formParams := url.Values{}
	
	// Update items if provided
	if len(params.PriceIDs) > 0 {
		// First, we need to clear existing items
		formParams.Set("items[0][clear_usage]", "true")
		formParams.Set("items[0][deleted]", "true")
		
		// Then add new items
		for i, priceID := range params.PriceIDs {
			formParams.Set(fmt.Sprintf("items[%d][price]", i+1), priceID)
		}
	}
	
	// Update trial period (can only extend, not shorten)
	if params.TrialPeriodDays > 0 {
		formParams.Set("trial_period_days", strconv.Itoa(params.TrialPeriodDays))
	}
	
	// Update cancel at period end
	formParams.Set("cancel_at_period_end", strconv.FormatBool(params.CancelAtPeriodEnd))
	
	// Update proration behavior
	if params.ProrationBehavior != "" {
		formParams.Set("proration_behavior", params.ProrationBehavior)
	}
	
	// Update payment behavior
	if params.PaymentBehavior != "" {
		formParams.Set("payment_behavior", params.PaymentBehavior)
	}
	
	// Update metadata
	if len(params.Metadata) > 0 {
		formatMetadata(formParams, params.Metadata)
	}
	
	// Make request
	resp, err := c.request("POST", "/subscriptions/"+subscriptionID, formParams)
	if err != nil {
		return nil, err
	}
	
	return c.parseSubscription(resp)
}

// CancelSubscription cancels a subscription
func (c *Client) CancelSubscription(subscriptionID string, immediately bool) (*payments.Subscription, error) {
	formParams := url.Values{}
	
	if immediately {
		// Cancel immediately
		formParams.Set("invoice_now", "true")
		formParams.Set("prorate", "true")
	} else {
		// Cancel at period end
		formParams.Set("cancel_at_period_end", "true")
	}
	
	endpoint := "/subscriptions/" + subscriptionID
	if immediately {
		endpoint = "/subscriptions/" + subscriptionID + "/cancel"
	}
	
	resp, err := c.request("POST", endpoint, formParams)
	if err != nil {
		return nil, err
	}
	
	return c.parseSubscription(resp)
}

// parseSubscription parses a Stripe subscription response
func (c *Client) parseSubscription(data []byte) (*payments.Subscription, error) {
	var stripeSub struct {
		ID                 string            `json:"id"`
		Customer           string            `json:"customer"`
		Status             string            `json:"status"`
		CurrentPeriodStart int64             `json:"current_period_start"`
		CurrentPeriodEnd   int64             `json:"current_period_end"`
		TrialStart         *int64            `json:"trial_start"`
		TrialEnd           *int64            `json:"trial_end"`
		CancelAt           *int64            `json:"cancel_at"`
		CanceledAt         *int64            `json:"canceled_at"`
		EndedAt            *int64            `json:"ended_at"`
		Created            int64             `json:"created"`
		Metadata           map[string]string `json:"metadata"`
		Items              struct {
			Data []struct {
				ID       string            `json:"id"`
				Price    struct {
					ID string `json:"id"`
				} `json:"price"`
				Quantity int               `json:"quantity"`
				Metadata map[string]string `json:"metadata"`
			} `json:"data"`
		} `json:"items"`
	}
	
	if err := json.Unmarshal(data, &stripeSub); err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}
	
	// Convert Stripe status to generic status
	status := c.parseSubscriptionStatus(stripeSub.Status)
	
	// Parse items
	items := make([]payments.SubscriptionItem, 0, len(stripeSub.Items.Data))
	for _, item := range stripeSub.Items.Data {
		items = append(items, payments.SubscriptionItem{
			ID:       item.ID,
			PriceID:  item.Price.ID,
			Quantity: item.Quantity,
			Metadata: item.Metadata,
		})
	}
	
	// Parse times
	var trialStart, trialEnd, cancelAt, cancelledAt, endedAt *time.Time
	if stripeSub.TrialStart != nil {
		t := parseTime(float64(*stripeSub.TrialStart))
		trialStart = &t
	}
	if stripeSub.TrialEnd != nil {
		t := parseTime(float64(*stripeSub.TrialEnd))
		trialEnd = &t
	}
	if stripeSub.CancelAt != nil {
		t := parseTime(float64(*stripeSub.CancelAt))
		cancelAt = &t
	}
	if stripeSub.CanceledAt != nil {
		t := parseTime(float64(*stripeSub.CanceledAt))
		cancelledAt = &t
	}
	if stripeSub.EndedAt != nil {
		t := parseTime(float64(*stripeSub.EndedAt))
		endedAt = &t
	}
	
	return &payments.Subscription{
		ID:                 stripeSub.ID,
		CustomerID:         stripeSub.Customer,
		Status:             status,
		CurrentPeriodStart: parseTime(float64(stripeSub.CurrentPeriodStart)),
		CurrentPeriodEnd:   parseTime(float64(stripeSub.CurrentPeriodEnd)),
		TrialStart:         trialStart,
		TrialEnd:           trialEnd,
		CancelAt:           cancelAt,
		CancelledAt:        cancelledAt,
		EndedAt:            endedAt,
		Items:              items,
		Metadata:           stripeSub.Metadata,
		Created:            parseTime(float64(stripeSub.Created)),
		Updated:            parseTime(float64(stripeSub.Created)), // Stripe doesn't have updated field
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