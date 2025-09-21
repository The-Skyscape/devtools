// +build ignore

package payments_test

import (
	"testing"

	"github.com/The-Skyscape/devtools/pkg/payments"
	"github.com/The-Skyscape/devtools/pkg/payments/backends/mock"
	"github.com/The-Skyscape/devtools/pkg/payments/backends/stripe"
)

func TestWithMockBackend(t *testing.T) {
	// In tests, use the mock backend
	var backend payments.Backend = mock.New()

	// Create a customer
	customer, err := backend.CreateCustomer(&payments.CustomerParams{
		Email: "test@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatal(err)
	}

	if customer.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", customer.Email)
	}
}

func TestWithStripeBackend(t *testing.T) {
	// In production, use the stripe backend
	var backend payments.Backend = stripe.NewClient(
		"sk_test_fake",
		"pk_test_fake",
		"whsec_fake",
	)

	// Backend is ready to use
	if backend.GetName() != "stripe" {
		t.Errorf("expected backend name stripe, got %s", backend.GetName())
	}
}

func Example_backendUsage() {
	// You can swap backends easily for testing
	var backend payments.Backend

	if testing.Testing() {
		// Use mock for tests
		backend = mock.New()
	} else {
		// Use real backend for production
		backend = stripe.NewClient("sk_live_xxx", "pk_live_xxx", "whsec_xxx")
	}

	// Use the backend the same way regardless of implementation
	_, _ = backend.CreateCheckoutSession(&payments.CheckoutParams{
		Mode:       "subscription",
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	})
}