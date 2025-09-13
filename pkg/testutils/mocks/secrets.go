package mocks

import (
	"fmt"
	"sync"
)

// MockSecrets implements the security.Secrets interface for testing
type MockSecrets struct {
	mu       sync.RWMutex
	secrets  map[string]map[string]any
	isAvail  bool
	mode     string
	status   string
	Error    error
	Closed   bool
}

// NewMockSecrets creates a new mock secrets store
func NewMockSecrets() *MockSecrets {
	return &MockSecrets{
		secrets: make(map[string]map[string]any),
		isAvail: true,
		mode:    "mock",
		status:  "healthy",
	}
}

// StoreSecret stores a secret at the given path
func (m *MockSecrets) StoreSecret(path string, data map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.Error != nil {
		return m.Error
	}
	
	if m.Closed {
		return fmt.Errorf("secrets store is closed")
	}
	
	// Make a copy of the data to avoid mutations
	secretData := make(map[string]any)
	for k, v := range data {
		secretData[k] = v
	}
	
	m.secrets[path] = secretData
	return nil
}

// GetSecret retrieves a secret from the given path
func (m *MockSecrets) GetSecret(path string) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.Error != nil {
		return nil, m.Error
	}
	
	if m.Closed {
		return nil, fmt.Errorf("secrets store is closed")
	}
	
	data, exists := m.secrets[path]
	if !exists {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}
	
	// Return a copy to avoid mutations
	result := make(map[string]any)
	for k, v := range data {
		result[k] = v
	}
	
	return result, nil
}

// DeleteSecret removes a secret at the given path
func (m *MockSecrets) DeleteSecret(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.Error != nil {
		return m.Error
	}
	
	if m.Closed {
		return fmt.Errorf("secrets store is closed")
	}
	
	delete(m.secrets, path)
	return nil
}

// ListSecrets returns all secret paths
func (m *MockSecrets) ListSecrets() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.Error != nil {
		return nil, m.Error
	}
	
	if m.Closed {
		return nil, fmt.Errorf("secrets store is closed")
	}
	
	paths := make([]string, 0, len(m.secrets))
	for path := range m.secrets {
		paths = append(paths, path)
	}
	
	return paths, nil
}

// IsAvailable returns whether the secrets store is available
func (m *MockSecrets) IsAvailable() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isAvail
}

// GetStorageMode returns the storage mode
func (m *MockSecrets) GetStorageMode() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// GetStatus returns the status of the secrets store
func (m *MockSecrets) GetStatus() any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// Init initializes the secrets store
func (m *MockSecrets) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.Error != nil {
		return m.Error
	}
	
	m.Closed = false
	return nil
}

// Close closes the secrets store
func (m *MockSecrets) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.Error != nil {
		return m.Error
	}
	
	m.Closed = true
	return nil
}

// SetAvailable sets whether the store is available
func (m *MockSecrets) SetAvailable(available bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isAvail = available
}

// SetMode sets the storage mode
func (m *MockSecrets) SetMode(mode string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mode = mode
}

// SetStatus sets the status
func (m *MockSecrets) SetStatus(status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = status
}

// SetError sets an error to be returned by operations
func (m *MockSecrets) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = err
}

// Reset clears all secrets and resets state
func (m *MockSecrets) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.secrets = make(map[string]map[string]any)
	m.isAvail = true
	m.mode = "mock"
	m.status = "healthy"
	m.Error = nil
	m.Closed = false
}

// AssertSecretExists verifies a secret exists at the given path
func (m *MockSecrets) AssertSecretExists(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if _, exists := m.secrets[path]; !exists {
		return fmt.Errorf("expected secret at path %s, but it does not exist", path)
	}
	return nil
}

// AssertSecretValue verifies a secret has the expected value
func (m *MockSecrets) AssertSecretValue(path, key string, expected any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	secret, exists := m.secrets[path]
	if !exists {
		return fmt.Errorf("secret not found at path: %s", path)
	}
	
	value, exists := secret[key]
	if !exists {
		return fmt.Errorf("key %s not found in secret at path %s", key, path)
	}
	
	if value != expected {
		return fmt.Errorf("expected value %v for key %s, got %v", expected, key, value)
	}
	
	return nil
}