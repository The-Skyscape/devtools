package fallback

import (
	"fmt"
	"log"
)

// Secrets interface that all vaults must implement
type Secrets interface {
	// Core operations
	StoreSecret(path string, data map[string]interface{}) error
	GetSecret(path string) (map[string]interface{}, error)
	DeleteSecret(path string) error
	ListSecrets() ([]string, error)
	
	// Status operations
	IsAvailable() bool
	GetStorageMode() string
	GetStatus() interface{}
	
	// Lifecycle
	Init() error
	Close() error
}

// Fallback implements automatic fallback between multiple vaults
type Fallback struct {
	vaults  []Secrets
	current Secrets
	name    string
}

// New creates a new fallback vault with multiple backends
func New(vaults ...Secrets) *Fallback {
	return &Fallback{
		vaults: vaults,
		name:   "fallback",
	}
}

// Init initializes the fallback vault and determines which backend to use
func (f *Fallback) Init() error {
	if len(f.vaults) == 0 {
		return fmt.Errorf("no vaults configured for fallback")
	}
	
	// Try to initialize each vault in order
	for i, vault := range f.vaults {
		err := vault.Init()
		if err != nil {
			log.Printf("Fallback: Vault %d (%s) initialization failed: %v", 
				i, vault.GetStorageMode(), err)
			continue
		}
		
		if vault.IsAvailable() {
			f.current = vault
			log.Printf("Fallback: Using %s as primary storage", vault.GetStorageMode())
			return nil
		}
		
		log.Printf("Fallback: Vault %d (%s) not available", i, vault.GetStorageMode())
	}
	
	// If no vault is available, use the last one as fallback
	// (it should be the most reliable, like memory or file)
	if len(f.vaults) > 0 {
		lastVault := f.vaults[len(f.vaults)-1]
		if err := lastVault.Init(); err == nil {
			f.current = lastVault
			log.Printf("Fallback: Using %s as last resort", lastVault.GetStorageMode())
			return nil
		}
	}
	
	return fmt.Errorf("no vault backends available")
}

// Close closes all vaults
func (f *Fallback) Close() error {
	var lastErr error
	for _, vault := range f.vaults {
		if err := vault.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// IsAvailable returns true if any vault is available
func (f *Fallback) IsAvailable() bool {
	return f.current != nil && f.current.IsAvailable()
}

// GetStorageMode returns the current vault's storage mode
func (f *Fallback) GetStorageMode() string {
	if f.current != nil {
		return f.current.GetStorageMode()
	}
	return "none"
}

// GetStatus returns status information about all vaults
func (f *Fallback) GetStatus() interface{} {
	status := map[string]interface{}{
		"current": "none",
		"vaults":  []map[string]interface{}{},
	}
	
	if f.current != nil {
		status["current"] = f.current.GetStorageMode()
	}
	
	vaultStatuses := []map[string]interface{}{}
	for _, vault := range f.vaults {
		vaultStatus := map[string]interface{}{
			"mode":      vault.GetStorageMode(),
			"available": vault.IsAvailable(),
			"status":    vault.GetStatus(),
		}
		vaultStatuses = append(vaultStatuses, vaultStatus)
	}
	status["vaults"] = vaultStatuses
	
	return status
}

// StoreSecret stores a secret in the current vault
func (f *Fallback) StoreSecret(path string, data map[string]interface{}) error {
	if f.current == nil {
		return fmt.Errorf("no vault available")
	}
	
	err := f.current.StoreSecret(path, data)
	if err != nil {
		// Try to failover to next available vault
		if newVault := f.findNextAvailable(); newVault != nil {
			f.current = newVault
			log.Printf("Fallback: Switching to %s due to error: %v", 
				newVault.GetStorageMode(), err)
			return newVault.StoreSecret(path, data)
		}
	}
	
	return err
}

// GetSecret retrieves a secret from the current vault
func (f *Fallback) GetSecret(path string) (map[string]interface{}, error) {
	if f.current == nil {
		return nil, fmt.Errorf("no vault available")
	}
	
	secret, err := f.current.GetSecret(path)
	if err != nil {
		// Try other vaults if current one fails
		for _, vault := range f.vaults {
			if vault != f.current && vault.IsAvailable() {
				if s, err := vault.GetSecret(path); err == nil {
					// Found the secret in another vault
					// Consider switching to this vault
					return s, nil
				}
			}
		}
	}
	
	return secret, err
}

// DeleteSecret removes a secret from all vaults
func (f *Fallback) DeleteSecret(path string) error {
	if f.current == nil {
		return fmt.Errorf("no vault available")
	}
	
	// Delete from current vault
	err := f.current.DeleteSecret(path)
	
	// Also try to delete from all other vaults to ensure consistency
	for _, vault := range f.vaults {
		if vault != f.current && vault.IsAvailable() {
			vault.DeleteSecret(path) // Ignore errors for other vaults
		}
	}
	
	return err
}

// ListSecrets returns all unique secret paths from all vaults
func (f *Fallback) ListSecrets() ([]string, error) {
	if f.current == nil {
		return nil, fmt.Errorf("no vault available")
	}
	
	// Get secrets from current vault
	paths, err := f.current.ListSecrets()
	if err != nil {
		paths = []string{}
	}
	
	// Build a set of unique paths from all vaults
	pathSet := make(map[string]bool)
	for _, path := range paths {
		pathSet[path] = true
	}
	
	// Try to get paths from other vaults too
	for _, vault := range f.vaults {
		if vault != f.current && vault.IsAvailable() {
			if otherPaths, err := vault.ListSecrets(); err == nil {
				for _, path := range otherPaths {
					pathSet[path] = true
				}
			}
		}
	}
	
	// Convert set back to slice
	uniquePaths := make([]string, 0, len(pathSet))
	for path := range pathSet {
		uniquePaths = append(uniquePaths, path)
	}
	
	return uniquePaths, nil
}

// findNextAvailable finds the next available vault
func (f *Fallback) findNextAvailable() Secrets {
	for _, vault := range f.vaults {
		if vault != f.current && vault.IsAvailable() {
			return vault
		}
	}
	return nil
}

// TryReconnect attempts to reconnect to a higher-priority vault
func (f *Fallback) TryReconnect() error {
	// Try to reconnect to vaults in priority order
	for i, vault := range f.vaults {
		// Skip if this is already the current vault
		if vault == f.current {
			continue
		}
		
		// Try to initialize and check availability
		if err := vault.Init(); err == nil && vault.IsAvailable() {
			oldMode := "none"
			if f.current != nil {
				oldMode = f.current.GetStorageMode()
			}
			
			f.current = vault
			log.Printf("Fallback: Reconnected to %s (was using %s)", 
				vault.GetStorageMode(), oldMode)
			
			// Stop after finding the first available vault
			// (assuming vaults are in priority order)
			if i == 0 {
				log.Printf("Fallback: Restored primary vault")
			}
			return nil
		}
	}
	
	return fmt.Errorf("no better vault available for reconnection")
}