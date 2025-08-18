package security

// Config holds the configuration for the secrets collection
type Config struct {
	useVault          bool
	vaultConfig       *VaultConfig
	useFallback       bool
	fallbackDir       string
	useMemoryFallback bool
}

// Option configures the secrets collection
type Option func(*Config)

// WithVault configures Vault as the primary backend
func WithVault(opts ...VaultOption) Option {
	return func(c *Config) {
		c.useVault = true
		if c.vaultConfig == nil {
			c.vaultConfig = &VaultConfig{
				Port:          8200,
				ContainerName: "skyscape-vault",
				DevMode:       true,
				RootToken:     "skyscape-dev-token",
			}
		}
		for _, opt := range opts {
			opt(c.vaultConfig)
		}
	}
}

// WithoutVault disables Vault backend
func WithoutVault() Option {
	return func(c *Config) {
		c.useVault = false
	}
}

// WithoutFallback disables fallback storage
func WithoutFallback() Option {
	return func(c *Config) {
		c.useFallback = false
	}
}

// WithFallbackDir sets the directory for fallback file storage
func WithFallbackDir(dir string) Option {
	return func(c *Config) {
		c.fallbackDir = dir
	}
}

// WithMemoryFallback adds memory as a last-resort fallback
func WithMemoryFallback() Option {
	return func(c *Config) {
		c.useMemoryFallback = true
	}
}