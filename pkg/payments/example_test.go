package payments_test

import (
	"fmt"
	"log"

	"github.com/The-Skyscape/devtools/pkg/payments"
	"github.com/The-Skyscape/devtools/pkg/payments/stripe"
)

func ExampleService() {
	// Create payment service
	service := payments.NewService()

	// Add Stripe provider
	stripeClient := stripe.NewClient(
		"sk_test_YOUR_SECRET_KEY",
		"pk_test_YOUR_PUBLISHABLE_KEY",
		"whsec_YOUR_WEBHOOK_SECRET",
	)
	service.AddProvider("stripe", stripeClient)
	service.SetActiveProvider("stripe")

	// Create a checkout session
	session, err := service.CreateCheckoutSession(&payments.CheckoutParams{
		Mode:       "subscription",
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
		LineItems: []payments.LineItem{
			{
				PriceID:  "price_YOUR_PRICE_ID",
				Quantity: 1,
			},
		},
		CustomerEmail:   "customer@example.com",
		TrialPeriodDays: 14,
		Metadata: map[string]string{
			"user_id": "user_123",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Checkout URL: %s\n", session.URL)

	// Create a customer
	customer, err := service.CreateCustomer(&payments.CustomerParams{
		Email:       "john@example.com",
		Name:        "John Doe",
		Description: "Premium customer",
		Metadata: map[string]string{
			"source": "website",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Customer created: %s\n", customer.ID)

	// Create a subscription
	subscription, err := service.CreateSubscription(&payments.SubscriptionParams{
		CustomerID: customer.ID,
		PriceIDs:   []string{"price_YOUR_PRICE_ID"},
		Metadata: map[string]string{
			"plan": "premium",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Subscription created: %s (status: %s)\n", subscription.ID, subscription.Status)

	// Verify webhook signature
	payload := []byte(`{"id":"evt_123","type":"customer.subscription.created"}`)
	signature := "t=123456789,v1=abc123..."
	
	if service.VerifyWebhookSignature(payload, signature) {
		// Parse the webhook event
		event, err := service.ParseWebhookEvent(payload)
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Printf("Webhook event: %s (type: %s)\n", event.ID, event.Type)
		
		// Handle the event based on type
		switch event.Type {
		case payments.EventTypeSubscriptionCreated:
			fmt.Println("New subscription created!")
		case payments.EventTypePaymentIntentSucceeded:
			fmt.Println("Payment successful!")
		}
	}

	// Create a customer portal session
	portal, err := service.CreatePortalSession(customer.ID, "https://example.com/account")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Portal URL: %s\n", portal.URL)
}