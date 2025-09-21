package payments

// BackendConfig holds configuration for backend initialization
type BackendConfig struct {
	Products []ProductConfig
}

// ProductConfig holds product configuration
type ProductConfig struct {
	Name        string
	Description string
	Price       int64  // Amount in cents
	Currency    string
	Interval    string // "month", "year", etc
	Metadata    map[string]string
}

// BackendOption configures a backend during initialization
type BackendOption func(*BackendConfig)

// WithProduct adds a product to be configured during backend initialization
func WithProduct(name, description string, opts ...ProductOption) BackendOption {
	return func(c *BackendConfig) {
		pc := ProductConfig{
			Name:        name,
			Description: description,
			Currency:    "usd",   // Default currency
			Interval:    "month", // Default interval
			Metadata:    make(map[string]string),
		}

		// Apply product options
		for _, opt := range opts {
			opt(&pc)
		}

		c.Products = append(c.Products, pc)
	}
}

// ProductOption configures a product
type ProductOption func(*ProductConfig)

// WithMonthlyPrice sets the monthly price for a product
func WithMonthlyPrice(dollars float64) ProductOption {
	return func(p *ProductConfig) {
		p.Price = int64(dollars * 100) // Convert to cents
		p.Interval = "month"
	}
}

// WithAnnualPrice sets the annual price for a product
func WithAnnualPrice(dollars float64) ProductOption {
	return func(p *ProductConfig) {
		p.Price = int64(dollars * 100) // Convert to cents
		p.Interval = "year"
	}
}

// WithPrice sets a custom price and interval
func WithPrice(cents int64, interval string) ProductOption {
	return func(p *ProductConfig) {
		p.Price = cents
		p.Interval = interval
	}
}

// WithCurrency sets the currency for a product
func WithCurrency(currency string) ProductOption {
	return func(p *ProductConfig) {
		p.Currency = currency
	}
}

// WithMetadata adds metadata to a product
func WithMetadata(key, value string) ProductOption {
	return func(p *ProductConfig) {
		if p.Metadata == nil {
			p.Metadata = make(map[string]string)
		}
		p.Metadata[key] = value
	}
}

// CustomerOption configures customer creation
type CustomerOption func(*customerConfig)

type customerConfig struct {
	Metadata map[string]string
	Phone    string
}

// WithCustomerMetadata adds metadata to a customer
func WithCustomerMetadata(key, value string) CustomerOption {
	return func(c *customerConfig) {
		if c.Metadata == nil {
			c.Metadata = make(map[string]string)
		}
		c.Metadata[key] = value
	}
}

// WithCustomerPhone sets the customer's phone number
func WithCustomerPhone(phone string) CustomerOption {
	return func(c *customerConfig) {
		c.Phone = phone
	}
}

// CheckoutOption configures checkout session creation
type CheckoutOption func(*checkoutConfig)

type checkoutConfig struct {
	SuccessURL      string
	CancelURL       string
	CustomerID      string
	CustomerEmail   string
	TrialDays       int
	Quantity        int
	Metadata        map[string]string
	AllowPromoCodes bool
}

// WithSuccessURL sets the success redirect URL
func WithSuccessURL(url string) CheckoutOption {
	return func(c *checkoutConfig) {
		c.SuccessURL = url
	}
}

// WithCancelURL sets the cancel redirect URL
func WithCancelURL(url string) CheckoutOption {
	return func(c *checkoutConfig) {
		c.CancelURL = url
	}
}

// WithCustomerID attaches checkout to existing customer
func WithCustomerID(id string) CheckoutOption {
	return func(c *checkoutConfig) {
		c.CustomerID = id
	}
}

// WithCustomerEmail sets email for new customer
func WithCustomerEmail(email string) CheckoutOption {
	return func(c *checkoutConfig) {
		c.CustomerEmail = email
	}
}

// WithTrialDays sets trial period for subscription
func WithTrialDays(days int) CheckoutOption {
	return func(c *checkoutConfig) {
		c.TrialDays = days
	}
}

// WithQuantity sets the quantity for checkout
func WithQuantity(qty int) CheckoutOption {
	return func(c *checkoutConfig) {
		c.Quantity = qty
	}
}

// WithCheckoutMetadata adds metadata to checkout session
func WithCheckoutMetadata(key, value string) CheckoutOption {
	return func(c *checkoutConfig) {
		if c.Metadata == nil {
			c.Metadata = make(map[string]string)
		}
		c.Metadata[key] = value
	}
}

// WithAllowPromoCodes enables promo codes in checkout
func WithAllowPromoCodes() CheckoutOption {
	return func(c *checkoutConfig) {
		c.AllowPromoCodes = true
	}
}

// BuildCheckoutConfig builds config from options
func BuildCheckoutConfig(opts ...CheckoutOption) *checkoutConfig {
	cfg := &checkoutConfig{
		Quantity: 1, // Default quantity
		Metadata: make(map[string]string),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// BuildCustomerConfig builds config from options
func BuildCustomerConfig(opts ...CustomerOption) *customerConfig {
	cfg := &customerConfig{
		Metadata: make(map[string]string),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}