package security

// VaultOption is a configuration option for VaultService
type VaultOption func(*VaultConfig)

// WithPort sets the port for Vault
func WithPort(port int) VaultOption {
	return func(c *VaultConfig) {
		c.Port = port
	}
}

// WithContainerName sets the container name
func WithContainerName(name string) VaultOption {
	return func(c *VaultConfig) {
		c.ContainerName = name
	}
}

// WithDataDir sets the data directory
func WithDataDir(dir string) VaultOption {
	return func(c *VaultConfig) {
		c.DataDir = dir
	}
}

// WithDevMode enables or disables dev mode
func WithDevMode(devMode bool) VaultOption {
	return func(c *VaultConfig) {
		c.DevMode = devMode
	}
}

// WithRootToken sets the root token for dev mode
func WithRootToken(token string) VaultOption {
	return func(c *VaultConfig) {
		c.RootToken = token
	}
}

// WithNetwork sets the Docker network
func WithNetwork(network string) VaultOption {
	return func(c *VaultConfig) {
		c.Network = network
	}
}

// WithPortBinding controls whether to expose vault port on host
func WithPortBinding(enabled bool) VaultOption {
	return func(c *VaultConfig) {
		c.ExposePort = enabled
	}
}