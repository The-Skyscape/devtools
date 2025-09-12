package security

// Secrets defines the interface for secret storage backends
type Secrets interface {
	// Core operations
	StoreSecret(path string, data map[string]any) error
	GetSecret(path string) (map[string]any, error)
	DeleteSecret(path string) error
	ListSecrets() ([]string, error)

	// Status operations
	IsAvailable() bool
	GetStorageMode() string
	GetStatus() any

	// Lifecycle
	Init() error
	Close() error
}
