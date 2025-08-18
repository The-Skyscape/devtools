package memory

import (
	"fmt"
	"log"
	"sync"
)

// Memory implements the Secrets interface using in-memory storage
type Memory struct {
	secrets map[string]map[string]interface{}
	mu      sync.RWMutex
}

// New creates a new in-memory storage vault
func New() *Memory {
	return &Memory{
		secrets: make(map[string]map[string]interface{}),
	}
}

// Init initializes the memory vault
func (m *Memory) Init() error {
	log.Println("Memory: Initialized in-memory storage (WARNING: Secrets will not persist!)")
	return nil
}

// Close closes the memory vault
func (m *Memory) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Clear all secrets
	m.secrets = make(map[string]map[string]interface{})
	return nil
}

// IsAvailable returns true (memory vault is always available)
func (m *Memory) IsAvailable() bool {
	return true
}

// GetStorageMode returns "memory"
func (m *Memory) GetStorageMode() string {
	return "memory"
}

// GetStatus returns the memory vault status
func (m *Memory) GetStatus() interface{} {
	m.mu.RLock()
	count := len(m.secrets)
	m.mu.RUnlock()
	
	return map[string]interface{}{
		"mode":         "memory",
		"secret_count": count,
		"persistent":   false,
	}
}

// StoreSecret stores a secret in memory
func (m *Memory) StoreSecret(path string, data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Create a copy to avoid reference issues
	secretCopy := make(map[string]interface{})
	for k, v := range data {
		secretCopy[k] = v
	}
	
	m.secrets[path] = secretCopy
	return nil
}

// GetSecret retrieves a secret from memory
func (m *Memory) GetSecret(path string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	secret, exists := m.secrets[path]
	if !exists {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	
	// Return a copy to avoid external modification
	secretCopy := make(map[string]interface{})
	for k, v := range secret {
		secretCopy[k] = v
	}
	
	return secretCopy, nil
}

// DeleteSecret removes a secret from memory
func (m *Memory) DeleteSecret(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.secrets, path)
	return nil
}

// ListSecrets returns all secret paths
func (m *Memory) ListSecrets() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	paths := make([]string, 0, len(m.secrets))
	for path := range m.secrets {
		paths = append(paths, path)
	}
	
	return paths, nil
}