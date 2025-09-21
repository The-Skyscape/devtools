package payments

// Backend defines the minimal interface for payment backends
// Products are configured at initialization time using BackendOptions
type Backend interface {
	// Product operations - products are configured at init, not created dynamically
	Product(id string) (Product, error)
	ProductByName(name string) (Product, error)
	Products() ([]Product, error)

	// Customer operations
	CreateCustomer(email, name string, opts ...CustomerOption) (Customer, error)
	Customer(id string) (Customer, error)
	Customers(limit int) ([]Customer, error)

	// Checkout operations - references products by name
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

// Product represents a product in the payment system (read-only after initialization)
type Product interface {
	ID() string
	Name() string
	Description() string
	PriceID() string // Default price ID
	Amount() int64   // Price in cents
	Currency() string
	Interval() string // "month", "year", etc
	Metadata() map[string]string
}

// Customer represents a customer in the payment system
type Customer interface {
	ID() string
	Email() string
	Name() string
}

// Checkout represents a checkout session
type Checkout interface {
	ID() string
	URL() string
	Status() string
	CustomerID() string
	SubscriptionID() string // Set after successful checkout
	Metadata() map[string]string
}

// Subscription represents a subscription
type Subscription interface {
	ID() string
	CustomerID() string
	ProductID() string
	PriceID() string
	Status() string
	CurrentPeriodEnd() int64 // Unix timestamp
}

// Event represents a webhook event
type Event interface {
	ID() string
	Type() string
	Data() map[string]interface{}
}