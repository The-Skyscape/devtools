package payments

import (
	"fmt"
	"time"
)

// Price represents a price/plan that can be subscribed to
type Price struct {
	ID          string
	ProductID   string
	Amount      int64  // Amount in cents
	Currency    string
	Interval    string // month, year, week, day
	IntervalCount int  // Number of intervals between billings
	TrialDays   int    // Trial period in days
	Active      bool
	Metadata    map[string]string
}

// Product represents a product that can be sold
type Product struct {
	ID          string
	Name        string
	Description string
	Active      bool
	Metadata    map[string]string
	Created     time.Time
	Updated     time.Time
}

// Invoice represents a billing invoice
type Invoice struct {
	ID                string
	CustomerID        string
	SubscriptionID    string
	Number            string
	Status            InvoiceStatus
	AmountDue         int64  // Amount in cents
	AmountPaid        int64
	AmountRemaining   int64
	Currency          string
	DueDate           *time.Time
	PaidAt            *time.Time
	HostedInvoiceURL  string
	InvoicePDF        string
	Created           time.Time
}

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusVoid          InvoiceStatus = "void"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
)

// PaymentIntent represents an intent to collect payment
type PaymentIntent struct {
	ID             string
	Amount         int64  // Amount in cents
	Currency       string
	Status         PaymentIntentStatus
	CustomerID     string
	PaymentMethodID string
	Description    string
	Metadata       map[string]string
	Created        time.Time
}

// PaymentIntentStatus represents the status of a payment intent
type PaymentIntentStatus string

const (
	PaymentIntentStatusRequiresPaymentMethod PaymentIntentStatus = "requires_payment_method"
	PaymentIntentStatusRequiresConfirmation  PaymentIntentStatus = "requires_confirmation"
	PaymentIntentStatusRequiresAction        PaymentIntentStatus = "requires_action"
	PaymentIntentStatusProcessing            PaymentIntentStatus = "processing"
	PaymentIntentStatusCanceled              PaymentIntentStatus = "canceled"
	PaymentIntentStatusSucceeded             PaymentIntentStatus = "succeeded"
)

// Refund represents a payment refund
type Refund struct {
	ID              string
	Amount          int64  // Amount in cents
	Currency        string
	PaymentIntentID string
	Reason          string
	Status          RefundStatus
	Metadata        map[string]string
	Created         time.Time
}

// RefundStatus represents the status of a refund
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
	RefundStatusCanceled  RefundStatus = "canceled"
)

// Coupon represents a discount coupon
type Coupon struct {
	ID               string
	Name             string
	AmountOff        int64   // Fixed amount discount in cents
	PercentOff       float64 // Percentage discount
	Currency         string
	Duration         string  // once, repeating, forever
	DurationInMonths int     // For repeating duration
	MaxRedemptions   int
	RedeemBy         *time.Time
	Valid            bool
	Created          time.Time
}

// UsageRecord represents usage for metered billing
type UsageRecord struct {
	ID               string
	SubscriptionItemID string
	Quantity         int64
	Timestamp        time.Time
	Action           string // increment or set
}

// BillingDetails represents billing information
type BillingDetails struct {
	Name    string
	Email   string
	Phone   string
	Address *Address
	TaxID   string
}

// Error represents a payment provider error
type Error struct {
	Code    string
	Message string
	Type    string // api_error, card_error, validation_error, etc.
	Param   string // The parameter the error relates to
}

func (e *Error) Error() string {
	if e.Param != "" {
		return fmt.Sprintf("%s: %s (param: %s)", e.Code, e.Message, e.Param)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Currency constants
const (
	CurrencyUSD = "usd"
	CurrencyEUR = "eur"
	CurrencyGBP = "gbp"
	CurrencyCAD = "cad"
	CurrencyAUD = "aud"
	CurrencyJPY = "jpy"
)

// Interval constants for recurring prices
const (
	IntervalDay   = "day"
	IntervalWeek  = "week"
	IntervalMonth = "month"
	IntervalYear  = "year"
)

// ProrationBehavior constants
const (
	ProrationCreate     = "create_prorations"
	ProrationNone       = "none"
	ProrationAlwaysInvoice = "always_invoice"
)

// PaymentBehavior constants
const (
	PaymentAllowIncomplete         = "allow_incomplete"
	PaymentDefaultIncomplete       = "default_incomplete"
	PaymentErrorIfIncomplete       = "error_if_incomplete"
	PaymentPendingIfIncomplete     = "pending_if_incomplete"
)

// Helper functions

// FormatAmount formats an amount in cents to a currency string
func FormatAmount(amountCents int64, currency string) string {
	// Convert cents to dollars (or equivalent)
	amount := float64(amountCents) / 100.0
	
	switch currency {
	case CurrencyUSD:
		return fmt.Sprintf("$%.2f", amount)
	case CurrencyEUR:
		return fmt.Sprintf("€%.2f", amount)
	case CurrencyGBP:
		return fmt.Sprintf("£%.2f", amount)
	case CurrencyJPY:
		// Japanese Yen doesn't use decimal places
		return fmt.Sprintf("¥%d", amountCents)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

// CalculateTrialEnd calculates the trial end date from now
func CalculateTrialEnd(trialDays int) *time.Time {
	if trialDays <= 0 {
		return nil
	}
	end := time.Now().AddDate(0, 0, trialDays)
	return &end
}

// IsSubscriptionActive checks if a subscription is in an active state
func IsSubscriptionActive(status SubscriptionStatus) bool {
	return status == SubscriptionStatusActive || status == SubscriptionStatusTrialing
}

// IsPaymentRequired checks if payment is required for a subscription status
func IsPaymentRequired(status SubscriptionStatus) bool {
	return status == SubscriptionStatusPastDue || status == SubscriptionStatusUnpaid
}