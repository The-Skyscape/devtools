package security

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

var (
	//go:embed resources/vault.hcl
	vaultConfig string
	
	//go:embed resources/init.sh
	initScript string
	
	//go:embed resources/app-policy.hcl
	appPolicy string
)

// WriteVaultConfig writes the Vault configuration to the specified directory
func WriteVaultConfig(configDir string) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	configPath := filepath.Join(configDir, "vault.hcl")
	if err := os.WriteFile(configPath, []byte(vaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write vault config: %w", err)
	}
	
	return nil
}

// GetInitScript returns the initialization script for Vault
func GetInitScript() string {
	return initScript
}

// GetAppPolicy returns the default application policy for Vault
func GetAppPolicy() string {
	return appPolicy
}