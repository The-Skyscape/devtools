package payments

import (
	"time"
)

// Provider defines the interface for payment providers
type Provider interface {
	// Checkout operations
	CreateCheckoutSession(params *CheckoutParams) (*CheckoutSession, error)
	GetCheckoutSession(sessionID string) (*CheckoutSession, error)
	
	// Customer operations
	CreateCustomer(params *CustomerParams) (*Customer, error)
	GetCustomer(customerID string) (*Customer, error)
	UpdateCustomer(customerID string, params *CustomerParams) (*Customer, error)
	DeleteCustomer(customerID string) error
	
	// Subscription operations
	CreateSubscription(params *SubscriptionParams) (*Subscription, error)
	GetSubscription(subscriptionID string) (*Subscription, error)
	UpdateSubscription(subscriptionID string, params *SubscriptionParams) (*Subscription, error)
	CancelSubscription(subscriptionID string, immediately bool) (*Subscription, error)
	
	// Payment method operations
	AttachPaymentMethod(customerID, paymentMethodID string) error
	DetachPaymentMethod(paymentMethodID string) error
	SetDefaultPaymentMethod(customerID, paymentMethodID string) error
	
	// Webhook operations
	VerifyWebhookSignature(payload []byte, signature string) bool
	ParseWebhookEvent(payload []byte) (*WebhookEvent, error)
	
	// Portal operations (for subscription management)
	CreatePortalSession(customerID, returnURL string) (*PortalSession, error)
	
	// Provider identification
	GetName() string
}

// CheckoutParams defines parameters for creating a checkout session
type CheckoutParams struct {
	Mode            string            // "payment" or "subscription"
	SuccessURL      string
	CancelURL       string
	CustomerID      string            // Optional existing customer
	CustomerEmail   string            // For new customers
	LineItems       []LineItem
	Metadata        map[string]string
	TrialPeriodDays int              // For subscriptions
	Discounts       []Discount       // Optional discounts/coupons
	AllowPromoCodes bool
	PaymentMethodTypes []string      // Optional payment method restrictions
}

// LineItem represents a product/price in checkout
type LineItem struct {
	PriceID     string // Provider's price ID
	ProductID   string // Optional product ID
	Quantity    int
	Description string // For one-time payments
	Amount      int64  // For one-time payments (in cents)
	Currency    string // For one-time payments
}

// Discount represents a discount/coupon
type Discount struct {
	CouponID   string
	PromotionCode string
}

// CheckoutSession represents a checkout session
type CheckoutSession struct {
	ID             string
	URL            string            // URL to redirect customer to
	Status         string            // open, complete, expired
	CustomerID     string
	CustomerEmail  string
	SubscriptionID string            // For subscription mode
	PaymentIntentID string           // For payment mode
	AmountTotal    int64             // Total amount in cents
	Currency       string
	Metadata       map[string]string
	ExpiresAt      time.Time
}

// CustomerParams defines parameters for customer operations
type CustomerParams struct {
	Email       string
	Name        string
	Phone       string
	Description string
	Metadata    map[string]string
	TaxID       string
	Address     *Address
}

// Customer represents a customer in the payment system
type Customer struct {
	ID               string
	Email            string
	Name             string
	Phone            string
	Description      string
	DefaultPaymentMethod string
	Metadata         map[string]string
	Created          time.Time
	Updated          time.Time
	Balance          int64  // Account balance in cents
	Currency         string
	Delinquent       bool   // Has any unpaid invoices
}

// Address represents a billing/shipping address
type Address struct {
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
}

// SubscriptionParams defines parameters for subscription operations
type SubscriptionParams struct {
	CustomerID      string
	PriceIDs        []string          // Price IDs to subscribe to
	TrialPeriodDays int
	Metadata        map[string]string
	CancelAtPeriodEnd bool            // Cancel at end of billing period
	ProrationBehavior string          // How to handle proration
	PaymentBehavior   string          // Payment collection behavior
}

// Subscription represents a subscription
type Subscription struct {
	ID                string
	CustomerID        string
	Status            SubscriptionStatus
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	TrialStart         *time.Time
	TrialEnd           *time.Time
	CancelAt           *time.Time
	CancelledAt        *time.Time
	EndedAt            *time.Time
	Items              []SubscriptionItem
	Metadata           map[string]string
	Created            time.Time
	Updated            time.Time
}

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

// SubscriptionItem represents a line item in a subscription
type SubscriptionItem struct {
	ID       string
	PriceID  string
	Quantity int
	Metadata map[string]string
}

// WebhookEvent represents a webhook event from the payment provider
type WebhookEvent struct {
	ID        string
	Type      string
	Created   time.Time
	Data      map[string]interface{}
	Object    interface{} // The actual object (subscription, payment, etc.)
}

// Common webhook event types (provider-agnostic)
const (
	EventTypeCheckoutComplete         = "checkout.session.completed"
	EventTypeCustomerCreated          = "customer.created"
	EventTypeCustomerUpdated          = "customer.updated"
	EventTypeCustomerDeleted          = "customer.deleted"
	EventTypeSubscriptionCreated      = "subscription.created"
	EventTypeSubscriptionUpdated      = "subscription.updated"
	EventTypeSubscriptionDeleted      = "subscription.deleted"
	EventTypePaymentIntentSucceeded  = "payment_intent.succeeded"
	EventTypePaymentIntentFailed     = "payment_intent.payment_failed"
	EventTypePaymentMethodAttached   = "payment_method.attached"
	EventTypePaymentMethodDetached   = "payment_method.detached"
)

// PortalSession represents a customer portal session
type PortalSession struct {
	ID        string
	URL       string    // URL to redirect customer to
	CustomerID string
	ReturnURL string
	Created   time.Time
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID         string
	Type       string // card, bank_account, etc.
	CustomerID string
	IsDefault  bool
	Card       *CardDetails
	Created    time.Time
}

// CardDetails represents credit card details
type CardDetails struct {
	Brand      string // visa, mastercard, amex, etc.
	Last4      string
	ExpMonth   int
	ExpYear    int
	Country    string
}