package payments

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WebhookHandler handles Stripe webhook events
type WebhookHandler struct {
	client *StripeClient
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(client *StripeClient) *WebhookHandler {
	return &WebhookHandler{
		client: client,
	}
}

// HandleWebhook processes a webhook request from Stripe
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) error {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return fmt.Errorf("failed to read body: %w", err)
	}
	
	// Verify the signature
	signature := r.Header.Get("Stripe-Signature")
	if !h.client.VerifyWebhookSignature(string(body), signature) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return fmt.Errorf("invalid webhook signature")
	}
	
	// Parse the event
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return fmt.Errorf("failed to parse event: %w", err)
	}
	
	// Get event type
	eventType, ok := event["type"].(string)
	if !ok {
		http.Error(w, "Missing event type", http.StatusBadRequest)
		return fmt.Errorf("missing event type")
	}
	
	// Process based on event type
	switch eventType {
	case "checkout.session.completed":
		if err := h.handleCheckoutCompleted(event); err != nil {
			return fmt.Errorf("failed to handle checkout completed: %w", err)
		}
		
	case "customer.subscription.created":
		if err := h.handleSubscriptionCreated(event); err != nil {
			return fmt.Errorf("failed to handle subscription created: %w", err)
		}
		
	case "customer.subscription.updated":
		if err := h.handleSubscriptionUpdated(event); err != nil {
			return fmt.Errorf("failed to handle subscription updated: %w", err)
		}
		
	case "customer.subscription.deleted":
		if err := h.handleSubscriptionDeleted(event); err != nil {
			return fmt.Errorf("failed to handle subscription deleted: %w", err)
		}
		
	case "invoice.payment_succeeded":
		if err := h.handlePaymentSucceeded(event); err != nil {
			return fmt.Errorf("failed to handle payment succeeded: %w", err)
		}
		
	case "invoice.payment_failed":
		if err := h.handlePaymentFailed(event); err != nil {
			return fmt.Errorf("failed to handle payment failed: %w", err)
		}
		
	default:
		// Log unhandled event types
		fmt.Printf("Unhandled webhook event type: %s\n", eventType)
	}
	
	// Return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return nil
}

// Event handler methods - these should be overridden by the application

func (h *WebhookHandler) handleCheckoutCompleted(event map[string]interface{}) error {
	// Default implementation - override in application
	return nil
}

func (h *WebhookHandler) handleSubscriptionCreated(event map[string]interface{}) error {
	// Default implementation - override in application
	return nil
}

func (h *WebhookHandler) handleSubscriptionUpdated(event map[string]interface{}) error {
	// Default implementation - override in application
	return nil
}

func (h *WebhookHandler) handleSubscriptionDeleted(event map[string]interface{}) error {
	// Default implementation - override in application
	return nil
}

func (h *WebhookHandler) handlePaymentSucceeded(event map[string]interface{}) error {
	// Default implementation - override in application
	return nil
}

func (h *WebhookHandler) handlePaymentFailed(event map[string]interface{}) error {
	// Default implementation - override in application
	return nil
}

// ExtractEventData extracts common data from webhook events
func ExtractEventData(event map[string]interface{}) (eventID, customerID, subscriptionID string) {
	if id, ok := event["id"].(string); ok {
		eventID = id
	}
	
	if data, ok := event["data"].(map[string]interface{}); ok {
		if obj, ok := data["object"].(map[string]interface{}); ok {
			if customer, ok := obj["customer"].(string); ok {
				customerID = customer
			}
			if subscription, ok := obj["subscription"].(string); ok {
				subscriptionID = subscription
			}
		}
	}
	
	return eventID, customerID, subscriptionID
}