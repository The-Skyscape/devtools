package security

// Example usage of the security package with the new Collection pattern
//
// This file demonstrates how to integrate the security package into your application
// for secure secret management with automatic fallback capabilities.

/*
Example 1: Basic Integration in main.go

```go
package main

import (
    "embed"
    "github.com/The-Skyscape/devtools/pkg/application"
    "github.com/The-Skyscape/devtools/pkg/security"
)

//go:embed views
var views embed.FS

func main() {
    // Create secrets collection with automatic fallback
    secrets := security.Manage(
        security.WithVault(
            security.WithContainerName("my-vault"),
            security.WithPort(8200),
            security.WithDevMode(true),
        ),
    )
    
    // Start application with SecretsController
    application.Serve(views,
        application.WithController(secrets.Controller()),
        application.WithController(controllers.Home()),
        // ... other controllers
    )
}
```

Example 2: Using SecretsController in Other Controllers

```go
package controllers

type AdminController struct {
    application.BaseController
    // ... other fields
}

func (c *AdminController) Setup(app *application.App) {
    c.BaseController.Setup(app)
    
    // Get the secrets controller
    secrets := app.Use("secrets").(*security.Controller)
    
    // Check if Vault is available
    if !secrets.IsVaultAvailable() {
        log.Println("Warning: Running with fallback storage")
    }
}

// Template method to check configuration
func (c *AdminController) IsStripeConfigured() bool {
    secrets := c.Use("secrets").(*security.Controller)
    return secrets.IsStripeConfigured()
}

// Store API keys
func (c *AdminController) saveStripeKeys(w http.ResponseWriter, r *http.Request) {
    secrets := c.Use("secrets").(*security.Controller)
    
    secretKey := r.FormValue("secret_key")
    publishKey := r.FormValue("publish_key")
    webhookSecret := r.FormValue("webhook_secret")
    
    if err := secrets.StoreStripeKeys(secretKey, publishKey, webhookSecret); err != nil {
        c.Render(w, r, "error.html", err)
        return
    }
    
    c.Redirect(w, r, "/admin/integrations")
}
```

Example 3: Template Usage

```html
<!-- Check storage mode in templates -->
{{if secrets.IsFallbackMode}}
    <div class="alert alert-warning">
        <span class="icon">⚠️</span>
        Running in fallback mode ({{secrets.GetStorageMode}})
        - Secrets may not persist across restarts
    </div>
{{end}}

<!-- Check if Vault is available -->
{{if secrets.IsVaultAvailable}}
    <a href="{{secrets.GetVaultURL}}" target="_blank" class="btn btn-sm">
        Open Vault UI
    </a>
{{else}}
    <span class="text-muted">Vault not available</span>
{{end}}

<!-- Check integration status -->
{{if secrets.IsStripeConfigured}}
    <span class="badge badge-success">Configured</span>
{{else}}
    <span class="badge badge-warning">Not Configured</span>
{{end}}
```

Example 4: Protected Routes with Middleware

```go
func (c *BillingController) Setup(app *application.App) {
    c.BaseController.Setup(app)
    
    secrets := app.Use("secrets").(*security.Controller)
    auth := app.Use("auth").(*authentication.Controller)
    
    // Route that requires secrets to be available
    http.Handle("POST /billing/checkout", 
        app.ProtectFunc(c.createCheckout, secrets.RequireSecrets, auth.Required))
    
    // Route that requires real Vault (no fallback)
    http.Handle("POST /admin/vault/restart", 
        app.ProtectFunc(c.restartVault, secrets.RequireVault, auth.AdminOnly))
}
```

Example 5: Handling Different Storage Modes

```go
func (c *AdminController) GetSecretStatus() map[string]interface{} {
    secrets := c.Use("secrets").(*security.Controller)
    
    status := map[string]interface{}{
        "mode": secrets.GetStorageMode(),
        "vault_available": secrets.IsVaultAvailable(),
        "fallback_mode": secrets.IsFallbackMode(),
    }
    
    // Add warnings based on storage mode
    switch secrets.GetStorageMode() {
    case "memory":
        status["warning"] = "Secrets stored in memory - will be lost on restart"
    case "file":
        status["warning"] = "Secrets stored in encrypted file - limited security"
    case "vault":
        status["info"] = "Secrets stored securely in Vault"
    default:
        status["error"] = "Unknown storage mode"
    }
    
    return status
}
```

Example 6: Programmatic Fallback Configuration

```go
func initSecrets() *security.Controller {
    vault := security.NewVaultService()
    
    // Create a hybrid fallback with multiple backends
    fallback := security.NewHybridBackend(
        security.NewEnvBackend("MYAPP"),      // Try environment variables first
        security.NewFileBackend(),            // Then encrypted files
        security.NewMemoryBackend(),          // Finally in-memory
    )
    
    // Set the fallback on the vault service
    vault.SetFallback(fallback)
    
    _, controller := security.NewController(vault)
    return controller
}
```

Example 7: Testing with Mock Storage

```go
func TestBillingController(t *testing.T) {
    // Create a memory backend for testing
    mockStorage := security.NewMemoryBackend()
    mockStorage.Init()
    
    // Store test credentials
    mockStorage.StoreSecret("integrations/stripe", map[string]interface{}{
        "secret_key": "sk_test_123",
        "publishable_key": "pk_test_456",
        "webhook_secret": "whsec_789",
    })
    
    // Create vault service with mock storage
    vault := security.NewVaultService()
    vault.SetFallback(mockStorage)
    
    // Create controller with mocked secrets
    _, secrets := security.NewController(vault)
    
    // Test your controller
    if !secrets.IsStripeConfigured() {
        t.Error("Expected Stripe to be configured")
    }
}
```

Example 8: Migration from Direct Vault Usage

```go
// Before: Direct vault usage
func (c *Controller) saveKeys() {
    services.Vault.StoreStripeKeys(key1, key2, key3)
}

// After: Using SecretsController
func (c *Controller) saveKeys() {
    secrets := c.Use("secrets").(*security.Controller)
    secrets.StoreStripeKeys(key1, key2, key3)
}

// Before: Checking configuration
func (c *Controller) IsConfigured() bool {
    return services.Vault.IsStripeConfigured()
}

// After: Using SecretsController
func (c *Controller) IsConfigured() bool {
    secrets := c.Use("secrets").(*security.Controller)
    return secrets.IsStripeConfigured()
}
```

Fallback Behavior Summary:
1. SecretsController tries to start Vault container in Setup()
2. If Docker is unavailable, falls back to file-based storage
3. If file storage fails, falls back to in-memory storage
4. Templates can check storage mode and show appropriate warnings
5. All secret operations work seamlessly regardless of backend
6. Middleware can enforce specific storage requirements

Best Practices:
- Always check IsFallbackMode() in production deployments
- Show warnings to users when not using Vault
- Use RequireVault() middleware for sensitive operations
- Test with different storage backends
- Log storage mode on startup
- Consider environment variables for read-only secrets
- Use hybrid backends for maximum flexibility
*/