package security

import (
	"fmt"
	"log"
	"time"

	"github.com/The-Skyscape/devtools/pkg/containers"
	"github.com/The-Skyscape/devtools/pkg/database"
)

// VaultConfig holds configuration for the vault service
type VaultConfig struct {
	Port          int
	ContainerName string
	DataDir       string
	DevMode       bool
	RootToken     string
	Network       string
	ExposePort    bool // Whether to bind port to host
}

// VaultService manages Hashicorp Vault for secure secret storage
type VaultService struct {
	config   *VaultConfig
	service  *containers.Service
	client   *VaultClient
	fallback Secrets
}

// VaultStatus represents the current status of the vault service
type VaultStatus struct {
	Running   bool
	Port      int
	DevMode   bool
	URL       string
	Health    string
	RootToken string
}

// NewVaultService creates a new vault service with default configuration
func NewVaultService(opts ...VaultOption) *VaultService {
	config := &VaultConfig{
		Port:          8200,
		ContainerName: "skyscape-vault",
		DataDir:       fmt.Sprintf("%s/vault", database.DataDir()),
		DevMode:       true,
		RootToken:     "skyscape-dev-token",
		Network:       "skyscape-internal",
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Determine vault URL based on network configuration
	vaultURL := fmt.Sprintf("http://localhost:%d", config.Port)
	if config.Network != "host" && config.Network != "" {
		// Use container name for internal network communication
		vaultURL = fmt.Sprintf("http://%s:%d", config.ContainerName, config.Port)
	}

	return &VaultService{
		config: config,
		client: NewVaultClient(vaultURL, config.RootToken),
	}
}

// GetService returns the container service definition for Vault
func (v *VaultService) GetService() *containers.Service {
	if v.service != nil {
		return v.service
	}

	// Create the service configuration
	v.service = &containers.Service{
		Name:          v.config.ContainerName,
		Image:         "hashicorp/vault:latest",
		Network:       v.config.Network,
		RestartPolicy: "always",
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID":  v.config.RootToken,
			"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
			"VAULT_ADDR":               "http://0.0.0.0:8200",
			"VAULT_API_ADDR":           "http://0.0.0.0:8200",
		},
	}

	// Only expose port if explicitly configured
	if v.config.ExposePort {
		v.service.Ports = map[string]string{
			fmt.Sprintf("%d", v.config.Port): "8200",
		}
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

	return v.service
}

// Init initializes the vault service and starts it if not already running
func (v *VaultService) Init() error {
	host := &containers.LocalHost{}
	return v.InitWithHost(host)
}

// InitWithHost initializes the vault service with a specific host
func (v *VaultService) InitWithHost(host containers.Host) error {

	// Check if service already exists and is running
	existing, err := containers.GetService(host, v.config.ContainerName)
	if err == nil && existing != nil && existing.IsRunning() {
		log.Println("Vault service already running")
		v.service = existing
		return nil
	}

	log.Println("Initializing Vault service...")

	// Create internal network if specified and doesn't exist
	if v.config.Network != "" && v.config.Network != "host" {
		log.Printf("Creating Docker network %s if not exists...", v.config.Network)
		// Try to create network, ignore error if already exists
		host.Exec("docker", "network", "create", v.config.Network)
	}

	// Get the service definition
	service := v.GetService()
	service.Host = host

	// Launch the container
	if err := containers.Launch(host, service); err != nil {
		return fmt.Errorf("failed to launch vault container: %w", err)
	}

	// Wait for vault to be ready
	if err := service.WaitForReady(30, func() error {
		// Simple health check - vault will respond on its API port
		return host.Exec("curl", "-f", fmt.Sprintf("http://localhost:%d/v1/sys/health", v.config.Port))
	}); err != nil {
		log.Printf("Warning: Vault may not be fully ready: %v", err)
	}

	log.Printf("Vault service started successfully on port %d", v.config.Port)
	if v.config.DevMode {
		log.Printf("Vault running in dev mode with root token: %s", v.config.RootToken)
		log.Printf("Access Vault UI at: http://localhost:%d", v.config.Port)
	}

	v.service = service
	return nil
}

// Start starts the vault service
func (v *VaultService) Start(host containers.Host) error {

	if v.service == nil {
		v.service = v.GetService()
		v.service.Host = host
	}

	if v.IsRunning() {
		return nil
	}

	log.Printf("Starting Vault service on port %d", v.config.Port)

	// Launch the container
	if err := containers.Launch(host, v.service); err != nil {
		return fmt.Errorf("failed to launch vault container: %w", err)
	}

	// Wait for vault to be ready with health check
	if err := v.service.WaitForReady(30*time.Second, func() error {
		return host.Exec("curl", "-f", fmt.Sprintf("http://localhost:%d/v1/sys/health", v.config.Port))
	}); err != nil {
		log.Printf("Warning: Vault may not be fully ready: %v", err)
	}

	return nil
}

// Stop stops the vault service
func (v *VaultService) Stop() error {

	if v.service == nil {
		return nil
	}

	log.Println("Stopping Vault service")
	return v.service.Stop()
}

// IsRunning checks if the vault service is running
func (v *VaultService) IsRunning() bool {
	if v.service == nil {
		return false
	}
	return v.service.IsRunning()
}

// IsAvailable returns true if Vault is running and accessible
func (v *VaultService) IsAvailable() bool {
	return v.IsRunning()
}

// GetStorageMode returns "vault"
func (v *VaultService) GetStorageMode() string {
	return "vault"
}

// Close closes the Vault service
func (v *VaultService) Close() error {
	return v.Stop()
}

// ListSecrets returns all secret paths (not implemented for Vault)
func (v *VaultService) ListSecrets() ([]string, error) {
	// Vault doesn't easily support listing all secrets
	// Return empty list or use fallback
	if v.fallback != nil {
		return v.fallback.ListSecrets()
	}
	return []string{}, nil
}

// Restart restarts the vault service
func (v *VaultService) Restart(host containers.Host) error {
	if err := v.Stop(); err != nil {
		return err
	}
	return v.Start(host)
}

// GetStatus returns the current status as any
func (v *VaultService) GetStatus() any {
	return v.GetVaultStatus()
}

// GetVaultStatus returns the current status of the vault service
func (v *VaultService) GetVaultStatus() VaultStatus {

	status := VaultStatus{
		Running:   v.IsRunning(),
		Port:      v.config.Port,
		DevMode:   v.config.DevMode,
		URL:       fmt.Sprintf("http://localhost:%d", v.config.Port),
		RootToken: v.config.RootToken,
	}

	if status.Running {
		status.Health = "healthy"
	} else {
		status.Health = "stopped"
	}

	return status
}

// GetClient returns the Vault API client
func (v *VaultService) GetClient() *VaultClient {
	return v.client
}

// SetFallback sets the fallback storage backend
func (v *VaultService) SetFallback(fallback Secrets) {
	v.fallback = fallback
}

// HasFallback returns true if a fallback backend is configured
func (v *VaultService) HasFallback() bool {
	return v.fallback != nil
}

// StoreSecret stores a secret at the given path
func (v *VaultService) StoreSecret(path string, data map[string]any) error {
	if v.client != nil {
		return v.client.StoreSecret(path, data)
	}
	if v.fallback != nil {
		return v.fallback.StoreSecret(path, data)
	}
	return fmt.Errorf("no storage backend available")
}

// GetSecret retrieves a secret from the given path
func (v *VaultService) GetSecret(path string) (map[string]any, error) {
	if v.client != nil {
		return v.client.GetSecret(path)
	}
	if v.fallback != nil {
		return v.fallback.GetSecret(path)
	}
	return nil, fmt.Errorf("no storage backend available")
}

// DeleteSecret removes a secret at the given path
func (v *VaultService) DeleteSecret(path string) error {
	if v.client != nil {
		return v.client.DeleteSecret(path)
	}
	if v.fallback != nil {
		return v.fallback.DeleteSecret(path)
	}
	return fmt.Errorf("no storage backend available")
}
