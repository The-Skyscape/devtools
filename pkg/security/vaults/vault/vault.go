package vault

import (
	"fmt"
	"log"
	"time"

	"github.com/The-Skyscape/devtools/pkg/containers"
	"github.com/The-Skyscape/devtools/pkg/database"
)

// Vault implements the Secrets interface using HashiCorp Vault
type Vault struct {
	config  *Config
	service *containers.Service
	client  *Client
}

// Config holds configuration for the vault
type Config struct {
	Port          int
	ContainerName string
	DataDir       string
	DevMode       bool
	RootToken     string
	Network       string
}

// New creates a new Vault backend
func New(opts ...Option) *Vault {
	config := &Config{
		Port:          8200,
		ContainerName: "skyscape-vault",
		DataDir:       database.DataDir() + "/vault",
		DevMode:       true,
		RootToken:     "skyscape-dev-token",
		Network:       "host",
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	return &Vault{
		config: config,
	}
}

// Option configures the vault
type Option func(*Config)

// WithPort sets the port
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithContainerName sets the container name
func WithContainerName(name string) Option {
	return func(c *Config) {
		c.ContainerName = name
	}
}

// WithDataDir sets the data directory
func WithDataDir(dir string) Option {
	return func(c *Config) {
		c.DataDir = dir
	}
}

// WithDevMode sets dev mode
func WithDevMode(dev bool) Option {
	return func(c *Config) {
		c.DevMode = dev
	}
}

// WithRootToken sets the root token
func WithRootToken(token string) Option {
	return func(c *Config) {
		c.RootToken = token
	}
}

// WithNetwork sets the network
func WithNetwork(network string) Option {
	return func(c *Config) {
		c.Network = network
	}
}

// Init initializes the Vault backend
func (v *Vault) Init() error {
	log.Println("Vault: Initializing HashiCorp Vault...")

	// Create container service
	v.service = &containers.Service{
		Name:          v.config.ContainerName,
		Image:         "hashicorp/vault:latest",
		Network:       v.config.Network,
		RestartPolicy: "always",
		Ports: map[int]int{
			8200: v.config.Port,
		},
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID":  v.config.RootToken,
			"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
			"VAULT_ADDR":               "http://0.0.0.0:8200",
			"VAULT_API_ADDR":           "http://0.0.0.0:8200",
		},
	}

	// Configure for dev or production mode
	if v.config.DevMode {
		v.service.Command = "vault server -dev -dev-listen-address=0.0.0.0:8200"
	} else {
		// For production mode, mount data directory and config
		v.service.Mounts = map[string]string{
			v.config.DataDir:             "/vault/data",
			v.config.DataDir + "/config": "/vault/config",
		}
		v.service.Command = "vault server -config=/vault/config"
		v.service.Env["VAULT_DISABLE_MLOCK"] = "true"
	}

	// Try to launch container
	host := &containers.LocalHost{}

	// Check if already running
	existing, err := containers.GetService(host, v.config.ContainerName)
	if err == nil && existing != nil && existing.IsRunning() {
		log.Println("Vault: Container already running")
		v.service = existing
	} else {
		// Launch new container
		v.service.Host = host
		if err := containers.Launch(host, v.service); err != nil {
			return fmt.Errorf("failed to launch vault container: %w", err)
		}

		// Wait for vault to be ready
		time.Sleep(3 * time.Second)
	}

	// Initialize client
	v.client = NewClient(
		fmt.Sprintf("http://localhost:%d", v.config.Port),
		v.config.RootToken,
	)

	log.Printf("Vault: Initialized on port %d", v.config.Port)
	if v.config.DevMode {
		log.Printf("Vault: Dev mode with root token: %s", v.config.RootToken)
	}

	return nil
}

// Close closes the Vault backend
func (v *Vault) Close() error {
	if v.service != nil {
		return v.service.Stop()
	}
	return nil
}

// IsAvailable returns true if Vault is running and accessible
func (v *Vault) IsAvailable() bool {
	if v.service == nil {
		return false
	}
	return v.service.IsRunning()
}

// GetStorageMode returns "vault"
func (v *Vault) GetStorageMode() string {
	return "vault"
}

// GetStatus returns the current Vault status
func (v *Vault) GetStatus() any {
	return map[string]any{
		"running":   v.IsAvailable(),
		"port":      v.config.Port,
		"dev_mode":  v.config.DevMode,
		"url":       fmt.Sprintf("http://localhost:%d", v.config.Port),
		"container": v.config.ContainerName,
	}
}

// StoreSecret stores a secret at the given path
func (v *Vault) StoreSecret(path string, data map[string]any) error {
	if v.client == nil {
		return fmt.Errorf("vault client not initialized")
	}
	return v.client.StoreSecret(path, data)
}

// GetSecret retrieves a secret from the given path
func (v *Vault) GetSecret(path string) (map[string]any, error) {
	if v.client == nil {
		return nil, fmt.Errorf("vault client not initialized")
	}
	return v.client.GetSecret(path)
}

// DeleteSecret removes a secret at the given path
func (v *Vault) DeleteSecret(path string) error {
	if v.client == nil {
		return fmt.Errorf("vault client not initialized")
	}
	return v.client.DeleteSecret(path)
}

// ListSecrets returns all secret paths (not easily supported by Vault)
func (v *Vault) ListSecrets() ([]string, error) {
	// Vault doesn't easily support listing all secrets
	// Return empty list
	return []string{}, nil
}
