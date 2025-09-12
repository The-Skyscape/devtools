package payments

import (
	"fmt"
	"sync"
	"time"
)

// MockProvider provides a mock payment provider for testing
type MockProvider struct {
	name string
	
	// Storage
	customers      map[string]*Customer
	subscriptions  map[string]*Subscription
	sessions       map[string]*CheckoutSession
	
	// Test control
	shouldFail     bool
	failMessage    string
	webhookSecret  string
	
	// Counters
	customerCount  int
	sessionCount   int
	subCount       int
	
	mu sync.RWMutex
}

// NewMockProvider creates a new mock payment provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		name:          "mock",
		customers:     make(map[string]*Customer),
		subscriptions: make(map[string]*Subscription),
		sessions:      make(map[string]*CheckoutSession),
		webhookSecret: "mock_webhook_secret",
	}
}

// GetName returns the provider name
func (m *MockProvider) GetName() string {
	return m.name
}

// CreateCheckoutSession creates a mock checkout session
func (m *MockProvider) CreateCheckoutSession(params *CheckoutParams) (*CheckoutSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	m.sessionCount++
	sessionID := fmt.Sprintf("cs_mock_%d", m.sessionCount)
	
	session := &CheckoutSession{
		ID:            sessionID,
		URL:           fmt.Sprintf("https://checkout.mock/%s", sessionID),
		Status:        "open",
		CustomerEmail: params.CustomerEmail,
		CustomerID:    params.CustomerID,
		Metadata:      params.Metadata,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}
	
	// Calculate amount
	var totalAmount int64
	for _, item := range params.LineItems {
		totalAmount += item.Amount * int64(item.Quantity)
	}
	session.AmountTotal = totalAmount
	
	m.sessions[sessionID] = session
	
	return session, nil
}

// GetCheckoutSession retrieves a mock checkout session
func (m *MockProvider) GetCheckoutSession(sessionID string) (*CheckoutSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	return session, nil
}

// CreateCustomer creates a mock customer
func (m *MockProvider) CreateCustomer(params *CustomerParams) (*Customer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	m.customerCount++
	customerID := fmt.Sprintf("cus_mock_%d", m.customerCount)
	
	customer := &Customer{
		ID:        customerID,
		Email:     params.Email,
		Name:      params.Name,
		Phone:     params.Phone,
		Metadata:  params.Metadata,
		Created:   time.Now(),
	}
	
	m.customers[customerID] = customer
	
	return customer, nil
}

// GetCustomer retrieves a mock customer
func (m *MockProvider) GetCustomer(customerID string) (*Customer, error) {
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

// UpdateCustomer updates a mock customer
func (m *MockProvider) UpdateCustomer(customerID string, params *CustomerParams) (*Customer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	customer, exists := m.customers[customerID]
	if !exists {
		return nil, fmt.Errorf("customer not found: %s", customerID)
	}
	
	// Update fields
	if params.Email != "" {
		customer.Email = params.Email
	}
	if params.Name != "" {
		customer.Name = params.Name
	}
	if params.Phone != "" {
		customer.Phone = params.Phone
	}
	if params.Metadata != nil {
		customer.Metadata = params.Metadata
	}
	
	return customer, nil
}

// DeleteCustomer deletes a mock customer
func (m *MockProvider) DeleteCustomer(customerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	if _, exists := m.customers[customerID]; !exists {
		return fmt.Errorf("customer not found: %s", customerID)
	}
	
	delete(m.customers, customerID)
	
	return nil
}

// CreateSubscription creates a mock subscription
func (m *MockProvider) CreateSubscription(params *SubscriptionParams) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	m.subCount++
	subID := fmt.Sprintf("sub_mock_%d", m.subCount)
	
	subscription := &Subscription{
		ID:                subID,
		CustomerID:        params.CustomerID,
		Status:           SubscriptionStatusActive,
		CurrentPeriodEnd: time.Now().AddDate(0, 1, 0), // 1 month from now
		Items:            []SubscriptionItem{}, // Empty for mock
		Metadata:         params.Metadata,
		Created:          time.Now(),
	}
	
	if params.TrialPeriodDays > 0 {
		subscription.Status = SubscriptionStatusTrialing
		trialEnd := time.Now().AddDate(0, 0, params.TrialPeriodDays)
		subscription.TrialEnd = &trialEnd
	}
	
	m.subscriptions[subID] = subscription
	
	return subscription, nil
}

