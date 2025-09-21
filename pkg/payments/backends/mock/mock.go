package mock

import (
	"fmt"
	"sync"
	"time"

	"github.com/The-Skyscape/devtools/pkg/payments"
)

// Backend provides a mock payment backend for testing
type Backend struct {
	name string

	// Storage
	products      map[string]*MockProduct
	customers     map[string]*MockCustomer
	subscriptions map[string]*MockSubscription
	sessions      map[string]*MockCheckout

	// Test control
	shouldFail    bool
	failMessage   string
	webhookSecret string

	// Counters
	customerCount int
	sessionCount  int
	subCount      int
	productCount  int

	mu sync.RWMutex
}

// New creates a new mock payment backend with configured products
func New(opts ...payments.BackendOption) *Backend {
	b := &Backend{
		name:          "mock",
		products:      make(map[string]*MockProduct),
		customers:     make(map[string]*MockCustomer),
		subscriptions: make(map[string]*MockSubscription),
		sessions:      make(map[string]*MockCheckout),
		webhookSecret: "mock_webhook_secret",
	}

	// Apply options to build configuration
	cfg := &payments.BackendConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Initialize products from configuration
	for _, product := range cfg.Products {
		b.productCount++
		productID := fmt.Sprintf("prod_mock_%d", b.productCount)
		priceID := fmt.Sprintf("price_mock_%d", b.productCount)

		b.products[product.Name] = &MockProduct{
			id:          productID,
			name:        product.Name,
			description: product.Description,
			priceID:     priceID,
			amount:      product.Price,
			currency:    product.Currency,
			interval:    product.Interval,
			metadata:    product.Metadata,
		}
	}

	return b
}

// MockProduct implements the payments.Product interface
type MockProduct struct {
	id          string
	name        string
	description string
	priceID     string
	amount      int64
	currency    string
	interval    string
	metadata    map[string]string
}

func (p *MockProduct) ID() string                 { return p.id }
func (p *MockProduct) Name() string               { return p.name }
func (p *MockProduct) Description() string        { return p.description }
func (p *MockProduct) PriceID() string            { return p.priceID }
func (p *MockProduct) Amount() int64              { return p.amount }
func (p *MockProduct) Currency() string           { return p.currency }
func (p *MockProduct) Interval() string           { return p.interval }
func (p *MockProduct) Metadata() map[string]string { return p.metadata }

// MockCustomer implements the payments.Customer interface
type MockCustomer struct {
	id    string
	email string
	name  string
}

func (c *MockCustomer) ID() string    { return c.id }
func (c *MockCustomer) Email() string { return c.email }
func (c *MockCustomer) Name() string  { return c.name }

// MockCheckout implements the payments.Checkout interface
type MockCheckout struct {
	id             string
	url            string
	status         string
	customerID     string
	subscriptionID string
	metadata       map[string]string
}

func (c *MockCheckout) ID() string                 { return c.id }
func (c *MockCheckout) URL() string                { return c.url }
func (c *MockCheckout) Status() string             { return c.status }
func (c *MockCheckout) CustomerID() string         { return c.customerID }
func (c *MockCheckout) SubscriptionID() string     { return c.subscriptionID }
func (c *MockCheckout) Metadata() map[string]string { return c.metadata }

// MockSubscription implements the payments.Subscription interface
type MockSubscription struct {
	id               string
	customerID       string
	productID        string
	priceID          string
	status           string
	currentPeriodEnd int64
}

func (s *MockSubscription) ID() string               { return s.id }
func (s *MockSubscription) CustomerID() string       { return s.customerID }
func (s *MockSubscription) ProductID() string        { return s.productID }
func (s *MockSubscription) PriceID() string          { return s.priceID }
func (s *MockSubscription) Status() string           { return s.status }
func (s *MockSubscription) CurrentPeriodEnd() int64  { return s.currentPeriodEnd }

// MockEvent implements the payments.Event interface
type MockEvent struct {
	id   string
	typ  string
	data map[string]interface{}
}

func (e *MockEvent) ID() string                   { return e.id }
func (e *MockEvent) Type() string                 { return e.typ }
func (e *MockEvent) Data() map[string]interface{} { return e.data }

// GetName returns the backend name
func (m *Backend) GetName() string {
	return m.name
}

// Product retrieves a product by ID (implements Backend interface)
func (m *Backend) Product(id string) (payments.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	for _, p := range m.products {
		if p.id == id {
			return p, nil
		}
	}

	return nil, fmt.Errorf("product not found: %s", id)
}

// ProductByName retrieves a product by name (implements Backend interface)
func (m *Backend) ProductByName(name string) (payments.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	if product, ok := m.products[name]; ok {
		return product, nil
	}

	return nil, fmt.Errorf("product %s not found", name)
}

// Products returns all configured products (implements Backend interface)
func (m *Backend) Products() ([]payments.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	products := make([]payments.Product, 0, len(m.products))
	for _, p := range m.products {
		products = append(products, p)
	}
	return products, nil
}

