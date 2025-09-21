package payments

import (
	"fmt"
	"time"
)

// Error represents a payment provider error
type Error struct {
	Type    string
	Code    string
	Message string
	Param   string
}

func (e *Error) Error() string {
	return e.Message
}

// SimpleProduct is a concrete implementation of the Product interface
type SimpleProduct struct {
	id          string
	name        string
	description string
	priceID     string
	amount      int64
	currency    string
	interval    string
	metadata    map[string]string
}

func (p *SimpleProduct) ID() string                 { return p.id }
func (p *SimpleProduct) Name() string               { return p.name }
func (p *SimpleProduct) Description() string        { return p.description }
func (p *SimpleProduct) PriceID() string            { return p.priceID }
func (p *SimpleProduct) Amount() int64              { return p.amount }
func (p *SimpleProduct) Currency() string           { return p.currency }
func (p *SimpleProduct) Interval() string           { return p.interval }
func (p *SimpleProduct) Metadata() map[string]string { return p.metadata }

// SimpleCustomer is a concrete implementation of the Customer interface
type SimpleCustomer struct {
	id    string
	email string
	name  string
}

func (c *SimpleCustomer) ID() string    { return c.id }
func (c *SimpleCustomer) Email() string { return c.email }
func (c *SimpleCustomer) Name() string  { return c.name }

// SimpleCheckout is a concrete implementation of the Checkout interface
type SimpleCheckout struct {
	id             string
	url            string
	status         string
	customerID     string
	subscriptionID string
	metadata       map[string]string
}

func (c *SimpleCheckout) ID() string                 { return c.id }
func (c *SimpleCheckout) URL() string                { return c.url }
func (c *SimpleCheckout) Status() string             { return c.status }
func (c *SimpleCheckout) CustomerID() string         { return c.customerID }
func (c *SimpleCheckout) SubscriptionID() string     { return c.subscriptionID }
func (c *SimpleCheckout) Metadata() map[string]string { return c.metadata }

// SimpleSubscription is a concrete implementation of the Subscription interface
type SimpleSubscription struct {
	id               string
	customerID       string
	productID        string
	priceID          string
	status           string
	currentPeriodEnd int64
}

func (s *SimpleSubscription) ID() string               { return s.id }
func (s *SimpleSubscription) CustomerID() string       { return s.customerID }
func (s *SimpleSubscription) ProductID() string        { return s.productID }
func (s *SimpleSubscription) PriceID() string          { return s.priceID }
func (s *SimpleSubscription) Status() string           { return s.status }
func (s *SimpleSubscription) CurrentPeriodEnd() int64  { return s.currentPeriodEnd }

// SimpleEvent is a concrete implementation of the Event interface
type SimpleEvent struct {
	id      string
	typ     string
	data    map[string]interface{}
}

func (e *SimpleEvent) ID() string                   { return e.id }
func (e *SimpleEvent) Type() string                 { return e.typ }
func (e *SimpleEvent) Data() map[string]interface{} { return e.data }

// Legacy types for compatibility - these will be phased out

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusUnpaid   SubscriptionStatus = "unpaid"
	SubscriptionStatusPaused   SubscriptionStatus = "paused"
)

// WebhookEvent represents a webhook event from the payment provider
type WebhookEvent struct {
	ID      string
	Type    string
	Created time.Time
	Data    map[string]any
	Object  any
}

// Common webhook event types
const (
	EventTypeCheckoutComplete        = "checkout.session.completed"
	EventTypeCustomerCreated         = "customer.created"
	EventTypeCustomerUpdated         = "customer.updated"
	EventTypeCustomerDeleted         = "customer.deleted"
	EventTypeSubscriptionCreated     = "subscription.created"
	EventTypeSubscriptionUpdated     = "subscription.updated"
	EventTypeSubscriptionDeleted     = "subscription.deleted"
	EventTypePaymentIntentSucceeded  = "payment_intent.succeeded"
	EventTypePaymentIntentFailed     = "payment_intent.payment_failed"
	EventTypePaymentMethodAttached   = "payment_method.attached"
	EventTypePaymentMethodDetached   = "payment_method.detached"
)

// Legacy types removed - Customer is now an interface in backend.go

// Legacy param types removed - using option pattern with functional options instead

// Helper functions

func FormatAmount(amountCents int64, currency string) string {
	amount := float64(amountCents) / 100.0

	switch currency {
	case "usd":
		return fmt.Sprintf("$%.2f", amount)
	case "eur":
		return fmt.Sprintf("€%.2f", amount)
	case "gbp":
		return fmt.Sprintf("£%.2f", amount)
	case "jpy":
		return fmt.Sprintf("¥%d", amountCents)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

func CalculateTrialEnd(trialDays int) *time.Time {
	if trialDays <= 0 {
		return nil
	}
	end := time.Now().AddDate(0, 0, trialDays)
	return &end
}

func IsSubscriptionActive(status SubscriptionStatus) bool {
	return status == SubscriptionStatusActive || status == SubscriptionStatusTrialing
}

func IsPaymentRequired(status SubscriptionStatus) bool {
	return status == SubscriptionStatusPastDue || status == SubscriptionStatusUnpaid
}