// GetSubscription retrieves a mock subscription
func (m *MockProvider) GetSubscription(subscriptionID string) (*Subscription, error) {
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

// UpdateSubscription updates a mock subscription
func (m *MockProvider) UpdateSubscription(subscriptionID string, params *SubscriptionParams) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	subscription, exists := m.subscriptions[subscriptionID]
	if !exists {
		return nil, fmt.Errorf("subscription not found: %s", subscriptionID)
	}
	
	// Update fields
	// Note: Items not part of SubscriptionParams, skipping
	if params.Metadata != nil {
		subscription.Metadata = params.Metadata
	}
	
	return subscription, nil
}

// CancelSubscription cancels a mock subscription
func (m *MockProvider) CancelSubscription(subscriptionID string, immediately bool) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	subscription, exists := m.subscriptions[subscriptionID]
	if !exists {
		return nil, fmt.Errorf("subscription not found: %s", subscriptionID)
	}
	
	if immediately {
		subscription.Status = SubscriptionStatusCanceled
		canceledAt := time.Now()
		subscription.CancelledAt = &canceledAt
	} else {
		subscription.Status = SubscriptionStatusActive
		cancelAt := subscription.CurrentPeriodEnd
		subscription.CancelAt = &cancelAt
	}
	
	return subscription, nil
}

// AttachPaymentMethod attaches a payment method to a customer
func (m *MockProvider) AttachPaymentMethod(customerID, paymentMethodID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	if _, exists := m.customers[customerID]; !exists {
		return fmt.Errorf("customer not found: %s", customerID)
	}
	
	// Mock success
	return nil
}

// DetachPaymentMethod detaches a payment method
func (m *MockProvider) DetachPaymentMethod(paymentMethodID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	// Mock success
	return nil
}

// SetDefaultPaymentMethod sets the default payment method for a customer
func (m *MockProvider) SetDefaultPaymentMethod(customerID, paymentMethodID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	if _, exists := m.customers[customerID]; !exists {
		return fmt.Errorf("customer not found: %s", customerID)
	}
	
	// Mock success
	return nil
}

// VerifyWebhookSignature verifies a webhook signature
func (m *MockProvider) VerifyWebhookSignature(payload []byte, signature string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Simple mock verification
	return signature == m.webhookSecret
}

// ParseWebhookEvent parses a webhook event
func (m *MockProvider) ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	// Return a mock event
	return &WebhookEvent{
		ID:        "evt_mock_123",
		Type:      "checkout.session.completed",
		Data:      make(map[string]interface{}), // Use empty map for mock
		Created:   time.Now(),
	}, nil
}

// CreatePortalSession creates a customer portal session
func (m *MockProvider) CreatePortalSession(customerID, returnURL string) (*PortalSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldFail {
		return nil, fmt.Errorf("mock error: %s", m.failMessage)
	}
	
	if _, exists := m.customers[customerID]; !exists {
		return nil, fmt.Errorf("customer not found: %s", customerID)
	}
	
	return &PortalSession{
		ID:        fmt.Sprintf("ps_mock_%d", time.Now().Unix()),
		URL:       fmt.Sprintf("https://portal.mock/session/%s", customerID),
		ReturnURL: returnURL,
		Created:   time.Now(),
	}, nil
}

// Test helper methods

// FailNext causes the next operation to fail
func (m *MockProvider) FailNext(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = true
	m.failMessage = message
}

// Reset clears the failure state and all data
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = false
	m.failMessage = ""
	m.customers = make(map[string]*Customer)
	m.subscriptions = make(map[string]*Subscription)
	m.sessions = make(map[string]*CheckoutSession)
	m.customerCount = 0
	m.sessionCount = 0
	m.subCount = 0
}

// SetWebhookSecret sets the webhook secret for testing
func (m *MockProvider) SetWebhookSecret(secret string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.webhookSecret = secret
}

// GetCustomers returns all customers (for testing)
func (m *MockProvider) GetCustomers() []*Customer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	customers := make([]*Customer, 0, len(m.customers))
	for _, c := range m.customers {
		customers = append(customers, c)
	}
	return customers
}

// GetSubscriptions returns all subscriptions (for testing)
func (m *MockProvider) GetSubscriptions() []*Subscription {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	subs := make([]*Subscription, 0, len(m.subscriptions))
	for _, s := range m.subscriptions {
		subs = append(subs, s)
	}
	return subs
}

// CompleteCheckoutSession simulates completing a checkout session
func (m *MockProvider) CompleteCheckoutSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	session.Status = "complete"
	// Note: PaymentStatus not part of CheckoutSession struct
	
	// Create customer if needed
	if session.CustomerID == "" && session.CustomerEmail != "" {
		m.customerCount++
		customerID := fmt.Sprintf("cus_mock_%d", m.customerCount)
		customer := &Customer{
			ID:        customerID,
			Email:     session.CustomerEmail,
			Created:   time.Now(),
		}
		m.customers[customerID] = customer
		session.CustomerID = customerID
	}
	
	return nil
}