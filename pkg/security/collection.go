package security

import (
	"log"

	"github.com/The-Skyscape/devtools/pkg/security/vaults/fallback"
	"github.com/The-Skyscape/devtools/pkg/security/vaults/file"
	"github.com/The-Skyscape/devtools/pkg/security/vaults/memory"
	"github.com/The-Skyscape/devtools/pkg/security/vaults/vault"
)

// Collection manages secrets with automatic fallback using vaults
type Collection struct {
	vault Secrets
}

// Manage creates a new secrets collection with the given options
func Manage(opts ...Option) *Collection {
	config := &Config{
		useVault:     true,
		useFallback:  true,
		vaultConfig:  nil,
	}
	
	// Apply options
	for _, opt := range opts {
		opt(config)
	}
	
	// Build vault chain based on configuration
	var vaults []fallback.Secrets
	
	// Add Vault if configured
	if config.useVault && config.vaultConfig != nil {
		v := vault.New(
			vault.WithPort(config.vaultConfig.Port),
			vault.WithContainerName(config.vaultConfig.ContainerName),
			vault.WithDataDir(config.vaultConfig.DataDir),
			vault.WithDevMode(config.vaultConfig.DevMode),
			vault.WithRootToken(config.vaultConfig.RootToken),
			vault.WithNetwork(config.vaultConfig.Network),
		)
		vaults = append(vaults, v)
	}
	
	// Add fallback vaults if enabled
	if config.useFallback {
		// Add file vault as primary fallback
		if config.fallbackDir != "" {
			vaults = append(vaults, file.New(config.fallbackDir))
		} else {
			vaults = append(vaults, file.New())
		}
		
		// Add memory vault as last resort if specified
		if config.useMemoryFallback {
			vaults = append(vaults, memory.New())
		}
	}
	
	// If no vaults configured, use file vault by default
	if len(vaults) == 0 {
		vaults = append(vaults, file.New())
	}
	
	// Create fallback vault with all configured vaults
	fb := fallback.New(vaults...)
	
	// Initialize the fallback vault
	if err := fb.Init(); err != nil {
		log.Printf("Collection: Failed to initialize vaults: %v", err)
		// Continue anyway, operations will fail but won't crash
	}
	
	return &Collection{
		vault: fb,
	}
}

// StoreSecret stores a secret at the given path
func (c *Collection) StoreSecret(path string, data map[string]interface{}) error {
	return c.vault.StoreSecret(path, data)
}

// GetSecret retrieves a secret from the given path
func (c *Collection) GetSecret(path string) (map[string]interface{}, error) {
	return c.vault.GetSecret(path)
}

// DeleteSecret removes a secret at the given path
func (c *Collection) DeleteSecret(path string) error {
	return c.vault.DeleteSecret(path)
}

// ListSecrets returns all secret paths
func (c *Collection) ListSecrets() ([]string, error) {
	return c.vault.ListSecrets()
}

// IsVaultAvailable returns true if Vault is the active backend
func (c *Collection) IsVaultAvailable() bool {
	return c.vault.GetStorageMode() == "vault"
}

// IsFallbackMode returns true if using fallback storage (not vault)
func (c *Collection) IsFallbackMode() bool {
	mode := c.vault.GetStorageMode()
	return mode != "vault" && mode != "none"
}

// GetStorageMode returns the current storage mode
func (c *Collection) GetStorageMode() string {
	return c.vault.GetStorageMode()
}

// GetStatus returns the current status
func (c *Collection) GetStatus() interface{} {
	return c.vault.GetStatus()
}

// Restart attempts to restart/reconnect to vaults
func (c *Collection) Restart() error {
	// If using fallback, try to reconnect to higher priority vaults
	if fb, ok := c.vault.(*fallback.Fallback); ok {
		return fb.TryReconnect()
	}
	
	// Otherwise just reinitialize
	return c.vault.Init()
}

// Close closes all vaults
func (c *Collection) Close() error {
	return c.vault.Close()
}