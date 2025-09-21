# Payments Package

A flexible payment processing package for Go applications with support for multiple payment providers.

## Architecture

The package uses a Backend interface pattern with static product configuration:
- Products are configured at initialization time (not created dynamically)
- Backends implement a common interface for provider-agnostic operations
- Option pattern for flexible configuration

## Installation

```go
import (
    "github.com/The-Skyscape/devtools/pkg/payments"
    "github.com/The-Skyscape/devtools/pkg/payments/backends/stripe"
    "github.com/The-Skyscape/devtools/pkg/payments/backends/mock"
)
```

## Usage

### Initialize a Payment Backend

```go
// Configure products at initialization
backend := stripe.NewClient(
    secretKey,
    publishableKey,
    webhookSecret,
    // Products are configured once at startup
    payments.WithProduct(
        "Workbench",
        "Cloud development environment",
        payments.WithMonthlyPrice(40.0),
        payments.WithMetadata("sku", "workbench-001"),
    ),
    payments.WithProduct(
        "Workspace Pro",
        "Professional workspace with AI",
        payments.WithMonthlyPrice(800.0),
        payments.WithTrialDays(14),
    ),
)
```

### Working with Products

```go
// Get a specific product by name
product, err := backend.ProductByName("Workbench")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Product: %s - $%.2f/month\n",
    product.Name(),
    float64(product.Amount())/100)

// List all configured products
products, err := backend.Products()
for _, p := range products {
    fmt.Printf("%s: %s\n", p.Name(), p.Description())
}
```

### Customer Management

```go
// Create a customer with options
customer, err := backend.CreateCustomer(
    "john@example.com",
    "John Doe",
    payments.WithCustomerPhone("+1234567890"),
    payments.WithCustomerMetadata("user_id", "usr_123"),
)

// Retrieve customer
customer, err := backend.Customer(customerID)
fmt.Printf("Customer: %s (%s)\n", customer.Name(), customer.Email())

// List customers
customers, err := backend.Customers(100) // limit: 100
```

### Checkout Sessions

```go
// Create checkout session for a product
checkout, err := backend.CreateCheckout(
    "Workbench", // Product name (must be configured)
    payments.WithCustomerID(customerID),
    payments.WithSuccessURL("https://example.com/success"),
    payments.WithCancelURL("https://example.com/cancel"),
    payments.WithTrialDays(7),
    payments.WithCheckoutMetadata("order_id", "ord_123"),
)

// Redirect user to checkout
fmt.Printf("Redirect to: %s\n", checkout.URL())

// Retrieve checkout status
checkout, err = backend.Checkout(sessionID)
if checkout.Status() == "complete" {
    fmt.Printf("Subscription ID: %s\n", checkout.SubscriptionID())
}
```

### Subscription Management

```go
// Get subscription details
subscription, err := backend.Subscription(subscriptionID)
fmt.Printf("Status: %s, Ends: %v\n",
    subscription.Status(),
    time.Unix(subscription.CurrentPeriodEnd(), 0))

// Pause subscription (collection_paused)
err = backend.PauseSubscription(subscriptionID)

// Resume subscription
err = backend.ResumeSubscription(subscriptionID)

// List subscriptions
subscriptions, err := backend.Subscriptions(100)
```

### Customer Portal

```go
// Create portal session for customer self-service
portalURL, err := backend.CreatePortalSession(
    customerID,
    "https://example.com/account", // Return URL
)

// Redirect customer to portal
fmt.Printf("Manage billing: %s\n", portalURL)
```

### Webhook Handling

```go
// In your webhook handler
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("Stripe-Signature")

    // Verify and construct event
    event, err := backend.ConstructWebhookEvent(payload, signature)
    if err != nil {
        http.Error(w, "Invalid signature", 400)
        return
    }

    // Handle different event types
    switch event.Type() {
    case "checkout.session.completed":
        handleCheckoutComplete(event)
    case "customer.subscription.updated":
        handleSubscriptionUpdate(event)
    }

    w.WriteHeader(200)
}

// Extract data from events
func handleSubscriptionUpdate(event payments.Event) {
    data := event.Data()

    // Access webhook data (provider-specific structure)
    if obj, ok := data["object"].(map[string]any); ok {
        customerID, _ := obj["customer"].(string)
        status, _ := obj["status"].(string)

        fmt.Printf("Customer %s subscription is %s\n",
            customerID, status)
    }
}
```

## Testing with Mock Backend

```go
// Use mock backend for testing
backend := mock.New(
    payments.WithProduct(
        "Test Product",
        "For testing",
        payments.WithMonthlyPrice(10.0),
    ),
)

// Mock backend supports all the same operations
customer, _ := backend.CreateCustomer("test@example.com", "Test User")
checkout, _ := backend.CreateCheckout("Test Product",
    payments.WithCustomerID(customer.ID()),
)

// Control test behavior
mockBackend := backend.(*mock.Backend)
mockBackend.FailNext("Simulated error")
_, err := backend.CreateCustomer("fail@example.com", "Fail User")
// err will be "mock error: Simulated error"
```

## Interfaces

The package defines clean interfaces for all entities:

```go
type Backend interface {
    // Product operations (read-only after init)
    Product(id string) (Product, error)
    ProductByName(name string) (Product, error)
    Products() ([]Product, error)

    // Customer operations
    CreateCustomer(email, name string, opts ...CustomerOption) (Customer, error)
    Customer(id string) (Customer, error)
    Customers(limit int) ([]Customer, error)

    // Checkout operations
    CreateCheckout(productName string, opts ...CheckoutOption) (Checkout, error)
    Checkout(id string) (Checkout, error)

    // Subscription operations
    Subscription(id string) (Subscription, error)
    Subscriptions(limit int) ([]Subscription, error)
    PauseSubscription(id string) error
    ResumeSubscription(id string) error

    // Portal operations
    CreatePortalSession(customerID, returnURL string) (string, error)

    // Webhook operations
    ConstructWebhookEvent(payload []byte, signature string) (Event, error)
}
```

## Helper Functions

The package includes utility functions in `types.go`:

```go
// Format amount for display
formatted := payments.FormatAmount(1999, "usd") // "$19.99"

// Check subscription status
active := payments.IsSubscriptionActive(status)

// Check if payment is required
needsPayment := payments.IsPaymentRequired(status)

// Calculate trial end date
trialEnd := payments.CalculateTrialEnd(14) // 14 days from now
```

## Best Practices

1. **Configure products once** at application startup
2. **Use the Backend interface** for provider-agnostic code
3. **Handle webhook events** asynchronously
4. **Store customer IDs** to link your users with payment providers
5. **Use metadata** to associate provider records with your data
6. **Test with mock backend** before using real providers

## Supported Providers

- **Stripe** - Full implementation with all features
- **Mock** - For testing and development

## Error Handling

```go
// The package uses standard Go error handling
customer, err := backend.CreateCustomer(email, name)
if err != nil {
    log.Printf("Failed to create customer: %v", err)
    return err
}
```

## Migration from Legacy API

If migrating from the old service-based API:
- Replace `Service` with `Backend` interface
- Configure products at initialization (not runtime)
- Use option functions instead of param structs
- Update webhook handling to use `Event` interface
- Interface types no longer have `I` prefix (e.g., `ICustomer` → `Customer`)