package payments

import (
	"fmt"
	"sync"
	"time"
)

// Service manages payment operations through various providers
type Service struct {
	providers map[string]Provider
	active    Provider
	mu        sync.RWMutex
}

// NewService creates a new payment service
func NewService() *Service {
	return &Service{
		providers: make(map[string]Provider),
	}
}

// AddProvider adds a payment provider to the service
func (s *Service) AddProvider(name string, provider Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	s.providers[name] = provider

	// Set as active if it's the first provider
	if s.active == nil {
		s.active = provider
	}

	return nil
}

// SetActiveProvider sets the active payment provider
func (s *Service) SetActiveProvider(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider, ok := s.providers[name]
	if !ok {
		return fmt.Errorf("provider %s not found", name)
	}

	s.active = provider
	return nil
}

// GetProvider returns a specific provider by name
func (s *Service) GetProvider(name string) (Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, ok := s.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ActiveProvider returns the currently active provider
func (s *Service) ActiveProvider() Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// Checkout operations (use active provider)

// CreateCheckoutSession creates a checkout session using the active provider
func (s *Service) CreateCheckoutSession(params *CheckoutParams) (*CheckoutSession, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.CreateCheckoutSession(params)
}

// GetCheckoutSession retrieves a checkout session using the active provider
func (s *Service) GetCheckoutSession(sessionID string) (*CheckoutSession, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.GetCheckoutSession(sessionID)
}

// Customer operations

// CreateCustomer creates a customer using the active provider
func (s *Service) CreateCustomer(params *CustomerParams) (*Customer, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.CreateCustomer(params)
}

// GetCustomer retrieves a customer using the active provider
func (s *Service) GetCustomer(customerID string) (*Customer, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.GetCustomer(customerID)
}

// UpdateCustomer updates a customer using the active provider
func (s *Service) UpdateCustomer(customerID string, params *CustomerParams) (*Customer, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.UpdateCustomer(customerID, params)
}

// DeleteCustomer deletes a customer using the active provider
func (s *Service) DeleteCustomer(customerID string) error {
	provider := s.ActiveProvider()
	if provider == nil {
		return fmt.Errorf("no active payment provider")
	}
	return provider.DeleteCustomer(customerID)
}

// Subscription operations

// CreateSubscription creates a subscription using the active provider
func (s *Service) CreateSubscription(params *SubscriptionParams) (*Subscription, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.CreateSubscription(params)
}

// GetSubscription retrieves a subscription using the active provider
func (s *Service) GetSubscription(subscriptionID string) (*Subscription, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.GetSubscription(subscriptionID)
}

// UpdateSubscription updates a subscription using the active provider
func (s *Service) UpdateSubscription(subscriptionID string, params *SubscriptionParams) (*Subscription, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.UpdateSubscription(subscriptionID, params)
}

// CancelSubscription cancels a subscription using the active provider
func (s *Service) CancelSubscription(subscriptionID string, immediately bool) (*Subscription, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.CancelSubscription(subscriptionID, immediately)
}

// PauseSubscription pauses payment collection for a subscription using the active provider
func (s *Service) PauseSubscription(subscriptionID string, resumesAt *time.Time) (*Subscription, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.PauseSubscription(subscriptionID, resumesAt)
}

// ResumeSubscription resumes payment collection for a subscription using the active provider
func (s *Service) ResumeSubscription(subscriptionID string) (*Subscription, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.ResumeSubscription(subscriptionID)
}

// Payment method operations

// AttachPaymentMethod attaches a payment method to a customer
func (s *Service) AttachPaymentMethod(customerID, paymentMethodID string) error {
	provider := s.ActiveProvider()
	if provider == nil {
		return fmt.Errorf("no active payment provider")
	}
	return provider.AttachPaymentMethod(customerID, paymentMethodID)
}

// DetachPaymentMethod detaches a payment method
func (s *Service) DetachPaymentMethod(paymentMethodID string) error {
	provider := s.ActiveProvider()
	if provider == nil {
		return fmt.Errorf("no active payment provider")
	}
	return provider.DetachPaymentMethod(paymentMethodID)
}

// SetDefaultPaymentMethod sets the default payment method for a customer
func (s *Service) SetDefaultPaymentMethod(customerID, paymentMethodID string) error {
	provider := s.ActiveProvider()
	if provider == nil {
		return fmt.Errorf("no active payment provider")
	}
	return provider.SetDefaultPaymentMethod(customerID, paymentMethodID)
}

// Webhook operations

// VerifyWebhookSignature verifies a webhook signature using the active provider
func (s *Service) VerifyWebhookSignature(payload []byte, signature string) bool {
	provider := s.ActiveProvider()
	if provider == nil {
		return false
	}
	return provider.VerifyWebhookSignature(payload, signature)
}

// ParseWebhookEvent parses a webhook event using the active provider
func (s *Service) ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.ParseWebhookEvent(payload)
}

// Portal operations

// CreatePortalSession creates a customer portal session
func (s *Service) CreatePortalSession(customerID, returnURL string) (*PortalSession, error) {
	provider := s.ActiveProvider()
	if provider == nil {
		return nil, fmt.Errorf("no active payment provider")
	}
	return provider.CreatePortalSession(customerID, returnURL)
}

// Helper methods

// IsConfigured returns true if at least one provider is configured
func (s *Service) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.providers) > 0 && s.active != nil
}

// ProviderName returns the name of the active provider
func (s *Service) ProviderName() string {
	provider := s.ActiveProvider()
	if provider == nil {
		return ""
	}
	return provider.GetName()
}

// ListProviders returns a list of configured provider names
func (s *Service) ListProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}
