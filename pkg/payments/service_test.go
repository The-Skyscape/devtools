package payments_test

import (
	"testing"

	"github.com/The-Skyscape/devtools/pkg/payments"
	"github.com/The-Skyscape/devtools/pkg/payments/stripe"
)

func TestServiceCreation(t *testing.T) {
	// Test service creation
	service := payments.NewService()
	
	if service == nil {
		t.Fatal("Service is nil")
	}
	
	// Initially not configured
	if service.IsConfigured() {
		t.Fatal("Service should not be configured initially")
	}
	
	// Add Stripe provider
	stripeClient := stripe.NewClient(
		"sk_test_fake",
		"pk_test_fake", 
		"whsec_fake",
	)
	
	err := service.AddProvider("stripe", stripeClient)
	if err != nil {
		t.Fatalf("Failed to add provider: %v", err)
	}
	
	// Now should be configured
	if !service.IsConfigured() {
		t.Fatal("Service should be configured after adding provider")
	}
	
	// Check provider name
	if service.ProviderName() != "stripe" {
		t.Fatalf("Expected provider 'stripe', got %s", service.ProviderName())
	}
	
	// Check providers list
	providers := service.ListProviders()
	if len(providers) != 1 {
		t.Fatalf("Expected 1 provider, got %d", len(providers))
	}
	if providers[0] != "stripe" {
		t.Fatalf("Expected provider 'stripe' in list, got %s", providers[0])
	}
}

func TestMultipleProviders(t *testing.T) {
	service := payments.NewService()
	
	// Add multiple providers
	stripe1 := stripe.NewClient("sk1", "pk1", "wh1")
	stripe2 := stripe.NewClient("sk2", "pk2", "wh2")
	
	service.AddProvider("stripe1", stripe1)
	service.AddProvider("stripe2", stripe2)
	
	// First added should be active (provider name is "stripe" not the key)
	if service.ProviderName() != "stripe" {
		t.Fatalf("Expected active provider 'stripe', got %s", service.ProviderName())
	}
	
	// Switch active provider
	err := service.SetActiveProvider("stripe2")
	if err != nil {
		t.Fatalf("Failed to set active provider: %v", err)
	}
	
	// Provider name is still "stripe" (that's what the Stripe client returns)
	if service.ProviderName() != "stripe" {
		t.Fatalf("Expected active provider 'stripe', got %s", service.ProviderName())
	}
	
	// Test getting specific provider
	provider, err := service.GetProvider("stripe1")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}
	if provider.GetName() != "stripe" {
		t.Fatalf("Expected provider name 'stripe', got %s", provider.GetName())
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test FormatAmount
	formatted := payments.FormatAmount(1999, "usd")
	if formatted != "$19.99" {
		t.Fatalf("Expected '$19.99', got %s", formatted)
	}
	
	formatted = payments.FormatAmount(1000, "eur")
	if formatted != "€10.00" {
		t.Fatalf("Expected '€10.00', got %s", formatted)
	}
	
	formatted = payments.FormatAmount(1000, "jpy")
	if formatted != "¥1000" {
		t.Fatalf("Expected '¥1000', got %s", formatted)
	}
	
	// Test IsSubscriptionActive
	if !payments.IsSubscriptionActive(payments.SubscriptionStatusActive) {
		t.Fatal("Active status should be considered active")
	}
	if !payments.IsSubscriptionActive(payments.SubscriptionStatusTrialing) {
		t.Fatal("Trialing status should be considered active")
	}
	if payments.IsSubscriptionActive(payments.SubscriptionStatusCanceled) {
		t.Fatal("Canceled status should not be considered active")
	}
	
	// Test IsPaymentRequired
	if !payments.IsPaymentRequired(payments.SubscriptionStatusPastDue) {
		t.Fatal("PastDue status should require payment")
	}
	if !payments.IsPaymentRequired(payments.SubscriptionStatusUnpaid) {
		t.Fatal("Unpaid status should require payment")
	}
	if payments.IsPaymentRequired(payments.SubscriptionStatusActive) {
		t.Fatal("Active status should not require payment")
	}
	
	// Test CalculateTrialEnd
	trialEnd := payments.CalculateTrialEnd(14)
	if trialEnd == nil {
		t.Fatal("Trial end should not be nil for positive days")
	}
	
	trialEnd = payments.CalculateTrialEnd(0)
	if trialEnd != nil {
		t.Fatal("Trial end should be nil for 0 days")
	}
}