// CreateCheckout creates a mock checkout session (implements Backend interface)
func (m *Backend) CreateCheckout(productName string, opts ...payments.CheckoutOption) (payments.Checkout, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	// Look up product by name
	product, ok := m.products[productName]
	if !ok {
		return nil, fmt.Errorf("product %s not configured", productName)
	}

	// Build config from options
	cfg := payments.BuildCheckoutConfig(opts...)

	m.sessionCount++
	sessionID := fmt.Sprintf("cs_mock_%d", m.sessionCount)

	checkout := &MockCheckout{
		id:         sessionID,
		url:        fmt.Sprintf("https://checkout.mock/%s", sessionID),
		status:     "open",
		customerID: cfg.CustomerID,
		metadata:   cfg.Metadata,
	}

	// Create subscription if successful
	if cfg.CustomerID != "" {
		m.subCount++
		subID := fmt.Sprintf("sub_mock_%d", m.subCount)
		subscription := &MockSubscription{
			id:               subID,
			customerID:       cfg.CustomerID,
			productID:        product.id,
			priceID:          product.priceID,
			status:           "active",
			currentPeriodEnd: time.Now().AddDate(0, 1, 0).Unix(),
		}
		m.subscriptions[subID] = subscription
		checkout.subscriptionID = subID
	}

	m.sessions[sessionID] = checkout

	return checkout, nil
}

// Checkout retrieves a mock checkout session (implements Backend interface)
func (m *Backend) Checkout(id string) (payments.Checkout, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	return session, nil
}

// CreateCustomer creates a mock customer (implements Backend interface)
func (m *Backend) CreateCustomer(email, name string, opts ...payments.CustomerOption) (payments.Customer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	// Build config from options (unused for now)
	_ = payments.BuildCustomerConfig(opts...)

	m.customerCount++
	customerID := fmt.Sprintf("cus_mock_%d", m.customerCount)

	customer := &MockCustomer{
		id:    customerID,
		email: email,
		name:  name,
	}

	m.customers[customerID] = customer

	return customer, nil
}

// Customer retrieves a mock customer (implements Backend interface)
func (m *Backend) Customer(customerID string) (payments.Customer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	customer, exists := m.customers[customerID]
	if !exists {
		return nil, fmt.Errorf("customer not found: %s", customerID)
	}

	return customer, nil
}

// Customers retrieves a list of customers (implements Backend interface)
func (m *Backend) Customers(limit int) ([]payments.Customer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	customers := make([]payments.Customer, 0, len(m.customers))
	for _, c := range m.customers {
		customers = append(customers, c)
		if limit > 0 && len(customers) >= limit {
			break
		}
	}

	return customers, nil
}

// Subscription retrieves a mock subscription (implements Backend interface)
func (m *Backend) Subscription(subscriptionID string) (payments.Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	subscription, exists := m.subscriptions[subscriptionID]
	if !exists {
		return nil, fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	return subscription, nil
}

// Subscriptions retrieves a list of subscriptions (implements Backend interface)
func (m *Backend) Subscriptions(limit int) ([]payments.Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	subscriptions := make([]payments.Subscription, 0, len(m.subscriptions))
	for _, s := range m.subscriptions {
		subscriptions = append(subscriptions, s)
		if limit > 0 && len(subscriptions) >= limit {
			break
		}
	}

	return subscriptions, nil
}

// PauseSubscription pauses a mock subscription (implements Backend interface)
func (m *Backend) PauseSubscription(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("mock error: %s", m.failMessage)
	}

	subscription, exists := m.subscriptions[id]
	if !exists {
		return fmt.Errorf("subscription not found: %s", id)
	}

	// Mock pausing by changing status
	subscription.status = "paused"

	return nil
}

// ResumeSubscription resumes a mock subscription (implements Backend interface)
func (m *Backend) ResumeSubscription(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("mock error: %s", m.failMessage)
	}

	subscription, exists := m.subscriptions[id]
	if !exists {
		return fmt.Errorf("subscription not found: %s", id)
	}

	// Mock resuming by changing status
	subscription.status = "active"

	return nil
}

// CreatePortalSession creates a customer portal session (implements Backend interface)
func (m *Backend) CreatePortalSession(customerID, returnURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return "", fmt.Errorf("mock error: %s", m.failMessage)
	}

	if _, exists := m.customers[customerID]; !exists {
		return "", fmt.Errorf("customer not found: %s", customerID)
	}

	return fmt.Sprintf("https://portal.mock/session/%s", customerID), nil
}

// ConstructWebhookEvent verifies signature and constructs the webhook event (implements Backend interface)
func (m *Backend) ConstructWebhookEvent(payload []byte, signature string) (payments.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}

	// Simple mock verification
	if signature != m.webhookSecret {
		return nil, fmt.Errorf("invalid webhook signature")
	}

	// Return a mock event
	return &MockEvent{
		id:   "evt_mock_123",
		typ:  "checkout.session.completed",
		data: make(map[string]interface{}),
	}, nil
}

// Test helper methods

// FailNext causes the next operation to fail
func (m *Backend) FailNext(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = true
	m.failMessage = message
}

// Reset clears the failure state and all data
func (m *Backend) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = false
	m.failMessage = ""
	m.products = make(map[string]*MockProduct)
	m.customers = make(map[string]*MockCustomer)
	m.subscriptions = make(map[string]*MockSubscription)
	m.sessions = make(map[string]*MockCheckout)
	m.customerCount = 0
	m.sessionCount = 0
	m.subCount = 0
	m.productCount = 0
}

// SetWebhookSecret sets the webhook secret for testing
func (m *Backend) SetWebhookSecret(secret string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.webhookSecret = secret
}