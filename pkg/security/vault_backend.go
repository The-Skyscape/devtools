package security

import (
	"fmt"
	"log"
	"time"

	"github.com/The-Skyscape/devtools/pkg/containers"
)

// VaultBackend wraps VaultService to implement the Secrets interface
type VaultBackend struct {
	*VaultService
}

// NewVaultBackend creates a new Vault backend
func NewVaultBackend(config *VaultConfig) *VaultBackend {
	if config == nil {
		config = &VaultConfig{
			Port:          8200,
			ContainerName: "skyscape-vault",
			DevMode:       true,
			RootToken:     "skyscape-dev-token",
		}
	}
	
	vault := NewVaultService(
		WithContainerName(config.ContainerName),
		WithPort(config.Port),
		WithDevMode(config.DevMode),
		WithRootToken(config.RootToken),
		WithDataDir(config.DataDir),
		WithNetwork(config.Network),
	)
	
	return &VaultBackend{
		VaultService: vault,
	}
}

// Init initializes the Vault backend
func (v *VaultBackend) Init() error {
	log.Println("VaultBackend: Initializing HashiCorp Vault...")
	
	// Try to launch container
	host := &containers.LocalHost{}
	
	// Check if already running
	existing, err := containers.GetService(host, v.config.ContainerName)
	if err == nil && existing != nil && existing.IsRunning() {
		log.Println("VaultBackend: Vault container already running")
		v.service = existing
	} else {
		// Launch new container using InitWithHost method
		if err := v.VaultService.InitWithHost(host); err != nil {
			return fmt.Errorf("failed to initialize vault: %w", err)
		}
		
		// Wait for vault to be ready
		time.Sleep(3 * time.Second)
	}
	
	log.Printf("VaultBackend: Vault initialized on port %d", v.config.Port)
	if v.config.DevMode {
		log.Printf("VaultBackend: Dev mode with root token: %s", v.config.RootToken)
	}
	
	return nil
}

// GetStatus returns the current status as interface{}
func (v *VaultBackend) GetStatus() interface{} {
	return v.VaultService.GetStatus()
}