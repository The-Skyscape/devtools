# Payments Package

A generic payment processing toolkit with support for multiple payment providers.

## Features

- **Provider Agnostic**: Clean interface for multiple payment providers
- **Stripe Support**: Full Stripe implementation included
- **Type Safety**: Strongly typed models for all payment entities
- **Webhook Handling**: Secure webhook verification and parsing
- **Subscription Management**: Complete subscription lifecycle support
- **Customer Portal**: Self-service portal for customers

## Installation

```go
import (
    "github.com/The-Skyscape/devtools/pkg/payments"
    "github.com/The-Skyscape/devtools/pkg/payments/stripe"
)
```

## Quick Start

```go
// Create service
service := payments.NewService()

// Add Stripe provider
stripeClient := stripe.NewClient(
    "sk_test_SECRET_KEY",
    "pk_test_PUBLISHABLE_KEY",
    "whsec_WEBHOOK_SECRET",
)
service.AddProvider("stripe", stripeClient)
service.SetActiveProvider("stripe")

// Create checkout session
session, err := service.CreateCheckoutSession(&payments.CheckoutParams{
    Mode:       "subscription",
    SuccessURL: "https://example.com/success",
    CancelURL:  "https://example.com/cancel",
    LineItems: []payments.LineItem{
        {PriceID: "price_XXX", Quantity: 1},
    },
})

// Redirect customer to session.URL
```

## Core Operations

### Checkout Sessions
```go
// Create checkout
session, err := service.CreateCheckoutSession(&payments.CheckoutParams{
    Mode:            "subscription", // or "payment"
    CustomerEmail:   "customer@example.com",
    TrialPeriodDays: 14,
    LineItems:       []payments.LineItem{...},
})

// Get session
session, err := service.GetCheckoutSession(sessionID)
```

### Customer Management
```go
// Create customer
customer, err := service.CreateCustomer(&payments.CustomerParams{
    Email: "john@example.com",
    Name:  "John Doe",
})

// Update customer
customer, err := service.UpdateCustomer(customerID, &payments.CustomerParams{
    Phone: "+1234567890",
})

// Delete customer
err := service.DeleteCustomer(customerID)
```

### Subscription Management
```go
// Create subscription
sub, err := service.CreateSubscription(&payments.SubscriptionParams{
    CustomerID: customerID,
    PriceIDs:   []string{"price_XXX"},
})

// Cancel subscription
sub, err := service.CancelSubscription(subID, false) // false = at period end
```

### Webhook Handling
```go
// Verify signature
if service.VerifyWebhookSignature(payload, signature) {
    // Parse event
    event, err := service.ParseWebhookEvent(payload)
    
    // Handle event
    switch event.Type {
    case payments.EventTypeSubscriptionCreated:
        // Handle new subscription
    case payments.EventTypePaymentIntentSucceeded:
        // Handle successful payment
    }
}
```

### Customer Portal
```go
// Create portal session for self-service
portal, err := service.CreatePortalSession(
    customerID,
    "https://example.com/account", // Return URL
)
// Redirect to portal.URL
```

## Common Types

### Subscription Status
- `SubscriptionStatusActive` - Active subscription
- `SubscriptionStatusTrialing` - In trial period
- `SubscriptionStatusPastDue` - Payment failed but retrying
- `SubscriptionStatusCanceled` - Canceled
- `SubscriptionStatusUnpaid` - First payment failed

### Event Types
- `EventTypeCheckoutComplete` - Checkout session completed
- `EventTypeSubscriptionCreated` - New subscription
- `EventTypeSubscriptionUpdated` - Subscription changed
- `EventTypePaymentIntentSucceeded` - Payment successful

## Helper Functions

```go
// Format amount for display
formatted := payments.FormatAmount(1999, "usd") // "$19.99"

// Check subscription status
active := payments.IsSubscriptionActive(sub.Status)

// Calculate trial end date
trialEnd := payments.CalculateTrialEnd(14) // 14 days from now
```

## Error Handling

```go
if err != nil {
    if paymentErr, ok := err.(*payments.Error); ok {
        log.Printf("Payment error: %s (code: %s)", 
            paymentErr.Message, 
            paymentErr.Code)
    }
}
```

## Future Providers

The package is designed to support additional providers:
- PayPal (planned)
- Square (planned)
- Custom implementations via the Provider interface