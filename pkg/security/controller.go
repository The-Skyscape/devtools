package security

import (
	"fmt"
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// Controller provides a web controller for managing secrets
type Controller struct {
	application.BaseController
	*Collection
}

// Controller returns a new secrets controller for the collection
func (c *Collection) Controller(opts ...ControllerOption) (string, *Controller) {
	controller := &Controller{
		Collection: c,
	}
	
	// Apply controller options
	for _, opt := range opts {
		opt(controller)
	}
	
	return "secrets", controller
}

// ControllerOption configures a secrets controller
type ControllerOption func(*Controller)

// Setup initializes the secrets controller
func (c *Controller) Setup(app *application.App) {
	c.BaseController.Setup(app)
	// Collection is already initialized, nothing else needed
}

// Handle returns the controller instance for the current request
func (c *Controller) Handle(r *http.Request) application.Controller {
	c.Request = r
	return c
}

// ========== Public Methods for Templates ==========

// IsVaultAvailable returns true if Vault is running and accessible
func (c *Controller) IsVaultAvailable() bool {
	return c.Collection.IsVaultAvailable()
}

// IsFallbackMode returns true if using fallback storage
func (c *Controller) IsFallbackMode() bool {
	return c.Collection.IsFallbackMode()
}

// GetStorageMode returns the current storage mode
func (c *Controller) GetStorageMode() string {
	return c.Collection.GetStorageMode()
}

// GetVaultStatus returns the current Vault status
func (c *Controller) GetVaultStatus() VaultStatus {
	if status, ok := c.Collection.GetStatus().(VaultStatus); ok {
		return status
	}
	
	// Return fallback status
	return VaultStatus{
		Running:   false,
		Health:    c.GetStorageMode(),
		URL:       "N/A",
		DevMode:   false,
		Port:      0,
		RootToken: "",
	}
}

// GetVaultURL returns the Vault UI URL if available
func (c *Controller) GetVaultURL() string {
	status := c.GetVaultStatus()
	if status.Running {
		return status.URL
	}
	return ""
}

// IsStripeConfigured checks if Stripe is configured
func (c *Controller) IsStripeConfigured() bool {
	secret, err := c.GetSecret("integrations/stripe")
	if err != nil {
		return false
	}
	_, hasKey := secret["secret_key"]
	return hasKey
}

// IsDigitalOceanConfigured checks if DigitalOcean is configured
func (c *Controller) IsDigitalOceanConfigured() bool {
	secret, err := c.GetSecret("integrations/digitalocean")
	if err != nil {
		return false
	}
	_, hasKey := secret["api_key"]
	return hasKey
}

// GetLastError returns the last error encountered
func (c *Controller) GetLastError() string {
	// This functionality was removed in the new structure
	// Return empty string for compatibility
	return ""
}

// ========== Secret Management Methods ==========

// StoreSecret stores a secret at the given path
func (c *Controller) StoreSecret(path string, data map[string]interface{}) error {
	return c.Collection.StoreSecret(path, data)
}

// GetSecret retrieves a secret from the given path
func (c *Controller) GetSecret(path string) (map[string]interface{}, error) {
	return c.Collection.GetSecret(path)
}

// DeleteSecret removes a secret at the given path
func (c *Controller) DeleteSecret(path string) error {
	return c.Collection.DeleteSecret(path)
}

// ========== Integration-Specific Methods ==========

// StoreStripeKeys stores Stripe API keys
func (c *Controller) StoreStripeKeys(secretKey, publishableKey, webhookSecret string) error {
	data := map[string]interface{}{
		"secret_key":      secretKey,
		"publishable_key": publishableKey,
		"webhook_secret":  webhookSecret,
	}
	return c.StoreSecret("integrations/stripe", data)
}

// GetStripeKeys retrieves Stripe API keys
func (c *Controller) GetStripeKeys() (secretKey, publishableKey, webhookSecret string, err error) {
	secret, err := c.GetSecret("integrations/stripe")
	if err != nil {
		return "", "", "", err
	}
	
	if sk, ok := secret["secret_key"].(string); ok {
		secretKey = sk
	}
	if pk, ok := secret["publishable_key"].(string); ok {
		publishableKey = pk
	}
	if ws, ok := secret["webhook_secret"].(string); ok {
		webhookSecret = ws
	}
	
	if secretKey == "" || publishableKey == "" {
		return "", "", "", fmt.Errorf("incomplete Stripe configuration")
	}
	
	return secretKey, publishableKey, webhookSecret, nil
}

// StoreDigitalOceanKey stores the DigitalOcean API key
func (c *Controller) StoreDigitalOceanKey(apiKey string) error {
	data := map[string]interface{}{
		"api_key": apiKey,
	}
	return c.StoreSecret("integrations/digitalocean", data)
}

// GetDigitalOceanKey retrieves the DigitalOcean API key
func (c *Controller) GetDigitalOceanKey() (string, error) {
	secret, err := c.GetSecret("integrations/digitalocean")
	if err != nil {
		return "", err
	}
	
	if apiKey, ok := secret["api_key"].(string); ok {
		return apiKey, nil
	}
	
	return "", fmt.Errorf("DigitalOcean API key not found")
}

// RestartVault attempts to restart the Vault backend
func (c *Controller) RestartVault() error {
	return c.Collection.Restart()
}

// ========== Middleware Functions ==========

// RequireSecrets is middleware that ensures secrets are available
func (c *Controller) RequireSecrets(app *application.App, w http.ResponseWriter, r *http.Request) bool {
	if c.Collection != nil && c.GetStorageMode() != "none" {
		return true
	}
	
	// No storage available
	app.Render(w, r, "error-message.html", fmt.Errorf("Secret storage is not available"))
	return false
}

// RequireVault is middleware that requires real Vault (no fallback)
func (c *Controller) RequireVault(app *application.App, w http.ResponseWriter, r *http.Request) bool {
	if c.IsVaultAvailable() {
		return true
	}
	
	app.Render(w, r, "error-message.html", fmt.Errorf("Vault is required but not available"))
	return false
}